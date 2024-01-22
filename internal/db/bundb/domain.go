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
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

type domainDB struct {
	db    *DB
	state *state.State
}

func (d *domainDB) CreateDomainAllow(ctx context.Context, allow *gtsmodel.DomainAllow) error {
	// Normalize the domain as punycode
	var err error
	allow.Domain, err = util.Punify(allow.Domain)
	if err != nil {
		return err
	}

	// Attempt to store domain allow in DB
	if _, err := d.db.NewInsert().
		Model(allow).
		Exec(ctx); err != nil {
		return err
	}

	// Clear the domain allow cache (for later reload)
	d.state.Caches.GTS.DomainAllow.Clear()

	return nil
}

func (d *domainDB) GetDomainAllow(ctx context.Context, domain string) (*gtsmodel.DomainAllow, error) {
	// Normalize the domain as punycode
	domain, err := util.Punify(domain)
	if err != nil {
		return nil, err
	}

	// Check for easy case, domain referencing *us*
	if domain == "" || domain == config.GetAccountDomain() ||
		domain == config.GetHost() {
		return nil, db.ErrNoEntries
	}

	var allow gtsmodel.DomainAllow

	// Look for allow matching domain in DB
	q := d.db.
		NewSelect().
		Model(&allow).
		Where("? = ?", bun.Ident("domain_allow.domain"), domain)
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return &allow, nil
}

func (d *domainDB) GetDomainAllows(ctx context.Context) ([]*gtsmodel.DomainAllow, error) {
	allows := []*gtsmodel.DomainAllow{}

	if err := d.db.
		NewSelect().
		Model(&allows).
		Scan(ctx); err != nil {
		return nil, err
	}

	return allows, nil
}

func (d *domainDB) GetDomainAllowByID(ctx context.Context, id string) (*gtsmodel.DomainAllow, error) {
	var allow gtsmodel.DomainAllow

	q := d.db.
		NewSelect().
		Model(&allow).
		Where("? = ?", bun.Ident("domain_allow.id"), id)
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return &allow, nil
}

func (d *domainDB) DeleteDomainAllow(ctx context.Context, domain string) error {
	// Normalize the domain as punycode
	domain, err := util.Punify(domain)
	if err != nil {
		return err
	}

	// Attempt to delete domain allow
	if _, err := d.db.NewDelete().
		Model((*gtsmodel.DomainAllow)(nil)).
		Where("? = ?", bun.Ident("domain_allow.domain"), domain).
		Exec(ctx); err != nil {
		return err
	}

	// Clear the domain allow cache (for later reload)
	d.state.Caches.GTS.DomainAllow.Clear()

	return nil
}

func (d *domainDB) CreateDomainBlock(ctx context.Context, block *gtsmodel.DomainBlock) error {
	// Normalize the domain as punycode
	var err error
	block.Domain, err = util.Punify(block.Domain)
	if err != nil {
		return err
	}

	// Attempt to store domain block in DB
	if _, err := d.db.NewInsert().
		Model(block).
		Exec(ctx); err != nil {
		return err
	}

	// Clear the domain block cache (for later reload)
	d.state.Caches.GTS.DomainBlock.Clear()

	return nil
}

func (d *domainDB) GetDomainBlock(ctx context.Context, domain string) (*gtsmodel.DomainBlock, error) {
	// Normalize the domain as punycode
	domain, err := util.Punify(domain)
	if err != nil {
		return nil, err
	}

	// Check for easy case, domain referencing *us*
	if domain == "" || domain == config.GetAccountDomain() ||
		domain == config.GetHost() {
		return nil, db.ErrNoEntries
	}

	var block gtsmodel.DomainBlock

	// Look for block matching domain in DB
	q := d.db.
		NewSelect().
		Model(&block).
		Where("? = ?", bun.Ident("domain_block.domain"), domain)
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return &block, nil
}

