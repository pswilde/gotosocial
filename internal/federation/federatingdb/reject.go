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

package federatingdb

import (
	"context"
	"errors"
	"fmt"

	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

func (f *federatingDB) Reject(ctx context.Context, reject vocab.ActivityStreamsReject) error {
	if log.Level() >= level.DEBUG {
		i, err := marshalItem(reject)
		if err != nil {
			return err
		}
		l := log.WithContext(ctx).
			WithField("reject", i)
		l.Debug("entering Reject")
	}

	receivingAccount, _, internal := extractFromCtx(ctx)
	if internal {
		return nil // Already processed.
	}

	for _, obj := range ap.ExtractObjects(reject) {

		if obj.IsIRI() {
			// we have just the URI of whatever is being rejected, so we need to find out what it is
			rejectedObjectIRI := obj.GetIRI()
			if uris.IsFollowPath(rejectedObjectIRI) {
				// REJECT FOLLOW
				followReq, err := f.state.DB.GetFollowRequestByURI(ctx, rejectedObjectIRI.String())
				if err != nil {
					return fmt.Errorf("Reject: couldn't get follow request with id %s from the database: %s", rejectedObjectIRI.String(), err)
				}

				// make sure the addressee of the original follow is the same as whatever inbox this landed in
				if followReq.AccountID != receivingAccount.ID {
					return errors.New("Reject: follow object account and inbox account were not the same")
				}

				return f.state.DB.RejectFollowRequest(ctx, followReq.AccountID, followReq.TargetAccountID)
			}
		}

		if t := obj.GetType(); t != nil {
			// we have the whole object so we can figure out what we're rejecting
			// REJECT FOLLOW
			asFollow, ok := t.(vocab.ActivityStreamsFollow)
			if !ok {
				return errors.New("Reject: couldn't parse follow into vocab.ActivityStreamsFollow")
			}

			// convert the follow to something we can understand
			gtsFollow, err := f.converter.ASFollowToFollow(ctx, asFollow)
			if err != nil {
				return fmt.Errorf("Reject: error converting asfollow to gtsfollow: %s", err)
			}

			// make sure the addressee of the original follow is the same as whatever inbox this landed in
			if gtsFollow.AccountID != receivingAccount.ID {
				return errors.New("Reject: follow object account and inbox account were not the same")
			}

			return f.state.DB.RejectFollowRequest(ctx, gtsFollow.AccountID, gtsFollow.TargetAccountID)
		}
	}

	return nil
}
