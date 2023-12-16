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

package workers

import (
	"context"
	"errors"

	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/processing/account"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

// clientAPI wraps processing functions
// specifically for messages originating
// from the client/REST API.
type clientAPI struct {
	state      *state.State
	converter  *typeutils.Converter
	surface    *surface
	federate   *federate
	wipeStatus wipeStatus
	account    *account.Processor
}

func (p *Processor) EnqueueClientAPI(cctx context.Context, msgs ...messages.FromClientAPI) {
	_ = p.workers.ClientAPI.MustEnqueueCtx(cctx, func(wctx context.Context) {
		// Copy caller ctx values to worker's.
		wctx = gtscontext.WithValues(wctx, cctx)

		// Process worker messages.
		for _, msg := range msgs {
			if err := p.ProcessFromClientAPI(wctx, msg); err != nil {
				log.Errorf(wctx, "error processing client API message: %v", err)
			}
		}
	})
}

func (p *Processor) ProcessFromClientAPI(ctx context.Context, cMsg messages.FromClientAPI) error {
	// Allocate new log fields slice
	fields := make([]kv.Field, 3, 4)
	fields[0] = kv.Field{"activityType", cMsg.APActivityType}
	fields[1] = kv.Field{"objectType", cMsg.APObjectType}
	fields[2] = kv.Field{"fromAccount", cMsg.OriginAccount.Username}

	// Include GTSModel in logs if appropriate.
	if cMsg.GTSModel != nil &&
		log.Level() >= level.DEBUG {
		fields = append(fields, kv.Field{
			"model", cMsg.GTSModel,
		})
	}

	l := log.WithContext(ctx).WithFields(fields...)
	l.Info("processing from client API")

	switch cMsg.APActivityType {

	// CREATE SOMETHING
	case ap.ActivityCreate:
		switch cMsg.APObjectType {

		// CREATE PROFILE/ACCOUNT
		case ap.ObjectProfile, ap.ActorPerson:
			return p.clientAPI.CreateAccount(ctx, cMsg)

		// CREATE NOTE/STATUS
		case ap.ObjectNote:
			return p.clientAPI.CreateStatus(ctx, cMsg)

		// CREATE QUESTION
		// (note we don't handle poll *votes* as AS
		// question type when federating (just notes),
		// but it makes for a nicer type switch here.
		case ap.ActivityQuestion:
			return p.clientAPI.CreatePollVote(ctx, cMsg)

		// CREATE FOLLOW (request)
		case ap.ActivityFollow:
			return p.clientAPI.CreateFollowReq(ctx, cMsg)

		// CREATE LIKE/FAVE
		case ap.ActivityLike:
			return p.clientAPI.CreateLike(ctx, cMsg)

		// CREATE ANNOUNCE/BOOST
		case ap.ActivityAnnounce:
			return p.clientAPI.CreateAnnounce(ctx, cMsg)

		// CREATE BLOCK
		case ap.ActivityBlock:
			return p.clientAPI.CreateBlock(ctx, cMsg)
		}

	// UPDATE SOMETHING
	case ap.ActivityUpdate:
		switch cMsg.APObjectType {

		// UPDATE NOTE/STATUS
		case ap.ObjectNote:
			return p.clientAPI.UpdateStatus(ctx, cMsg)

		// UPDATE PROFILE/ACCOUNT
		case ap.ObjectProfile, ap.ActorPerson:
			return p.clientAPI.UpdateAccount(ctx, cMsg)

		// UPDATE A FLAG/REPORT (mark as resolved/closed)
		case ap.ActivityFlag:
			return p.clientAPI.UpdateReport(ctx, cMsg)
		}

	// ACCEPT SOMETHING
	case ap.ActivityAccept:
		switch cMsg.APObjectType { //nolint:gocritic

		// ACCEPT FOLLOW (request)
		case ap.ActivityFollow:
			return p.clientAPI.AcceptFollow(ctx, cMsg)
		}

	// REJECT SOMETHING
	case ap.ActivityReject:
		switch cMsg.APObjectType { //nolint:gocritic

		// REJECT FOLLOW (request)
		case ap.ActivityFollow:
			return p.clientAPI.RejectFollowRequest(ctx, cMsg)
		}

	// UNDO SOMETHING
	case ap.ActivityUndo:
		switch cMsg.APObjectType {

		// UNDO FOLLOW (request)
		case ap.ActivityFollow:
			return p.clientAPI.UndoFollow(ctx, cMsg)

		// UNDO BLOCK
		case ap.ActivityBlock:
			return p.clientAPI.UndoBlock(ctx, cMsg)

		// UNDO LIKE/FAVE
		case ap.ActivityLike:
			return p.clientAPI.UndoFave(ctx, cMsg)

		// UNDO ANNOUNCE/BOOST
		case ap.ActivityAnnounce:
			return p.clientAPI.UndoAnnounce(ctx, cMsg)
		}

	// DELETE SOMETHING
	case ap.ActivityDelete:
		switch cMsg.APObjectType {

		// DELETE NOTE/STATUS
		case ap.ObjectNote:
			return p.clientAPI.DeleteStatus(ctx, cMsg)

		// DELETE PROFILE/ACCOUNT
		case ap.ObjectProfile, ap.ActorPerson:
			return p.clientAPI.DeleteAccount(ctx, cMsg)
		}

	// FLAG/REPORT SOMETHING
	case ap.ActivityFlag:
		switch cMsg.APObjectType { //nolint:gocritic

		// FLAG/REPORT A PROFILE
		case ap.ObjectProfile:
			return p.clientAPI.ReportAccount(ctx, cMsg)
		}
	}

	return gtserror.Newf("unhandled: %s %s", cMsg.APActivityType, cMsg.APObjectType)
}

func (p *clientAPI) CreateAccount(ctx context.Context, cMsg messages.FromClientAPI) error {
	account, ok := cMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Account", cMsg.GTSModel)
	}

	// Send a confirmation email to the newly created account.
	user, err := p.state.DB.GetUserByAccountID(ctx, account.ID)
	if err != nil {
		return gtserror.Newf("db error getting user for account id %s: %w", account.ID, err)
	}

	if err := p.surface.emailPleaseConfirm(ctx, user, account.Username); err != nil {
		log.Errorf(ctx, "error emailing confirm: %v", err)
	}

	return nil
}

