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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type pollDB struct {
	db    *DB
	state *state.State
}

func (p *pollDB) GetPollByID(ctx context.Context, id string) (*gtsmodel.Poll, error) {
	return p.getPoll(
		ctx,
		"ID",
		func(poll *gtsmodel.Poll) error {
			return p.db.NewSelect().
				Model(poll).
				Where("? = ?", bun.Ident("poll.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (p *pollDB) getPoll(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Poll) error, keyParts ...any) (*gtsmodel.Poll, error) {
	// Fetch poll from database cache with loader callback
	poll, err := p.state.Caches.GTS.Poll().Load(lookup, func() (*gtsmodel.Poll, error) {
		var poll gtsmodel.Poll

		// Not cached! Perform database query.
		if err := dbQuery(&poll); err != nil {
			return nil, err
		}

		// Ensure vote slice
		// is non nil and set.
		poll.CheckVotes()

		return &poll, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return poll, nil
	}

	// Further populate the poll fields where applicable.
	if err := p.PopulatePoll(ctx, poll); err != nil {
		return nil, err
	}

	return poll, nil
}

func (p *pollDB) GetOpenPolls(ctx context.Context) ([]*gtsmodel.Poll, error) {
	var pollIDs []string

	// Select all polls with unset `closed_at` time.
	if err := p.db.NewSelect().
		Table("polls").
		Column("polls.id").
		Join("JOIN ? ON ? = ?", bun.Ident("statuses"), bun.Ident("polls.id"), bun.Ident("statuses.poll_id")).
		Where("? = true", bun.Ident("statuses.local")).
		Where("? IS NULL", bun.Ident("polls.closed_at")).
		Scan(ctx, &pollIDs); err != nil {
		return nil, err
	}

	// Preallocate a slice to contain the poll models.
	polls := make([]*gtsmodel.Poll, 0, len(pollIDs))

	for _, id := range pollIDs {
		// Attempt to fetch poll from DB.
		poll, err := p.GetPollByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting poll %s: %v", id, err)
			continue
		}

		// Append poll to return slice.
		polls = append(polls, poll)
	}

	return polls, nil
}

func (p *pollDB) PopulatePoll(ctx context.Context, poll *gtsmodel.Poll) error {
	var (
		err  error
		errs gtserror.MultiError
	)

	if poll.Status == nil {
		// Vote account is not set, fetch from database.
		poll.Status, err = p.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			poll.StatusID,
		)
		if err != nil {
			errs.Appendf("error populating poll status: %w", err)
		}
	}

	return errs.Combine()
}

func (p *pollDB) PutPoll(ctx context.Context, poll *gtsmodel.Poll) error {
	// Ensure vote slice
	// is non nil and set.
	poll.CheckVotes()

	return p.state.Caches.GTS.Poll().Store(poll, func() error {
		_, err := p.db.NewInsert().Model(poll).Exec(ctx)
		return err
	})
}

func (p *pollDB) UpdatePoll(ctx context.Context, poll *gtsmodel.Poll, cols ...string) error {
	// Ensure vote slice
	// is non nil and set.
	poll.CheckVotes()

	return p.state.Caches.GTS.Poll().Store(poll, func() error {
		return p.db.RunInTx(ctx, func(tx Tx) error {
			// Update the status' "updated_at" field.
			if _, err := tx.NewUpdate().
				Table("statuses").
				Where("? = ?", bun.Ident("id"), poll.StatusID).
				SetColumn("updated_at", "?", time.Now()).
				Exec(ctx); err != nil {
				return err
			}

			// Finally, update poll
			// columns in database.
			_, err := tx.NewUpdate().
				Model(poll).
				Column(cols...).
				Where("? = ?", bun.Ident("id"), poll.ID).
				Exec(ctx)
			return err
		})
	})
}

func (p *pollDB) DeletePollByID(ctx context.Context, id string) error {
	// Delete poll by ID from database.
	if _, err := p.db.NewDelete().
		Table("polls").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return err
	}

	// Invalidate poll by ID from cache.
	p.state.Caches.GTS.Poll().Invalidate("ID", id)
	p.state.Caches.GTS.PollVoteIDs().Invalidate(id)

	return nil
}

func (p *pollDB) GetPollVoteByID(ctx context.Context, id string) (*gtsmodel.PollVote, error) {
	return p.getPollVote(
		ctx,
		"ID",
		func(vote *gtsmodel.PollVote) error {
			return p.db.NewSelect().
				Model(vote).
				Where("? = ?", bun.Ident("poll_vote.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (p *pollDB) GetPollVoteBy(ctx context.Context, pollID string, accountID string) (*gtsmodel.PollVote, error) {
	return p.getPollVote(
		ctx,
		"PollID.AccountID",
		func(vote *gtsmodel.PollVote) error {
			return p.db.NewSelect().
				Model(vote).
				Where("? = ?", bun.Ident("poll_vote.account_id"), accountID).
				Where("? = ?", bun.Ident("poll_vote.poll_id"), pollID).
				Scan(ctx)
		},
		pollID,
		accountID,
	)
}

func (p *pollDB) getPollVote(ctx context.Context, lookup string, dbQuery func(*gtsmodel.PollVote) error, keyParts ...any) (*gtsmodel.PollVote, error) {
	// Fetch vote from database cache with loader callback
	vote, err := p.state.Caches.GTS.PollVote().Load(lookup, func() (*gtsmodel.PollVote, error) {
		var vote gtsmodel.PollVote

		// Not cached! Perform database query.
		if err := dbQuery(&vote); err != nil {
			return nil, err
		}

		return &vote, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return vote, nil
	}

	// Further populate the vote fields where applicable.
	if err := p.PopulatePollVote(ctx, vote); err != nil {
		return nil, err
	}

	return vote, nil
}

func (p *pollDB) GetPollVotes(ctx context.Context, pollID string) ([]*gtsmodel.PollVote, error) {
	voteIDs, err := p.state.Caches.GTS.PollVoteIDs().Load(pollID, func() ([]string, error) {
		var voteIDs []string

		// Vote IDs not in cache, perform DB query!
		q := newSelectPollVotes(p.db, pollID)
		if _, err := q.Exec(ctx, &voteIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return voteIDs, nil
	})
	if err != nil {
		return nil, err
	}

	// Preallocate slice of expected length.
	votes := make([]*gtsmodel.PollVote, 0, len(voteIDs))

	for _, id := range voteIDs {
		// Fetch poll vote model for this ID.
		vote, err := p.GetPollVoteByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting poll vote %s: %v", id, err)
			continue
		}

		// Append to return slice.
		votes = append(votes, vote)
	}

	return votes, nil
}

func (p *pollDB) PopulatePollVote(ctx context.Context, vote *gtsmodel.PollVote) error {
	var (
		err  error
		errs gtserror.MultiError
	)

	if vote.Account == nil {
		// Vote account is not set, fetch from database.
		vote.Account, err = p.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			vote.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating vote account: %w", err)
		}
	}

	if vote.Poll == nil {
		// Vote poll is not set, fetch from database.
		vote.Poll, err = p.GetPollByID(
			gtscontext.SetBarebones(ctx),
			vote.PollID,
		)
		if err != nil {
			errs.Appendf("error populating vote poll: %w", err)
		}
	}

	return errs.Combine()
}

func (p *pollDB) PutPollVote(ctx context.Context, vote *gtsmodel.PollVote) error {
	return p.state.Caches.GTS.PollVote().Store(vote, func() error {
		return p.db.RunInTx(ctx, func(tx Tx) error {
			// Try insert vote into database.
			if _, err := tx.NewInsert().
				Model(vote).
				Exec(ctx); err != nil {
				return err
			}

			var poll gtsmodel.Poll

			// Select current poll counts from DB,
			// taking minimal columns needed to
			// increment/decrement votes.
			if err := tx.NewSelect().
				Model(&poll).
				Column("options", "votes", "voters").
				Where("? = ?", bun.Ident("id"), vote.PollID).
				Scan(ctx); err != nil {
				return err
			}

			// Increment poll votes for choices.
			poll.IncrementVotes(vote.Choices)

			// Finally, update the poll entry.
			_, err := tx.NewUpdate().
				Model(&poll).
				Column("votes", "voters").
				Where("? = ?", bun.Ident("id"), vote.PollID).
				Exec(ctx)
			return err
		})
	})
}

func (p *pollDB) DeletePollVotes(ctx context.Context, pollID string) error {
	err := p.db.RunInTx(ctx, func(tx Tx) error {
		// Delete all votes in poll.
		res, err := tx.NewDelete().
			Table("poll_votes").
			Where("? = ?", bun.Ident("poll_id"), pollID).
			Exec(ctx)
		if err != nil {
			// irrecoverable
			return err
		}

		ra, err := res.RowsAffected()
		if err != nil {
			// irrecoverable
			return err
		}

		if ra == 0 {
			// No poll votes deleted,
			// nothing to update.
			return nil
		}

		// Select current poll counts from DB,
		// taking minimal columns needed to
		// increment/decrement votes.
		var poll gtsmodel.Poll
		switch err := tx.NewSelect().
			Model(&poll).
			Column("options", "votes", "voters").
			Where("? = ?", bun.Ident("id"), pollID).
			Scan(ctx); {

		case err == nil:
			// no issue.

		case errors.Is(err, db.ErrNoEntries):
			// no votes found,
			// return here.
			return nil

		default:
			// irrecoverable.
			return err
		}

		// Zero all counts.
		poll.ResetVotes()

		// Finally, update the poll entry.
		_, err = tx.NewUpdate().
			Model(&poll).
			Column("votes", "voters").
			Where("? = ?", bun.Ident("id"), pollID).
			Exec(ctx)
		return err
	})

	if err != nil {
		return err
	}

	// Invalidate poll vote and poll entry from caches.
	p.state.Caches.GTS.Poll().Invalidate("ID", pollID)
	p.state.Caches.GTS.PollVote().Invalidate("PollID", pollID)
	p.state.Caches.GTS.PollVoteIDs().Invalidate(pollID)

	return nil
}

func (p *pollDB) DeletePollVoteBy(ctx context.Context, pollID string, accountID string) error {
	err := p.db.RunInTx(ctx, func(tx Tx) error {
		// Slice should only ever be of length
		// 0 or 1; it's a slice of slices only
		// because we can't LIMIT deletes to 1.
		var choicesSl [][]int

		// Delete vote in poll by account,
		// returning the ID + choices of the vote.
		if err := tx.NewDelete().
			Table("poll_votes").
			Where("? = ?", bun.Ident("poll_id"), pollID).
			Where("? = ?", bun.Ident("account_id"), accountID).
			Returning("?", bun.Ident("choices")).
			Scan(ctx, &choicesSl); err != nil {
			// irrecoverable.
			return err
		}

		if len(choicesSl) != 1 {
			// No poll votes by this
			// acct on this poll.
			return nil
		}
		choices := choicesSl[0]

		// Select current poll counts from DB,
		// taking minimal columns needed to
		// increment/decrement votes.
		var poll gtsmodel.Poll
		switch err := tx.NewSelect().
			Model(&poll).
			Column("options", "votes", "voters").
			Where("? = ?", bun.Ident("id"), pollID).
			Scan(ctx); {

		case err == nil:
			// no issue.

		case errors.Is(err, db.ErrNoEntries):
			// no poll found,
			// return here.
			return nil

		default:
			// irrecoverable.
			return err
		}

		// Decrement votes for choices.
		poll.DecrementVotes(choices)

		// Finally, update the poll entry.
		_, err := tx.NewUpdate().
			Model(&poll).
			Column("votes", "voters").
			Where("? = ?", bun.Ident("id"), pollID).
			Exec(ctx)
		return err
	})

	if err != nil {
		return err
	}

	// Invalidate poll vote and poll entry from caches.
	p.state.Caches.GTS.Poll().Invalidate("ID", pollID)
	p.state.Caches.GTS.PollVote().Invalidate("PollID.AccountID", pollID, accountID)
	p.state.Caches.GTS.PollVoteIDs().Invalidate(pollID)

	return nil
}

func (p *pollDB) DeletePollVotesByAccountID(ctx context.Context, accountID string) error {
	var pollIDs []string

	// Select all polls this account
	// has registered a poll vote in.
	if err := p.db.NewSelect().
		Table("poll_votes").
		Column("poll_id").
		Where("? = ?", bun.Ident("account_id"), accountID).
		Scan(ctx, &pollIDs); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	for _, id := range pollIDs {
		// Delete all votes by this account in each of the polls,
		// this way ensures that all necessary caches are invalidated.
		if err := p.DeletePollVoteBy(ctx, id, accountID); err != nil {
			log.Errorf(ctx, "error deleting vote by %s in %s: %v", accountID, id, err)
		}
	}

	return nil
}

// newSelectPollVotes returns a new select query for all rows in the poll_votes table with poll_id = pollID.
func newSelectPollVotes(db *DB, pollID string) *bun.SelectQuery {
	return db.NewSelect().
		TableExpr("?", bun.Ident("poll_votes")).
		ColumnExpr("?", bun.Ident("id")).
		Where("? = ?", bun.Ident("poll_id"), pollID).
		OrderExpr("? DESC", bun.Ident("id"))
}
