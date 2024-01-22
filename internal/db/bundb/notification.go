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

package bundb

import (
	"context"
	"errors"
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

type notificationDB struct {
	db    *DB
	state *state.State
}

func (n *notificationDB) GetNotificationByID(ctx context.Context, id string) (*gtsmodel.Notification, error) {
	return n.getNotification(
		ctx,
		"ID",
		func(notif *gtsmodel.Notification) error {
			return n.db.NewSelect().
				Model(notif).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (n *notificationDB) GetNotification(
	ctx context.Context,
	notificationType gtsmodel.NotificationType,
	targetAccountID string,
	originAccountID string,
	statusID string,
) (*gtsmodel.Notification, error) {
	return n.getNotification(
		ctx,
		"NotificationType,TargetAccountID,OriginAccountID,StatusID",
		func(notif *gtsmodel.Notification) error {
			return n.db.NewSelect().
				Model(notif).
				Where("? = ?", bun.Ident("notification_type"), notificationType).
				Where("? = ?", bun.Ident("target_account_id"), targetAccountID).
				Where("? = ?", bun.Ident("origin_account_id"), originAccountID).
				Where("? = ?", bun.Ident("status_id"), statusID).
				Scan(ctx)
		},
		notificationType, targetAccountID, originAccountID, statusID,
	)
}

func (n *notificationDB) getNotification(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Notification) error, keyParts ...any) (*gtsmodel.Notification, error) {
	// Fetch notification from cache with loader callback
	notif, err := n.state.Caches.GTS.Notification.LoadOne(lookup, func() (*gtsmodel.Notification, error) {
		var notif gtsmodel.Notification

		// Not cached! Perform database query
		if err := dbQuery(&notif); err != nil {
			return nil, err
		}

		return &notif, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return notif, nil
	}

	if err := n.state.DB.PopulateNotification(ctx, notif); err != nil {
		return nil, err
	}

	return notif, nil
}

func (n *notificationDB) GetNotificationsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Notification, error) {
	// Preallocate at-worst possible length.
	uncached := make([]string, 0, len(ids))

	// Load all notif IDs via cache loader callbacks.
	notifs, err := n.state.Caches.GTS.Notification.Load("ID",

		// Load cached + check for uncached.
		func(load func(keyParts ...any) bool) {
			for _, id := range ids {
				if !load(id) {
					uncached = append(uncached, id)
				}
			}
		},

		// Uncached notification loader function.
		func() ([]*gtsmodel.Notification, error) {
			// Preallocate expected length of uncached notifications.
			notifs := make([]*gtsmodel.Notification, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := n.db.NewSelect().
				Model(&notifs).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return notifs, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the notifs by their
	// IDs to ensure in correct order.
	getID := func(n *gtsmodel.Notification) string { return n.ID }
	util.OrderBy(notifs, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return notifs, nil
	}

	// Populate all loaded notifs, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	notifs = slices.DeleteFunc(notifs, func(notif *gtsmodel.Notification) bool {
		if err := n.PopulateNotification(ctx, notif); err != nil {
			log.Errorf(ctx, "error populating notif %s: %v", notif.ID, err)
			return true
		}
		return false
	})

	return notifs, nil
}

func (n *notificationDB) PopulateNotification(ctx context.Context, notif *gtsmodel.Notification) error {
	var (
		errs gtserror.MultiError
		err  error
	)

	if notif.TargetAccount == nil {
		notif.TargetAccount, err = n.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			notif.TargetAccountID,
		)
		if err != nil {
			errs.Appendf("error populating notif target account: %w", err)
		}
	}

	if notif.OriginAccount == nil {
		notif.OriginAccount, err = n.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			notif.OriginAccountID,
		)
		if err != nil {
			errs.Appendf("error populating notif origin account: %w", err)
		}
	}

	if notif.StatusID != "" && notif.Status == nil {
		notif.Status, err = n.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			notif.StatusID,
		)
		if err != nil {
			errs.Appendf("error populating notif status: %w", err)
		}
	}

	return errs.Combine()
}

func (n *notificationDB) GetAccountNotifications(
	ctx context.Context,
	accountID string,
	maxID string,
	sinceID string,
	minID string,
	limit int,
	excludeTypes []string,
) ([]*gtsmodel.Notification, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	var (
		notifIDs    = make([]string, 0, limit)
		frontToBack = true
	)

	q := n.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("notifications"), bun.Ident("notification")).
		Column("notification.id")

	if maxID == "" {
		maxID = id.Highest
	}

	// Return only notifs LOWER (ie., older) than maxID.
	q = q.Where("? < ?", bun.Ident("notification.id"), maxID)

	if sinceID != "" {
		// Return only notifs HIGHER (ie., newer) than sinceID.
		q = q.Where("? > ?", bun.Ident("notification.id"), sinceID)
	}

	if minID != "" {
		// Return only notifs HIGHER (ie., newer) than minID.
		q = q.Where("? > ?", bun.Ident("notification.id"), minID)

		frontToBack = false // page up
	}

	for _, excludeType := range excludeTypes {
		// Filter out unwanted notif types.
		q = q.Where("? != ?", bun.Ident("notification.notification_type"), excludeType)
	}

	// Return only notifs for this account.
	q = q.Where("? = ?", bun.Ident("notification.target_account_id"), accountID)

	if limit > 0 {
		q = q.Limit(limit)
	}

	if frontToBack {
		// Page down.
		q = q.Order("notification.id DESC")
	} else {
		// Page up.
		q = q.Order("notification.id ASC")
	}

	if err := q.Scan(ctx, &notifIDs); err != nil {
		return nil, err
	}

	if len(notifIDs) == 0 {
		return nil, nil
	}

	// If we're paging up, we still want notifications
	// to be sorted by ID desc, so reverse ids slice.
	// https://zchee.github.io/golang-wiki/SliceTricks/#reversing
	if !frontToBack {
		for l, r := 0, len(notifIDs)-1; l < r; l, r = l+1, r-1 {
			notifIDs[l], notifIDs[r] = notifIDs[r], notifIDs[l]
		}
	}

	// Fetch notification models by their IDs.
	return n.GetNotificationsByIDs(ctx, notifIDs)
}

