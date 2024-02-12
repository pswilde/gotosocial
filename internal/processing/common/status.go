// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package common

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// GetTargetStatusBy fetches the target status with db load
// function, given the authorized (or, nil) requester's
// account. This returns an approprate gtserror.WithCode
// accounting for not found and visibility to requester.
//
// window can be used to force refresh of the target if it's
// deemed to be stale. Falls back to default window if nil.
func (p *Processor) GetTargetStatusBy(
	ctx context.Context,
	requester *gtsmodel.Account,
	getTargetFromDB func() (*gtsmodel.Status, error),
	window *dereferencing.FreshnessWindow,
) (
	status *gtsmodel.Status,
	visible bool,
	errWithCode gtserror.WithCode,
) {
	// Fetch the target status from db.
	target, err := getTargetFromDB()
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, false, gtserror.NewErrorInternalError(err)
	}

	if target == nil {
		// DB loader could not find status in database.
		const text = "target status not found"
		return nil, false, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	// Check whether target status is visible to requesting account.
	visible, err = p.filter.StatusVisible(ctx, requester, target)
	if err != nil {
		return nil, false, gtserror.NewErrorInternalError(err)
	}

	if requester != nil && visible {
		// We only bother refreshing if this status
		// is visible to requester, AND there *is*
		// a requester (i.e. request is authorized)
		// to prevent a possible DOS vector.

		if window != nil {
			// Window is explicitly set, so likely
			// tighter than the default window.
			// Do refresh synchronously.
			_, _, err := p.federator.RefreshStatus(ctx,
				requester.Username,
				target,
				nil,
				window,
			)
			if err != nil {
				log.Errorf(ctx, "error refreshing status: %v", err)
			}
		} else {
			// Only refresh async *if* out-of-date.
			p.federator.RefreshStatusAsync(ctx,
				requester.Username,
				target,
				nil,
				nil,
			)
		}
	}

	return target, visible, nil
}