func (p *clientAPI) CreateStatus(ctx context.Context, cMsg messages.FromClientAPI) error {
	status, ok := cMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", cMsg.GTSModel)
	}

	if err := p.surface.timelineAndNotifyStatus(ctx, status); err != nil {
		log.Errorf(ctx, "error timelining and notifying status: %v", err)
	}

	if status.InReplyToID != "" {
		// Interaction counts changed on the replied status;
		// uncache the prepared version from all timelines.
		p.surface.invalidateStatusFromTimelines(ctx, status.InReplyToID)
	}

	if err := p.federate.CreateStatus(ctx, status); err != nil {
		log.Errorf(ctx, "error federating status: %v", err)
	}

	return nil
}

func (p *clientAPI) CreatePollVote(ctx context.Context, cMsg messages.FromClientAPI) error {
	// Cast the create poll vote attached to message.
	vote, ok := cMsg.GTSModel.(*gtsmodel.PollVote)
	if !ok {
		return gtserror.Newf("cannot cast %T -> *gtsmodel.Pollvote", cMsg.GTSModel)
	}

	// Ensure the vote is fully populated in order to get original poll.
	if err := p.state.DB.PopulatePollVote(ctx, vote); err != nil {
		return gtserror.Newf("error populating poll vote from db: %w", err)
	}

	// Ensure the poll on the vote is fully populated to get origin status.
	if err := p.state.DB.PopulatePoll(ctx, vote.Poll); err != nil {
		return gtserror.Newf("error populating poll from db: %w", err)
	}

	// Get the origin status,
	// (also set the poll on it).
	status := vote.Poll.Status
	status.Poll = vote.Poll

	// Interaction counts changed on the source status, uncache from timelines.
	p.surface.invalidateStatusFromTimelines(ctx, vote.Poll.StatusID)

	if *status.Local {
		// These are poll votes in a local status, we only need to
		// federate the updated status model with latest vote counts.
		if err := p.federate.UpdateStatus(ctx, status); err != nil {
			log.Errorf(ctx, "error federating status update: %v", err)
		}
	} else {
		// These are votes in a remote poll, federate to origin the new poll vote(s).
		if err := p.federate.CreatePollVote(ctx, vote.Poll, vote); err != nil {
			log.Errorf(ctx, "error federating poll vote: %v", err)
		}
	}

	return nil
}