func (n *notificationDB) PutNotification(ctx context.Context, notif *gtsmodel.Notification) error {
	return n.state.Caches.GTS.Notification.Store(notif, func() error {
		_, err := n.db.NewInsert().Model(notif).Exec(ctx)
		return err
	})
}

func (n *notificationDB) DeleteNotificationByID(ctx context.Context, id string) error {
	defer n.state.Caches.GTS.Notification.Invalidate("ID", id)

	// Load notif into cache before attempting a delete,
	// as we need it cached in order to trigger the invalidate
	// callback. This in turn invalidates others.
	_, err := n.GetNotificationByID(gtscontext.SetBarebones(ctx), id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// not an issue.
			err = nil
		}
		return err
	}

	// Finally delete notif from DB.
	_, err = n.db.NewDelete().
		TableExpr("? AS ?", bun.Ident("notifications"), bun.Ident("notification")).
		Where("? = ?", bun.Ident("notification.id"), id).
		Exec(ctx)
	return err
}

func (n *notificationDB) DeleteNotifications(ctx context.Context, types []string, targetAccountID string, originAccountID string) error {
	if targetAccountID == "" && originAccountID == "" {
		return errors.New("DeleteNotifications: one of targetAccountID or originAccountID must be set")
	}

	var notifIDs []string

	q := n.db.
		NewSelect().
		Column("id").
		Table("notifications")

	if len(types) > 0 {
		q = q.Where("? IN (?)", bun.Ident("notification_type"), bun.In(types))
	}

	if targetAccountID != "" {
		q = q.Where("? = ?", bun.Ident("target_account_id"), targetAccountID)
	}

	if originAccountID != "" {
		q = q.Where("? = ?", bun.Ident("origin_account_id"), originAccountID)
	}

	if _, err := q.Exec(ctx, &notifIDs); err != nil {
		return err
	}

	defer func() {
		// Invalidate all IDs on return.
		for _, id := range notifIDs {
			n.state.Caches.GTS.Notification.Invalidate("ID", id)
		}
	}()

	// Load all notif into cache, this *really* isn't great
	// but it is the only way we can ensure we invalidate all
	// related caches correctly (e.g. visibility).
	for _, id := range notifIDs {
		_, err := n.GetNotificationByID(ctx, id)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return err
		}
	}

	// Finally delete all from DB.
	_, err := n.db.NewDelete().
		Table("notifications").
		Where("? IN (?)", bun.Ident("id"), bun.In(notifIDs)).
		Exec(ctx)
	return err
}

func (n *notificationDB) DeleteNotificationsForStatus(ctx context.Context, statusID string) error {
	var notifIDs []string

	q := n.db.
		NewSelect().
		Column("id").
		Table("notifications").
		Where("? = ?", bun.Ident("status_id"), statusID)

	if _, err := q.Exec(ctx, &notifIDs); err != nil {
		return err
	}

	defer func() {
		// Invalidate all IDs on return.
		for _, id := range notifIDs {
			n.state.Caches.GTS.Notification.Invalidate("ID", id)
		}
	}()

	// Load all notif into cache, this *really* isn't great
	// but it is the only way we can ensure we invalidate all
	// related caches correctly (e.g. visibility).
	for _, id := range notifIDs {
		_, err := n.GetNotificationByID(ctx, id)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return err
		}
	}

	// Finally delete all from DB.
	_, err := n.db.NewDelete().
		Table("notifications").
		Where("? IN (?)", bun.Ident("id"), bun.In(notifIDs)).
		Exec(ctx)
	return err
}