// GetVisibleTargetStatus calls GetTargetStatusBy(),
// but converts a non-visible result to not-found error.
//
// window can be used to force refresh of the target if it's
// deemed to be stale. Falls back to default window if nil.
func (p *Processor) GetVisibleTargetStatusBy(
	ctx context.Context,
	requester *gtsmodel.Account,
	getTargetFromDB func() (*gtsmodel.Status, error),
	window *dereferencing.FreshnessWindow,
) (
	status *gtsmodel.Status,
	errWithCode gtserror.WithCode,
) {
	// Fetch the target status by ID from the database.
	target, visible, errWithCode := p.GetTargetStatusBy(ctx,
		requester,
		getTargetFromDB,
		window,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if !visible {
		// Target should not be seen by requester.
		const text = "target status not found"
		return nil, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	return target, nil
}

// GetVisibleTargetStatus calls GetVisibleTargetStatusBy(),
// passing in a database function that fetches by status ID.
//
// window can be used to force refresh of the target if it's
// deemed to be stale. Falls back to default window if nil.
func (p *Processor) GetVisibleTargetStatus(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetID string,
	window *dereferencing.FreshnessWindow,
) (
	status *gtsmodel.Status,
	errWithCode gtserror.WithCode,
) {
	return p.GetVisibleTargetStatusBy(ctx, requester, func() (*gtsmodel.Status, error) {
		return p.state.DB.GetStatusByID(ctx, targetID)
	}, window)
}

// UnwrapIfBoost "unwraps" the given status if
// it's a boost wrapper, by returning the boosted
// status it targets (pending visibility checks).
//
// Just returns the input status if it's not a boost.
func (p *Processor) UnwrapIfBoost(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (*gtsmodel.Status, gtserror.WithCode) {
	if status.BoostOfID == "" {
		return status, nil
	}
	return p.GetVisibleTargetStatus(ctx,
		requester,
		status.BoostOfID,
		nil,
	)
}

// GetAPIStatus fetches the appropriate API status model for target.
func (p *Processor) GetAPIStatus(
	ctx context.Context,
	requester *gtsmodel.Account,
	target *gtsmodel.Status,
) (
	apiStatus *apimodel.Status,
	errWithCode gtserror.WithCode,
) {
	apiStatus, err := p.converter.StatusToAPIStatus(ctx, target, requester)
	if err != nil {
		err = gtserror.Newf("error converting status: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return apiStatus, nil
}

// GetVisibleAPIStatuses converts an array of gtsmodel.Status (inputted by next function) into
// API model statuses, checking first for visibility. Please note that all errors will be
// logged at ERROR level, but will not be returned. Callers are likely to run into show-stopping
// errors in the lead-up to this function, whereas calling this should not be a show-stopper.
func (p *Processor) GetVisibleAPIStatuses(
	ctx context.Context,
	requester *gtsmodel.Account,
	next func(int) *gtsmodel.Status,
	length int,
) []*apimodel.Status {
	return p.getVisibleAPIStatuses(ctx, 3, requester, next, length)
}

// GetVisibleAPIStatusesPaged is functionally equivalent to GetVisibleAPIStatuses(),
// except the statuses are returned as a converted slice of statuses as interface{}.
func (p *Processor) GetVisibleAPIStatusesPaged(
	ctx context.Context,
	requester *gtsmodel.Account,
	next func(int) *gtsmodel.Status,
	length int,
) []interface{} {
	statuses := p.getVisibleAPIStatuses(ctx, 3, requester, next, length)
	if len(statuses) == 0 {
		return nil
	}
	items := make([]interface{}, len(statuses))
	for i, status := range statuses {
		items[i] = status
	}
	return items
}

func (p *Processor) getVisibleAPIStatuses(
	ctx context.Context,
	calldepth int, // used to skip wrapping func above these's names
	requester *gtsmodel.Account,
	next func(int) *gtsmodel.Status,
	length int,
) []*apimodel.Status {
	// Start new log entry with
	// the above calling func's name.
	l := log.
		WithContext(ctx).
		WithField("caller", log.Caller(calldepth+1))

	// Preallocate slice according to expected length.
	statuses := make([]*apimodel.Status, 0, length)

	for i := 0; i < length; i++ {
		// Get next status.
		status := next(i)
		if status == nil {
			continue
		}

		// Check whether this status is visible to requesting account.
		visible, err := p.filter.StatusVisible(ctx, requester, status)
		if err != nil {
			l.Errorf("error checking status visibility: %v", err)
			continue
		}

		if !visible {
			// Not visible to requester.
			continue
		}

		// Convert the status to an API model representation.
		apiStatus, err := p.converter.StatusToAPIStatus(ctx, status, requester)
		if err != nil {
			l.Errorf("error converting status: %v", err)
			continue
		}

		// Append API model to return slice.
		statuses = append(statuses, apiStatus)
	}

	return statuses
}

// InvalidateTimelinedStatus is a shortcut function for invalidating the cached
// representation one status in the home timeline and all list timelines of the
// given accountID. It should only be called in cases where a status update
// does *not* need to be passed into the processor via the worker queue, since
// such invalidation will, in that case, be handled by the processor instead.
func (p *Processor) InvalidateTimelinedStatus(ctx context.Context, accountID string, statusID string) error {
	// Get lists first + bail if this fails.
	lists, err := p.state.DB.GetListsForAccountID(ctx, accountID)
	if err != nil {
		return gtserror.Newf("db error getting lists for account %s: %w", accountID, err)
	}

	// Start new log entry with
	// the above calling func's name.
	l := log.
		WithContext(ctx).
		WithField("caller", log.Caller(3)).
		WithField("accountID", accountID).
		WithField("statusID", statusID)

	// Unprepare item from home + list timelines, just log
	// if something goes wrong since this is not a showstopper.

	if err := p.state.Timelines.Home.UnprepareItem(ctx, accountID, statusID); err != nil {
		l.Errorf("error unpreparing item from home timeline: %v", err)
	}

	for _, list := range lists {
		if err := p.state.Timelines.List.UnprepareItem(ctx, list.ID, statusID); err != nil {
			l.Errorf("error unpreparing item from list timeline %s: %v", list.ID, err)
		}
	}

	return nil
}