func (d *domainDB) GetDomainBlocks(ctx context.Context) ([]*gtsmodel.DomainBlock, error) {
	blocks := []*gtsmodel.DomainBlock{}

	if err := d.db.
		NewSelect().
		Model(&blocks).
		Scan(ctx); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (d *domainDB) GetDomainBlockByID(ctx context.Context, id string) (*gtsmodel.DomainBlock, error) {
	var block gtsmodel.DomainBlock

	q := d.db.
		NewSelect().
		Model(&block).
		Where("? = ?", bun.Ident("domain_block.id"), id)
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return &block, nil
}

func (d *domainDB) DeleteDomainBlock(ctx context.Context, domain string) error {
	// Normalize the domain as punycode
	domain, err := util.Punify(domain)
	if err != nil {
		return err
	}

	// Attempt to delete domain block
	if _, err := d.db.NewDelete().
		Model((*gtsmodel.DomainBlock)(nil)).
		Where("? = ?", bun.Ident("domain_block.domain"), domain).
		Exec(ctx); err != nil {
		return err
	}

	// Clear the domain block cache (for later reload)
	d.state.Caches.GTS.DomainBlock.Clear()

	return nil
}

func (d *domainDB) IsDomainBlocked(ctx context.Context, domain string) (bool, error) {
	// Normalize the domain as punycode
	domain, err := util.Punify(domain)
	if err != nil {
		return false, err
	}

	// Domain referencing *us* cannot be blocked.
	if domain == "" || domain == config.GetAccountDomain() ||
		domain == config.GetHost() {
		return false, nil
	}

	// Check the cache for an explicit domain allow (hydrating the cache with callback if necessary).
	explicitAllow, err := d.state.Caches.GTS.DomainAllow.Matches(domain, func() ([]string, error) {
		var domains []string

		// Scan list of all explicitly allowed domains from DB
		q := d.db.NewSelect().
			Table("domain_allows").
			Column("domain")
		if err := q.Scan(ctx, &domains); err != nil {
			return nil, err
		}

		return domains, nil
	})
	if err != nil {
		return false, err
	}

	// Check the cache for a domain block (hydrating the cache with callback if necessary)
	explicitBlock, err := d.state.Caches.GTS.DomainBlock.Matches(domain, func() ([]string, error) {
		var domains []string

		// Scan list of all blocked domains from DB
		q := d.db.NewSelect().
			Table("domain_blocks").
			Column("domain")
		if err := q.Scan(ctx, &domains); err != nil {
			return nil, err
		}

		return domains, nil
	})
	if err != nil {
		return false, err
	}

	// Calculate if blocked
	// based on federation mode.
	switch mode := config.GetInstanceFederationMode(); mode {

	case config.InstanceFederationModeBlocklist:
		// Blocklist/default mode: explicit allow
		// takes precedence over explicit block.
		//
		// Domains that have neither block
		// or allow entries are allowed.
		return !(explicitAllow || !explicitBlock), nil

	case config.InstanceFederationModeAllowlist:
		// Allowlist mode: explicit block takes
		// precedence over explicit allow.
		//
		// Domains that have neither block
		// or allow entries are blocked.
		return (explicitBlock || !explicitAllow), nil

	default:
		// This should never happen but account
		// for it anyway to make the code tidier.
		return false, gtserror.Newf("unrecognized federation mode: %s", mode)
	}
}

func (d *domainDB) AreDomainsBlocked(ctx context.Context, domains []string) (bool, error) {
	for _, domain := range domains {
		if blocked, err := d.IsDomainBlocked(ctx, domain); err != nil {
			return false, err
		} else if blocked {
			return blocked, nil
		}
	}
	return false, nil
}

func (d *domainDB) IsURIBlocked(ctx context.Context, uri *url.URL) (bool, error) {
	return d.IsDomainBlocked(ctx, uri.Hostname())
}

func (d *domainDB) AreURIsBlocked(ctx context.Context, uris []*url.URL) (bool, error) {
	for _, uri := range uris {
		if blocked, err := d.IsDomainBlocked(ctx, uri.Hostname()); err != nil {
			return false, err
		} else if blocked {
			return blocked, nil
		}
	}
	return false, nil
}
