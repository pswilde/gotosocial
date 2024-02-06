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

package typeutils_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type InternalToASTestSuite struct {
	TypeUtilsTestSuite
}

func (suite *InternalToASTestSuite) TestAccountToAS() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"] // take zork for this test

	asPerson, err := suite.typeconverter.AccountToAS(context.Background(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(asPerson)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	// trim off everything up to 'discoverable';
	// this is necessary because the order of multiple 'context' entries is not determinate
	trimmed := strings.Split(string(bytes), "\"discoverable\"")[1]

	suite.Equal(`: true,
  "featured": "http://localhost:8080/users/the_mighty_zork/collections/featured",
  "followers": "http://localhost:8080/users/the_mighty_zork/followers",
  "following": "http://localhost:8080/users/the_mighty_zork/following",
  "icon": {
    "mediaType": "image/jpeg",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg"
  },
  "id": "http://localhost:8080/users/the_mighty_zork",
  "image": {
    "mediaType": "image/jpeg",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg"
  },
  "inbox": "http://localhost:8080/users/the_mighty_zork/inbox",
  "manuallyApprovesFollowers": false,
  "name": "original zork (he/they)",
  "outbox": "http://localhost:8080/users/the_mighty_zork/outbox",
  "preferredUsername": "the_mighty_zork",
  "publicKey": {
    "id": "http://localhost:8080/users/the_mighty_zork/main-key",
    "owner": "http://localhost:8080/users/the_mighty_zork",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwXTcOAvM1Jiw5Ffpk0qn\nr0cwbNvFe/5zQ+Tp7tumK/ZnT37o7X0FUEXrxNi+dkhmeJ0gsaiN+JQGNUewvpSk\nPIAXKvi908aSfCGjs7bGlJCJCuDuL5d6m7hZnP9rt9fJc70GElPpG0jc9fXwlz7T\nlsPb2ecatmG05Y4jPwdC+oN4MNCv9yQzEvCVMzl76EJaM602kIHC1CISn0rDFmYd\n9rSN7XPlNJw1F6PbpJ/BWQ+pXHKw3OEwNTETAUNYiVGnZU+B7a7bZC9f6/aPbJuV\nt8Qmg+UnDvW1Y8gmfHnxaWG2f5TDBvCHmcYtucIZPLQD4trAozC4ryqlmCWQNKbt\n0wIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "summary": "\u003cp\u003ehey yo this is my profile!\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://localhost:8080/@the_mighty_zork"
}`, trimmed)
}

func (suite *InternalToASTestSuite) TestAccountToASWithFields() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_2"]

	asPerson, err := suite.typeconverter.AccountToAS(context.Background(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(asPerson)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	// trim off everything up to 'attachment';
	// this is necessary because the order of multiple 'context' entries is not determinate
	trimmed := strings.Split(string(bytes), "\"attachment\"")[1]

	fmt.Printf("\n\n\n%s\n\n\n", string(bytes))

	suite.Equal(`: [
    {
      "name": "should you follow me?",
      "type": "PropertyValue",
      "value": "maybe!"
    },
    {
      "name": "age",
      "type": "PropertyValue",
      "value": "120"
    }
  ],
  "discoverable": false,
  "featured": "http://localhost:8080/users/1happyturtle/collections/featured",
  "followers": "http://localhost:8080/users/1happyturtle/followers",
  "following": "http://localhost:8080/users/1happyturtle/following",
  "id": "http://localhost:8080/users/1happyturtle",
  "inbox": "http://localhost:8080/users/1happyturtle/inbox",
  "manuallyApprovesFollowers": true,
  "name": "happy little turtle :3",
  "outbox": "http://localhost:8080/users/1happyturtle/outbox",
  "preferredUsername": "1happyturtle",
  "publicKey": {
    "id": "http://localhost:8080/users/1happyturtle#main-key",
    "owner": "http://localhost:8080/users/1happyturtle",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtTc6Jpg6LrRPhVQG4KLz\n2+YqEUUtZPd4YR+TKXuCnwEG9ZNGhgP046xa9h3EWzrZXaOhXvkUQgJuRqPrAcfN\nvc8jBHV2xrUeD8pu/MWKEabAsA/tgCv3nUC47HQ3/c12aHfYoPz3ufWsGGnrkhci\nv8PaveJ3LohO5vjCn1yZ00v6osMJMViEZvZQaazyE9A8FwraIexXabDpoy7tkHRg\nA1fvSkg4FeSG1XMcIz2NN7xyUuFACD+XkuOk7UqzRd4cjPUPLxiDwIsTlcgGOd3E\nUFMWVlPxSGjY2hIKa3lEHytaYK9IMYdSuyCsJshd3/yYC9LqxZY2KdlKJ80VOVyh\nyQIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "summary": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://localhost:8080/@1happyturtle"
}`, trimmed)
}

func (suite *InternalToASTestSuite) TestAccountToASAliasedAndMoved() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"] // take zork for this test

	ctx := context.Background()

	// Suppose zork has moved account to turtle.
	testAccount.AlsoKnownAsURIs = []string{"http://localhost:8080/users/1happyturtle"}
	testAccount.MovedToURI = "http://localhost:8080/users/1happyturtle"
	if err := suite.state.DB.UpdateAccount(ctx,
		testAccount,
		"also_known_as_uris",
		"moved_to_uri",
	); err != nil {
		suite.FailNow(err.Error())
	}

	asPerson, err := suite.typeconverter.AccountToAS(context.Background(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(asPerson)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	// trim off everything up to 'alsoKnownAs';
	// this is necessary because the order of multiple 'context' entries is not determinate
	trimmed := strings.Split(string(bytes), "\"alsoKnownAs\"")[1]

	suite.Equal(`: [
    "http://localhost:8080/users/1happyturtle"
  ],
  "discoverable": true,
  "featured": "http://localhost:8080/users/the_mighty_zork/collections/featured",
  "followers": "http://localhost:8080/users/the_mighty_zork/followers",
  "following": "http://localhost:8080/users/the_mighty_zork/following",
  "icon": {
    "mediaType": "image/jpeg",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg"
  },
  "id": "http://localhost:8080/users/the_mighty_zork",
  "image": {
    "mediaType": "image/jpeg",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg"
  },
  "inbox": "http://localhost:8080/users/the_mighty_zork/inbox",
  "manuallyApprovesFollowers": false,
  "movedTo": "http://localhost:8080/users/1happyturtle",
  "name": "original zork (he/they)",
  "outbox": "http://localhost:8080/users/the_mighty_zork/outbox",
  "preferredUsername": "the_mighty_zork",
  "publicKey": {
    "id": "http://localhost:8080/users/the_mighty_zork/main-key",
    "owner": "http://localhost:8080/users/the_mighty_zork",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwXTcOAvM1Jiw5Ffpk0qn\nr0cwbNvFe/5zQ+Tp7tumK/ZnT37o7X0FUEXrxNi+dkhmeJ0gsaiN+JQGNUewvpSk\nPIAXKvi908aSfCGjs7bGlJCJCuDuL5d6m7hZnP9rt9fJc70GElPpG0jc9fXwlz7T\nlsPb2ecatmG05Y4jPwdC+oN4MNCv9yQzEvCVMzl76EJaM602kIHC1CISn0rDFmYd\n9rSN7XPlNJw1F6PbpJ/BWQ+pXHKw3OEwNTETAUNYiVGnZU+B7a7bZC9f6/aPbJuV\nt8Qmg+UnDvW1Y8gmfHnxaWG2f5TDBvCHmcYtucIZPLQD4trAozC4ryqlmCWQNKbt\n0wIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "summary": "\u003cp\u003ehey yo this is my profile!\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://localhost:8080/@the_mighty_zork"
}`, trimmed)
}

func (suite *InternalToASTestSuite) TestAccountToASWithOneField() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_2"]
	testAccount.Fields = testAccount.Fields[0:1] // Take only one field.

	asPerson, err := suite.typeconverter.AccountToAS(context.Background(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(asPerson)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	// trim off everything up to 'attachment';
	// this is necessary because the order of multiple 'context' entries is not determinate
	trimmed := strings.Split(string(bytes), "\"attachment\"")[1]

	// Despite only one field being set, attachments should still be a slice/array.
	suite.Equal(`: [
    {
      "name": "should you follow me?",
      "type": "PropertyValue",
      "value": "maybe!"
    }
  ],
  "discoverable": false,
  "featured": "http://localhost:8080/users/1happyturtle/collections/featured",
  "followers": "http://localhost:8080/users/1happyturtle/followers",
  "following": "http://localhost:8080/users/1happyturtle/following",
  "id": "http://localhost:8080/users/1happyturtle",
  "inbox": "http://localhost:8080/users/1happyturtle/inbox",
  "manuallyApprovesFollowers": true,
  "name": "happy little turtle :3",
  "outbox": "http://localhost:8080/users/1happyturtle/outbox",
  "preferredUsername": "1happyturtle",
  "publicKey": {
    "id": "http://localhost:8080/users/1happyturtle#main-key",
    "owner": "http://localhost:8080/users/1happyturtle",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtTc6Jpg6LrRPhVQG4KLz\n2+YqEUUtZPd4YR+TKXuCnwEG9ZNGhgP046xa9h3EWzrZXaOhXvkUQgJuRqPrAcfN\nvc8jBHV2xrUeD8pu/MWKEabAsA/tgCv3nUC47HQ3/c12aHfYoPz3ufWsGGnrkhci\nv8PaveJ3LohO5vjCn1yZ00v6osMJMViEZvZQaazyE9A8FwraIexXabDpoy7tkHRg\nA1fvSkg4FeSG1XMcIz2NN7xyUuFACD+XkuOk7UqzRd4cjPUPLxiDwIsTlcgGOd3E\nUFMWVlPxSGjY2hIKa3lEHytaYK9IMYdSuyCsJshd3/yYC9LqxZY2KdlKJ80VOVyh\nyQIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "summary": "\u003cp\u003ei post about things that concern me\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://localhost:8080/@1happyturtle"
}`, trimmed)
}

func (suite *InternalToASTestSuite) TestAccountToASWithEmoji() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"] // take zork for this test
	testAccount.Emojis = []*gtsmodel.Emoji{suite.testEmojis["rainbow"]}

	asPerson, err := suite.typeconverter.AccountToAS(context.Background(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(asPerson)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	// trim off everything up to 'discoverable';
	// this is necessary because the order of multiple 'context' entries is not determinate
	trimmed := strings.Split(string(bytes), "\"discoverable\"")[1]

	suite.Equal(`: true,
  "featured": "http://localhost:8080/users/the_mighty_zork/collections/featured",
  "followers": "http://localhost:8080/users/the_mighty_zork/followers",
  "following": "http://localhost:8080/users/the_mighty_zork/following",
  "icon": {
    "mediaType": "image/jpeg",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg"
  },
  "id": "http://localhost:8080/users/the_mighty_zork",
  "image": {
    "mediaType": "image/jpeg",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg"
  },
  "inbox": "http://localhost:8080/users/the_mighty_zork/inbox",
  "manuallyApprovesFollowers": false,
  "name": "original zork (he/they)",
  "outbox": "http://localhost:8080/users/the_mighty_zork/outbox",
  "preferredUsername": "the_mighty_zork",
  "publicKey": {
    "id": "http://localhost:8080/users/the_mighty_zork/main-key",
    "owner": "http://localhost:8080/users/the_mighty_zork",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwXTcOAvM1Jiw5Ffpk0qn\nr0cwbNvFe/5zQ+Tp7tumK/ZnT37o7X0FUEXrxNi+dkhmeJ0gsaiN+JQGNUewvpSk\nPIAXKvi908aSfCGjs7bGlJCJCuDuL5d6m7hZnP9rt9fJc70GElPpG0jc9fXwlz7T\nlsPb2ecatmG05Y4jPwdC+oN4MNCv9yQzEvCVMzl76EJaM602kIHC1CISn0rDFmYd\n9rSN7XPlNJw1F6PbpJ/BWQ+pXHKw3OEwNTETAUNYiVGnZU+B7a7bZC9f6/aPbJuV\nt8Qmg+UnDvW1Y8gmfHnxaWG2f5TDBvCHmcYtucIZPLQD4trAozC4ryqlmCWQNKbt\n0wIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "summary": "\u003cp\u003ehey yo this is my profile!\u003c/p\u003e",
  "tag": {
    "icon": {
      "mediaType": "image/png",
      "type": "Image",
      "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png"
    },
    "id": "http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ",
    "name": ":rainbow:",
    "type": "Emoji",
    "updated": "2021-09-20T12:40:37+02:00"
  },
  "type": "Person",
  "url": "http://localhost:8080/@the_mighty_zork"
}`, trimmed)
}

func (suite *InternalToASTestSuite) TestAccountToASWithSharedInbox() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"] // take zork for this test
	sharedInbox := "http://localhost:8080/sharedInbox"
	testAccount.SharedInboxURI = &sharedInbox

	asPerson, err := suite.typeconverter.AccountToAS(context.Background(), testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(asPerson)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	// trim off everything up to 'discoverable';
	// this is necessary because the order of multiple 'context' entries is not determinate
	trimmed := strings.Split(string(bytes), "\"discoverable\"")[1]

	suite.Equal(`: true,
  "endpoints": {
    "sharedInbox": "http://localhost:8080/sharedInbox"
  },
  "featured": "http://localhost:8080/users/the_mighty_zork/collections/featured",
  "followers": "http://localhost:8080/users/the_mighty_zork/followers",
  "following": "http://localhost:8080/users/the_mighty_zork/following",
  "icon": {
    "mediaType": "image/jpeg",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpg"
  },
  "id": "http://localhost:8080/users/the_mighty_zork",
  "image": {
    "mediaType": "image/jpeg",
    "type": "Image",
    "url": "http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpg"
  },
  "inbox": "http://localhost:8080/users/the_mighty_zork/inbox",
  "manuallyApprovesFollowers": false,
  "name": "original zork (he/they)",
  "outbox": "http://localhost:8080/users/the_mighty_zork/outbox",
  "preferredUsername": "the_mighty_zork",
  "publicKey": {
    "id": "http://localhost:8080/users/the_mighty_zork/main-key",
    "owner": "http://localhost:8080/users/the_mighty_zork",
    "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwXTcOAvM1Jiw5Ffpk0qn\nr0cwbNvFe/5zQ+Tp7tumK/ZnT37o7X0FUEXrxNi+dkhmeJ0gsaiN+JQGNUewvpSk\nPIAXKvi908aSfCGjs7bGlJCJCuDuL5d6m7hZnP9rt9fJc70GElPpG0jc9fXwlz7T\nlsPb2ecatmG05Y4jPwdC+oN4MNCv9yQzEvCVMzl76EJaM602kIHC1CISn0rDFmYd\n9rSN7XPlNJw1F6PbpJ/BWQ+pXHKw3OEwNTETAUNYiVGnZU+B7a7bZC9f6/aPbJuV\nt8Qmg+UnDvW1Y8gmfHnxaWG2f5TDBvCHmcYtucIZPLQD4trAozC4ryqlmCWQNKbt\n0wIDAQAB\n-----END PUBLIC KEY-----\n"
  },
  "summary": "\u003cp\u003ehey yo this is my profile!\u003c/p\u003e",
  "tag": [],
  "type": "Person",
  "url": "http://localhost:8080/@the_mighty_zork"
}`, trimmed)
}

func (suite *InternalToASTestSuite) TestOutboxToASCollection() {
	testAccount := suite.testAccounts["admin_account"]
	ctx := context.Background()

	collection, err := suite.typeconverter.OutboxToASCollection(ctx, testAccount.OutboxURI)
	suite.NoError(err)

	ser, err := ap.Serialize(collection)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "first": "http://localhost:8080/users/admin/outbox?page=true",
  "id": "http://localhost:8080/users/admin/outbox",
  "type": "OrderedCollection"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToAS() {
	testStatus := suite.testStatuses["local_account_1_status_1"]
	ctx := context.Background()

	asStatus, err := suite.typeconverter.StatusToAS(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asStatus)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "attachment": [],
  "attributedTo": "http://localhost:8080/users/the_mighty_zork",
  "cc": "http://localhost:8080/users/the_mighty_zork/followers",
  "content": "hello everyone!",
  "contentMap": {
    "en": "hello everyone!"
  },
  "id": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "published": "2021-10-20T12:40:37+02:00",
  "replies": {
    "first": {
      "id": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies?page=true",
      "next": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies?only_other_accounts=false\u0026page=true",
      "partOf": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies",
      "type": "CollectionPage"
    },
    "id": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies",
    "type": "Collection"
  },
  "sensitive": true,
  "summary": "introduction post",
  "tag": [],
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusWithTagsToASWithIDs() {
	// use the status with just IDs of attachments and emojis pinned on it
	testStatus := suite.testStatuses["admin_account_status_1"]
	ctx := context.Background()

	asStatus, err := suite.typeconverter.StatusToAS(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asStatus)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	// we can't be sure in what order the two context entries --
	// http://joinmastodon.org/ns, https://www.w3.org/ns/activitystreams --
	// will appear, so trim them out of the string for consistency
	trimmed := strings.SplitAfter(string(bytes), `"attachment":`)[1]
	suite.Equal(` [
    {
      "blurhash": "LNJRdVM{00Rj%Mayt7j[4nWBofRj",
      "mediaType": "image/jpeg",
      "name": "Black and white image of some 50's style text saying: Welcome On Board",
      "type": "Document",
      "url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg"
    }
  ],
  "attributedTo": "http://localhost:8080/users/admin",
  "cc": "http://localhost:8080/users/admin/followers",
  "content": "hello world! #welcome ! first post on the instance :rainbow: !",
  "contentMap": {
    "en": "hello world! #welcome ! first post on the instance :rainbow: !"
  },
  "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
  "published": "2021-10-20T11:36:45Z",
  "replies": {
    "first": {
      "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies?page=true",
      "next": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies?only_other_accounts=false\u0026page=true",
      "partOf": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies",
      "type": "CollectionPage"
    },
    "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies",
    "type": "Collection"
  },
  "sensitive": false,
  "summary": "",
  "tag": [
    {
      "icon": {
        "mediaType": "image/png",
        "type": "Image",
        "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png"
      },
      "id": "http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ",
      "name": ":rainbow:",
      "type": "Emoji",
      "updated": "2021-09-20T10:40:37Z"
    },
    {
      "href": "http://localhost:8080/tags/welcome",
      "name": "#welcome",
      "type": "Hashtag"
    }
  ],
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R"
}`, trimmed)
}

func (suite *InternalToASTestSuite) TestStatusWithTagsToASFromDB() {
	ctx := context.Background()
	// get the entire status with all tags
	testStatus, err := suite.db.GetStatusByID(ctx, suite.testStatuses["admin_account_status_1"].ID)
	suite.NoError(err)

	asStatus, err := suite.typeconverter.StatusToAS(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asStatus)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	// we can't be sure in what order the two context entries --
	// http://joinmastodon.org/ns, https://www.w3.org/ns/activitystreams --
	// will appear, so trim them out of the string for consistency
	trimmed := strings.SplitAfter(string(bytes), `"attachment":`)[1]
	suite.Equal(` [
    {
      "blurhash": "LNJRdVM{00Rj%Mayt7j[4nWBofRj",
      "mediaType": "image/jpeg",
      "name": "Black and white image of some 50's style text saying: Welcome On Board",
      "type": "Document",
      "url": "http://localhost:8080/fileserver/01F8MH17FWEB39HZJ76B6VXSKF/attachment/original/01F8MH6NEM8D7527KZAECTCR76.jpg"
    }
  ],
  "attributedTo": "http://localhost:8080/users/admin",
  "cc": "http://localhost:8080/users/admin/followers",
  "content": "hello world! #welcome ! first post on the instance :rainbow: !",
  "contentMap": {
    "en": "hello world! #welcome ! first post on the instance :rainbow: !"
  },
  "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
  "published": "2021-10-20T11:36:45Z",
  "replies": {
    "first": {
      "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies?page=true",
      "next": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies?only_other_accounts=false\u0026page=true",
      "partOf": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies",
      "type": "CollectionPage"
    },
    "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies",
    "type": "Collection"
  },
  "sensitive": false,
  "summary": "",
  "tag": [
    {
      "icon": {
        "mediaType": "image/png",
        "type": "Image",
        "url": "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png"
      },
      "id": "http://localhost:8080/emoji/01F8MH9H8E4VG3KDYJR9EGPXCQ",
      "name": ":rainbow:",
      "type": "Emoji",
      "updated": "2021-09-20T10:40:37Z"
    },
    {
      "href": "http://localhost:8080/tags/welcome",
      "name": "#welcome",
      "type": "Hashtag"
    }
  ],
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R"
}`, trimmed)
}

func (suite *InternalToASTestSuite) TestStatusToASWithMentions() {
	testStatusID := suite.testStatuses["admin_account_status_3"].ID
	ctx := context.Background()

	testStatus, err := suite.db.GetStatusByID(ctx, testStatusID)
	suite.NoError(err)

	asStatus, err := suite.typeconverter.StatusToAS(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asStatus)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "attachment": [],
  "attributedTo": "http://localhost:8080/users/admin",
  "cc": [
    "http://localhost:8080/users/admin/followers",
    "http://localhost:8080/users/the_mighty_zork"
  ],
  "content": "hi @the_mighty_zork welcome to the instance!",
  "contentMap": {
    "en": "hi @the_mighty_zork welcome to the instance!"
  },
  "id": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
  "inReplyTo": "http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY",
  "published": "2021-11-20T13:32:16Z",
  "replies": {
    "first": {
      "id": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0/replies?page=true",
      "next": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0/replies?only_other_accounts=false\u0026page=true",
      "partOf": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0/replies",
      "type": "CollectionPage"
    },
    "id": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0/replies",
    "type": "Collection"
  },
  "sensitive": false,
  "summary": "",
  "tag": {
    "href": "http://localhost:8080/users/the_mighty_zork",
    "name": "@the_mighty_zork@localhost:8080",
    "type": "Mention"
  },
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Note",
  "url": "http://localhost:8080/@admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToASDeletePublicReply() {
	testStatus := suite.testStatuses["admin_account_status_3"]
	ctx := context.Background()

	asDelete, err := suite.typeconverter.StatusToASDelete(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asDelete)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/admin",
  "cc": [
    "http://localhost:8080/users/admin/followers",
    "http://localhost:8080/users/the_mighty_zork"
  ],
  "object": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Delete"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToASDeletePublicReplyOriginalDeleted() {
	testStatus := suite.testStatuses["admin_account_status_3"]
	ctx := context.Background()

	// Delete the status this replies to.
	if err := suite.db.DeleteStatusByID(ctx, testStatus.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Delete the mention the reply created.
	mention := suite.testMentions["admin_account_mention_zork"]
	if err := suite.db.DeleteByID(ctx, mention.ID, mention); err != nil {
		suite.FailNow(err.Error())
	}

	// The delete should still be created OK.
	asDelete, err := suite.typeconverter.StatusToASDelete(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asDelete)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/admin",
  "cc": [
    "http://localhost:8080/users/admin/followers",
    "http://localhost:8080/users/the_mighty_zork"
  ],
  "object": "http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Delete"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToASDeletePublic() {
	testStatus := suite.testStatuses["admin_account_status_1"]
	ctx := context.Background()

	asDelete, err := suite.typeconverter.StatusToASDelete(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asDelete)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/admin",
  "cc": "http://localhost:8080/users/admin/followers",
  "object": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
  "to": "https://www.w3.org/ns/activitystreams#Public",
  "type": "Delete"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToASDeleteDirectMessage() {
	testStatus := suite.testStatuses["local_account_2_status_6"]
	ctx := context.Background()

	asDelete, err := suite.typeconverter.StatusToASDelete(ctx, testStatus)
	suite.NoError(err)

	ser, err := ap.Serialize(asDelete)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/1happyturtle",
  "cc": [],
  "object": "http://localhost:8080/users/1happyturtle/statuses/01FN3VJGFH10KR7S2PB0GFJZYG",
  "to": "http://localhost:8080/users/the_mighty_zork",
  "type": "Delete"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusesToASOutboxPage() {
	testAccount := suite.testAccounts["admin_account"]
	ctx := context.Background()

	// get public statuses from testaccount
	statuses, err := suite.db.GetAccountStatuses(ctx, testAccount.ID, 30, true, true, "", "", false, true)
	suite.NoError(err)

	page, err := suite.typeconverter.StatusesToASOutboxPage(ctx, testAccount.OutboxURI, "", "", statuses)
	suite.NoError(err)

	ser, err := ap.Serialize(page)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://localhost:8080/users/admin/outbox?page=true",
  "next": "http://localhost:8080/users/admin/outbox?page=true\u0026max_id=01F8MH75CBF9JFX4ZAD54N0W0R",
  "orderedItems": [
    {
      "actor": "http://localhost:8080/users/admin",
      "cc": "http://localhost:8080/users/admin/followers",
      "id": "http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37/activity#Create",
      "object": "http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
      "published": "2021-10-20T12:36:45Z",
      "to": "https://www.w3.org/ns/activitystreams#Public",
      "type": "Create"
    },
    {
      "actor": "http://localhost:8080/users/admin",
      "cc": "http://localhost:8080/users/admin/followers",
      "id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/activity#Create",
      "object": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
      "published": "2021-10-20T11:36:45Z",
      "to": "https://www.w3.org/ns/activitystreams#Public",
      "type": "Create"
    }
  ],
  "partOf": "http://localhost:8080/users/admin/outbox",
  "prev": "http://localhost:8080/users/admin/outbox?page=true\u0026min_id=01F8MHAAY43M6RJ473VQFCVH37",
  "type": "OrderedCollectionPage"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestSelfBoostFollowersOnlyToAS() {
	ctx := context.Background()

	testStatus := suite.testStatuses["local_account_1_status_5"]
	testAccount := suite.testAccounts["local_account_1"]

	boostWrapperStatus, err := suite.typeconverter.StatusToBoost(ctx, testStatus, testAccount, "")
	suite.NoError(err)
	suite.NotNil(boostWrapperStatus)

	// Set some fields to predictable values for the test.
	boostWrapperStatus.ID = "01G74JJ1KS331G2JXHRMZCE0ER"
	boostWrapperStatus.URI = "http://localhost:8080/users/the_mighty_zork/statuses/01G74JJ1KS331G2JXHRMZCE0ER"
	boostWrapperStatus.CreatedAt = testrig.TimeMustParse("2022-06-09T13:12:00Z")

	asBoost, err := suite.typeconverter.BoostToAS(ctx, boostWrapperStatus, testAccount, testAccount)
	suite.NoError(err)

	ser, err := ap.Serialize(asBoost)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/the_mighty_zork",
  "cc": "http://localhost:8080/users/the_mighty_zork",
  "id": "http://localhost:8080/users/the_mighty_zork/statuses/01G74JJ1KS331G2JXHRMZCE0ER",
  "object": "http://localhost:8080/users/the_mighty_zork/statuses/01FCTA44PW9H1TB328S9AQXKDS",
  "published": "2022-06-09T13:12:00Z",
  "to": "http://localhost:8080/users/the_mighty_zork/followers",
  "type": "Announce"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestReportToAS() {
	ctx := context.Background()

	testReport := suite.testReports["local_account_2_report_remote_account_1"]
	account := suite.testAccounts["local_account_2"]
	targetAccount := suite.testAccounts["remote_account_1"]
	statuses := []*gtsmodel.Status{suite.testStatuses["remote_account_1_status_1"]}

	testReport.Account = account
	testReport.TargetAccount = targetAccount
	testReport.Statuses = statuses

	flag, err := suite.typeconverter.ReportToASFlag(ctx, testReport)
	suite.NoError(err)

	ser, err := ap.Serialize(flag)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/localhost:8080",
  "content": "dark souls sucks, please yeet this nerd",
  "id": "http://localhost:8080/reports/01GP3AWY4CRDVRNZKW0TEAMB5R",
  "object": [
    "http://fossbros-anonymous.io/users/foss_satan",
    "http://fossbros-anonymous.io/users/foss_satan/statuses/01FVW7JHQFSFK166WWKR8CBA6M"
  ],
  "type": "Flag"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestPinnedStatusesToASSomeItems() {
	ctx := context.Background()

	testAccount := suite.testAccounts["admin_account"]
	statuses, err := suite.db.GetAccountPinnedStatuses(ctx, testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	collection, err := suite.typeconverter.StatusesToASFeaturedCollection(ctx, testAccount.FeaturedCollectionURI, statuses)
	if err != nil {
		suite.FailNow(err.Error())
	}

	ser, err := ap.Serialize(collection)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://localhost:8080/users/admin/collections/featured",
  "orderedItems": [
    "http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
    "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R"
  ],
  "totalItems": 2,
  "type": "OrderedCollection"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestPinnedStatusesToASNoItems() {
	ctx := context.Background()

	testAccount := suite.testAccounts["local_account_1"]
	statuses, err := suite.db.GetAccountPinnedStatuses(ctx, testAccount.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		suite.FailNow(err.Error())
	}

	collection, err := suite.typeconverter.StatusesToASFeaturedCollection(ctx, testAccount.FeaturedCollectionURI, statuses)
	if err != nil {
		suite.FailNow(err.Error())
	}

	ser, err := ap.Serialize(collection)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://localhost:8080/users/the_mighty_zork/collections/featured",
  "orderedItems": [],
  "totalItems": 0,
  "type": "OrderedCollection"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestPinnedStatusesToASOneItem() {
	ctx := context.Background()

	testAccount := suite.testAccounts["local_account_2"]
	statuses, err := suite.db.GetAccountPinnedStatuses(ctx, testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	collection, err := suite.typeconverter.StatusesToASFeaturedCollection(ctx, testAccount.FeaturedCollectionURI, statuses)
	if err != nil {
		suite.FailNow(err.Error())
	}

	ser, err := ap.Serialize(collection)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(ser, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "http://localhost:8080/users/1happyturtle/collections/featured",
  "orderedItems": [
    "http://localhost:8080/users/1happyturtle/statuses/01G20ZM733MGN8J344T4ZDDFY1"
  ],
  "totalItems": 1,
  "type": "OrderedCollection"
}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestPollVoteToASCreate() {
	vote := suite.testPollVotes["remote_account_1_status_2_poll_vote_local_account_1"]

	create, err := suite.typeconverter.PollVoteToASCreate(context.Background(), vote)
	if err != nil {
		suite.FailNow(err.Error())
	}

	createI, err := ap.Serialize(create)
	suite.NoError(err)

	bytes, err := json.MarshalIndent(createI, "", "  ")
	suite.NoError(err)

	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/the_mighty_zork",
  "id": "http://localhost:8080/users/the_mighty_zork/activity#vote/http://fossbros-anonymous.io/users/foss_satan/statuses/01HEN2QRFA8H3C6QPN7RD4KSR6",
  "object": [
    {
      "attributedTo": "http://localhost:8080/users/the_mighty_zork",
      "id": "http://localhost:8080/users/the_mighty_zork#01HEN2R65468ZG657C4ZPHJ4EX/votes/1",
      "inReplyTo": "http://fossbros-anonymous.io/users/foss_satan/statuses/01HEN2QRFA8H3C6QPN7RD4KSR6",
      "name": "tissues",
      "to": "http://fossbros-anonymous.io/users/foss_satan",
      "type": "Note"
    },
    {
      "attributedTo": "http://localhost:8080/users/the_mighty_zork",
      "id": "http://localhost:8080/users/the_mighty_zork#01HEN2R65468ZG657C4ZPHJ4EX/votes/2",
      "inReplyTo": "http://fossbros-anonymous.io/users/foss_satan/statuses/01HEN2QRFA8H3C6QPN7RD4KSR6",
      "name": "financial times",
      "to": "http://fossbros-anonymous.io/users/foss_satan",
      "type": "Note"
    }
  ],
  "published": "2021-09-11T11:45:37+02:00",
  "to": "http://fossbros-anonymous.io/users/foss_satan",
  "type": "Create"
}`, string(bytes))
}

func TestInternalToASTestSuite(t *testing.T) {
	suite.Run(t, new(InternalToASTestSuite))
}