func (p *clientAPI) CreateFollowReq(ctx context.Context, cMsg messages.FromClientAPI) error {
	followRequest, ok := cMsg.GTSModel.(*gtsmodel.FollowRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.FollowRequest", cMsg.GTSModel)
	}

	if err := p.surface.notifyFollowRequest(ctx, followRequest); err != nil {
		log.Errorf(ctx, "error notifying follow request: %v", err)
	}

	// Convert the follow request to follow model (requests are sent as follows).
	follow := p.converter.FollowRequestToFollow(ctx, followRequest)

	if err := p.federate.Follow(
		ctx,
		follow,
	); err != nil {
		log.Errorf(ctx, "error federating follow request: %v", err)
	}

	return nil
}

func (p *clientAPI) CreateLike(ctx context.Context, cMsg messages.FromClientAPI) error {
	fave, ok := cMsg.GTSModel.(*gtsmodel.StatusFave)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.StatusFave", cMsg.GTSModel)
	}

	// Ensure fave populated.
	if err := p.state.DB.PopulateStatusFave(ctx, fave); err != nil {
		return gtserror.Newf("error populating status fave: %w", err)
	}

	if err := p.surface.notifyFave(ctx, fave); err != nil {
		log.Errorf(ctx, "error notifying fave: %v", err)
	}

	// Interaction counts changed on the faved status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, fave.StatusID)

	if err := p.federate.Like(ctx, fave); err != nil {
		log.Errorf(ctx, "error federating like: %v", err)
	}

	return nil
}

func (p *clientAPI) CreateAnnounce(ctx context.Context, cMsg messages.FromClientAPI) error {
	boost, ok := cMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", cMsg.GTSModel)
	}

	// Timeline and notify the boost wrapper status.
	if err := p.surface.timelineAndNotifyStatus(ctx, boost); err != nil {
		log.Errorf(ctx, "error timelining and notifying status: %v", err)
	}

	// Notify the boost target account.
	if err := p.surface.notifyAnnounce(ctx, boost); err != nil {
		log.Errorf(ctx, "error notifying boost: %v", err)
	}

	// Interaction counts changed on the boosted status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, boost.BoostOfID)

	if err := p.federate.Announce(ctx, boost); err != nil {
		log.Errorf(ctx, "error federating announce: %v", err)
	}

	return nil
}

func (p *clientAPI) CreateBlock(ctx context.Context, cMsg messages.FromClientAPI) error {
	block, ok := cMsg.GTSModel.(*gtsmodel.Block)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Block", cMsg.GTSModel)
	}

	// Remove blockee's statuses from blocker's timeline.
	if err := p.state.Timelines.Home.WipeItemsFromAccountID(
		ctx,
		block.AccountID,
		block.TargetAccountID,
	); err != nil {
		return gtserror.Newf("error wiping timeline items for block: %w", err)
	}

	// Remove blocker's statuses from blockee's timeline.
	if err := p.state.Timelines.Home.WipeItemsFromAccountID(
		ctx,
		block.TargetAccountID,
		block.AccountID,
	); err != nil {
		return gtserror.Newf("error wiping timeline items for block: %w", err)
	}

	// TODO: same with notifications?
	// TODO: same with bookmarks?

	if err := p.federate.Block(ctx, block); err != nil {
		log.Errorf(ctx, "error federating block: %v", err)
	}

	return nil
}

