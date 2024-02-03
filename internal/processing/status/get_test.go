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

package status_test

import (
	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/processing/status"
	"testing"
)

type topoSortTestSuite struct {
	suite.Suite
}

func statusIDs(apiStatuses []*apimodel.Status) []string {
	ids := make([]string, 0, len(apiStatuses))
	for _, apiStatus := range apiStatuses {
		ids = append(ids, apiStatus.ID)
	}
	return ids
}

func (suite *topoSortTestSuite) TestBranched() {
	// https://commons.wikimedia.org/wiki/File:Sorted_binary_tree_ALL_RGB.svg
	f := &apimodel.Status{ID: "F"}
	b := &apimodel.Status{ID: "B", InReplyToID: &f.ID}
	a := &apimodel.Status{ID: "A", InReplyToID: &b.ID}
	d := &apimodel.Status{ID: "D", InReplyToID: &b.ID}
	c := &apimodel.Status{ID: "C", InReplyToID: &d.ID}
	e := &apimodel.Status{ID: "E", InReplyToID: &d.ID}
	g := &apimodel.Status{ID: "G", InReplyToID: &f.ID}
	i := &apimodel.Status{ID: "I", InReplyToID: &g.ID}
	h := &apimodel.Status{ID: "H", InReplyToID: &i.ID}

	expected := statusIDs([]*apimodel.Status{f, b, a, d, c, e, g, i, h})
	list := []*apimodel.Status{a, b, c, d, e, f, g, h, i}
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.Equal(expected, actual)
}

func (suite *topoSortTestSuite) TestBranchedWithSelfReplyChain() {
	targetAccount := &apimodel.Account{ID: "1"}
	otherAccount := &apimodel.Account{ID: "2"}

	f := &apimodel.Status{
		ID:      "F",
		Account: targetAccount,
	}
	b := &apimodel.Status{
		ID:                 "B",
		Account:            targetAccount,
		InReplyToID:        &f.ID,
		InReplyToAccountID: &f.Account.ID,
	}
	a := &apimodel.Status{
		ID:                 "A",
		Account:            otherAccount,
		InReplyToID:        &b.ID,
		InReplyToAccountID: &b.Account.ID,
	}
	d := &apimodel.Status{
		ID:                 "D",
		Account:            targetAccount,
		InReplyToID:        &b.ID,
		InReplyToAccountID: &b.Account.ID,
	}
	c := &apimodel.Status{
		ID:                 "C",
		Account:            otherAccount,
		InReplyToID:        &d.ID,
		InReplyToAccountID: &d.Account.ID,
	}
	e := &apimodel.Status{
		ID:                 "E",
		Account:            targetAccount,
		InReplyToID:        &d.ID,
		InReplyToAccountID: &d.Account.ID,
	}
	g := &apimodel.Status{
		ID:                 "G",
		Account:            otherAccount,
		InReplyToID:        &f.ID,
		InReplyToAccountID: &f.Account.ID,
	}
	i := &apimodel.Status{
		ID:                 "I",
		Account:            targetAccount,
		InReplyToID:        &g.ID,
		InReplyToAccountID: &g.Account.ID,
	}
	h := &apimodel.Status{
		ID:                 "H",
		Account:            otherAccount,
		InReplyToID:        &i.ID,
		InReplyToAccountID: &i.Account.ID,
	}

	expected := statusIDs([]*apimodel.Status{f, b, d, e, c, a, g, i, h})
	list := []*apimodel.Status{a, b, c, d, e, f, g, h, i}
	status.TopoSort(list, targetAccount.ID)
	actual := statusIDs(list)

	suite.Equal(expected, actual)
}

func (suite *topoSortTestSuite) TestDisconnected() {
	f := &apimodel.Status{ID: "F"}
	b := &apimodel.Status{ID: "B", InReplyToID: &f.ID}
	dID := "D"
	e := &apimodel.Status{ID: "E", InReplyToID: &dID}

	expected := statusIDs([]*apimodel.Status{e, f, b})
	list := []*apimodel.Status{b, e, f}
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.Equal(expected, actual)
}

func (suite *topoSortTestSuite) TestTrivialCycle() {
	xID := "X"
	x := &apimodel.Status{ID: xID, InReplyToID: &xID}

	expected := statusIDs([]*apimodel.Status{x})
	list := []*apimodel.Status{x}
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.ElementsMatch(expected, actual)
}

func (suite *topoSortTestSuite) TestCycle() {
	yID := "Y"
	x := &apimodel.Status{ID: "X", InReplyToID: &yID}
	y := &apimodel.Status{ID: yID, InReplyToID: &x.ID}

	expected := statusIDs([]*apimodel.Status{x, y})
	list := []*apimodel.Status{x, y}
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.ElementsMatch(expected, actual)
}

func (suite *topoSortTestSuite) TestMixedCycle() {
	yID := "Y"
	x := &apimodel.Status{ID: "X", InReplyToID: &yID}
	y := &apimodel.Status{ID: yID, InReplyToID: &x.ID}
	z := &apimodel.Status{ID: "Z"}

	expected := statusIDs([]*apimodel.Status{x, y, z})
	list := []*apimodel.Status{x, y, z}
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.ElementsMatch(expected, actual)
}

func (suite *topoSortTestSuite) TestEmpty() {
	expected := statusIDs([]*apimodel.Status{})
	list := []*apimodel.Status{}
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.Equal(expected, actual)
}

func (suite *topoSortTestSuite) TestNil() {
	expected := statusIDs(nil)
	var list []*apimodel.Status
	status.TopoSort(list, "")
	actual := statusIDs(list)

	suite.Equal(expected, actual)
}

func TestTopoSortTestSuite(t *testing.T) {
	suite.Run(t, &topoSortTestSuite{})
}