func (p *clientAPI) UpdateStatus(ctx context.Context, cMsg messages.FromClientAPI) error {
	// Cast the updated Status model attached to msg.
	status, ok := cMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("cannot cast %T -> *gtsmodel.Status", cMsg.GTSModel)
	}

	// Federate the updated status changes out remotely.
	if err := p.federate.UpdateStatus(ctx, status); err != nil {
		log.Errorf(ctx, "error federating status update: %v", err)
	}

	// Status representation has changed, invalidate from timelines.
	p.surface.invalidateStatusFromTimelines(ctx, status.ID)

	if status.Poll != nil && status.Poll.Closing {

		// If the latest status has a newly closed poll, at least compared
		// to the existing version, then notify poll close to all voters.
		if err := p.surface.notifyPollClose(ctx, status); err != nil {
			log.Errorf(ctx, "error notifying poll close: %v", err)
		}
	}

	// Push message that the status has been edited to streams.
	if err := p.surface.timelineStatusUpdate(ctx, status); err != nil {
		log.Errorf(ctx, "error streaming status edit: %v", err)
	}

	return nil
}

func (p *clientAPI) UpdateAccount(ctx context.Context, cMsg messages.FromClientAPI) error {
	account, ok := cMsg.GTSModel.(*gtsmodel.Account)
	if !ok {
		return gtserror.Newf("cannot cast %T -> *gtsmodel.Account", cMsg.GTSModel)
	}

	if err := p.federate.UpdateAccount(ctx, account); err != nil {
		log.Errorf(ctx, "error federating account update: %v", err)
	}

	return nil
}

func (p *clientAPI) UpdateReport(ctx context.Context, cMsg messages.FromClientAPI) error {
	report, ok := cMsg.GTSModel.(*gtsmodel.Report)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Report", cMsg.GTSModel)
	}

	if report.Account.IsRemote() {
		// Report creator is a remote account,
		// we shouldn't try to email them!
		return nil
	}

	if err := p.surface.emailReportClosed(ctx, report); err != nil {
		log.Errorf(ctx, "error emailing report closed: %v", err)
	}

	return nil
}

func (p *clientAPI) AcceptFollow(ctx context.Context, cMsg messages.FromClientAPI) error {
	follow, ok := cMsg.GTSModel.(*gtsmodel.Follow)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Follow", cMsg.GTSModel)
	}

	if err := p.surface.notifyFollow(ctx, follow); err != nil {
		log.Errorf(ctx, "error notifying follow: %v", err)
	}

	if err := p.federate.AcceptFollow(ctx, follow); err != nil {
		log.Errorf(ctx, "error federating follow accept: %v", err)
	}

	return nil
}

func (p *clientAPI) RejectFollowRequest(ctx context.Context, cMsg messages.FromClientAPI) error {
	followReq, ok := cMsg.GTSModel.(*gtsmodel.FollowRequest)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.FollowRequest", cMsg.GTSModel)
	}

	if err := p.federate.RejectFollow(
		ctx,
		p.converter.FollowRequestToFollow(ctx, followReq),
	); err != nil {
		log.Errorf(ctx, "error federating follow reject: %v", err)
	}

	return nil
}

func (p *clientAPI) UndoFollow(ctx context.Context, cMsg messages.FromClientAPI) error {
	follow, ok := cMsg.GTSModel.(*gtsmodel.Follow)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Follow", cMsg.GTSModel)
	}

	if err := p.federate.UndoFollow(ctx, follow); err != nil {
		log.Errorf(ctx, "error federating follow undo: %v", err)
	}

	return nil
}

func (p *clientAPI) UndoBlock(ctx context.Context, cMsg messages.FromClientAPI) error {
	block, ok := cMsg.GTSModel.(*gtsmodel.Block)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Block", cMsg.GTSModel)
	}

	if err := p.federate.UndoBlock(ctx, block); err != nil {
		log.Errorf(ctx, "error federating block undo: %v", err)
	}

	return nil
}

func (p *clientAPI) UndoFave(ctx context.Context, cMsg messages.FromClientAPI) error {
	statusFave, ok := cMsg.GTSModel.(*gtsmodel.StatusFave)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.StatusFave", cMsg.GTSModel)
	}

	// Interaction counts changed on the faved status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, statusFave.StatusID)

	if err := p.federate.UndoLike(ctx, statusFave); err != nil {
		log.Errorf(ctx, "error federating like undo: %v", err)
	}

	return nil
}

func (p *clientAPI) UndoAnnounce(ctx context.Context, cMsg messages.FromClientAPI) error {
	status, ok := cMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", cMsg.GTSModel)
	}

	if err := p.state.DB.DeleteStatusByID(ctx, status.ID); err != nil {
		return gtserror.Newf("db error deleting status: %w", err)
	}

	if err := p.surface.deleteStatusFromTimelines(ctx, status.ID); err != nil {
		log.Errorf(ctx, "error removing timelined status: %v", err)
	}

	// Interaction counts changed on the boosted status;
	// uncache the prepared version from all timelines.
	p.surface.invalidateStatusFromTimelines(ctx, status.BoostOfID)

	if err := p.federate.UndoAnnounce(ctx, status); err != nil {
		log.Errorf(ctx, "error federating announce undo: %v", err)
	}

	return nil
}

func (p *clientAPI) DeleteStatus(ctx context.Context, cMsg messages.FromClientAPI) error {
	// Don't delete attachments, just unattach them:
	// this request comes from the client API and the
	// poster may want to use attachments again later.
	const deleteAttachments = false

	status, ok := cMsg.GTSModel.(*gtsmodel.Status)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Status", cMsg.GTSModel)
	}

	// Try to populate status structs if possible,
	// in order to more thoroughly remove them.
	if err := p.state.DB.PopulateStatus(
		ctx, status,
	); err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("db error populating status: %w", err)
	}

	if err := p.wipeStatus(ctx, status, deleteAttachments); err != nil {
		log.Errorf(ctx, "error wiping status: %v", err)
	}

	if status.InReplyToID != "" {
		// Interaction counts changed on the replied status;
		// uncache the prepared version from all timelines.
		p.surface.invalidateStatusFromTimelines(ctx, status.InReplyToID)
	}

	if err := p.federate.DeleteStatus(ctx, status); err != nil {
		log.Errorf(ctx, "error federating status delete: %v", err)
	}

	return nil
}

func (p *clientAPI) DeleteAccount(ctx context.Context, cMsg messages.FromClientAPI) error {
	// The originID of the delete, one of:
	//   - ID of a domain block, for which
	//     this account delete is a side effect.
	//   - ID of the deleted account itself (self delete).
	//   - ID of an admin account (account suspension).
	var originID string

	if domainBlock, ok := cMsg.GTSModel.(*gtsmodel.DomainBlock); ok {
		// Origin is a domain block.
		originID = domainBlock.ID
	} else {
		// Origin is whichever account
		// originated this message.
		originID = cMsg.OriginAccount.ID
	}

	if err := p.federate.DeleteAccount(ctx, cMsg.TargetAccount); err != nil {
		log.Errorf(ctx, "error federating account delete: %v", err)
	}

	if err := p.account.Delete(ctx, cMsg.TargetAccount, originID); err != nil {
		log.Errorf(ctx, "error deleting account: %v", err)
	}

	return nil
}

func (p *clientAPI) ReportAccount(ctx context.Context, cMsg messages.FromClientAPI) error {
	report, ok := cMsg.GTSModel.(*gtsmodel.Report)
	if !ok {
		return gtserror.Newf("%T not parseable as *gtsmodel.Report", cMsg.GTSModel)
	}

	// Federate this report to the
	// remote instance if desired.
	if *report.Forwarded {
		if err := p.federate.Flag(ctx, report); err != nil {
			log.Errorf(ctx, "error federating flag: %v", err)
		}
	}

	if err := p.surface.emailReportOpened(ctx, report); err != nil {
		log.Errorf(ctx, "error emailing report opened: %v", err)
	}

	return nil
}
