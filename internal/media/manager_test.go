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

package media_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"testing"
	"time"

	"codeberg.org/gruf/go-store/v2/storage"
	"github.com/stretchr/testify/suite"
	gtsmodel "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	gtsstorage "github.com/superseriousbusiness/gotosocial/internal/storage"
)

type ManagerTestSuite struct {
	MediaStandardTestSuite
}

func (suite *ManagerTestSuite) TestEmojiProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/rainbow-original.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	emojiID := "01GDQ9G782X42BAMFASKP64343"
	emojiURI := "http://localhost:8080/emoji/01GDQ9G782X42BAMFASKP64343"

	processingEmoji, err := suite.manager.ProcessEmoji(ctx, data, "rainbow_test", emojiID, emojiURI, nil, false)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processingEmoji.LoadEmoji(ctx)
	suite.NoError(err)
	suite.NotNil(emoji)

	// make sure it's got the stuff set on it that we expect
	suite.Equal(emojiID, emoji.ID)

	// file meta should be correctly derived from the image
	suite.Equal("image/png", emoji.ImageContentType)
	suite.Equal("image/png", emoji.ImageStaticContentType)
	suite.Equal(36702, emoji.ImageFileSize)

	// now make sure the emoji is in the database
	dbEmoji, err := suite.db.GetEmojiByID(ctx, emojiID)
	suite.NoError(err)
	suite.NotNil(dbEmoji)

	// make sure the processed emoji file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, emoji.ImagePath)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/rainbow-original.png")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedStaticBytes, err := suite.storage.Get(ctx, emoji.ImageStaticPath)
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytes)

	processedStaticBytesExpected, err := os.ReadFile("./test/rainbow-static.png")
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytesExpected)

	suite.Equal(processedStaticBytesExpected, processedStaticBytes)
}

func (suite *ManagerTestSuite) TestEmojiProcessBlockingRefresh() {
	ctx := context.Background()

	// we're going to 'refresh' the remote 'yell' emoji by changing the image url to the pixellated gts logo
	originalEmoji := suite.testEmojis["yell"]

	emojiToUpdate := &gtsmodel.Emoji{}
	*emojiToUpdate = *originalEmoji
	newImageRemoteURL := "http://fossbros-anonymous.io/some/image/path.png"

	oldEmojiImagePath := emojiToUpdate.ImagePath
	oldEmojiImageStaticPath := emojiToUpdate.ImageStaticPath

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		b, err := os.ReadFile("./test/gts_pixellated-original.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	emojiID := emojiToUpdate.ID
	emojiURI := emojiToUpdate.URI

	processingEmoji, err := suite.manager.ProcessEmoji(ctx, data, "yell", emojiID, emojiURI, &media.AdditionalEmojiInfo{
		CreatedAt:      &emojiToUpdate.CreatedAt,
		Domain:         &emojiToUpdate.Domain,
		ImageRemoteURL: &newImageRemoteURL,
	}, true)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processingEmoji.LoadEmoji(ctx)
	suite.NoError(err)
	suite.NotNil(emoji)

	// make sure it's got the stuff set on it that we expect
	suite.Equal(emojiID, emoji.ID)

	// file meta should be correctly derived from the image
	suite.Equal("image/png", emoji.ImageContentType)
	suite.Equal("image/png", emoji.ImageStaticContentType)
	suite.Equal(10296, emoji.ImageFileSize)

	// now make sure the emoji is in the database
	dbEmoji, err := suite.db.GetEmojiByID(ctx, emojiID)
	suite.NoError(err)
	suite.NotNil(dbEmoji)

	// make sure the processed emoji file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, emoji.ImagePath)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/gts_pixellated-original.png")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedStaticBytes, err := suite.storage.Get(ctx, emoji.ImageStaticPath)
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytes)

	processedStaticBytesExpected, err := os.ReadFile("./test/gts_pixellated-static.png")
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytesExpected)

	suite.Equal(processedStaticBytesExpected, processedStaticBytes)

	// most fields should be different on the emoji now from what they were before
	suite.Equal(originalEmoji.ID, dbEmoji.ID)
	suite.NotEqual(originalEmoji.ImageRemoteURL, dbEmoji.ImageRemoteURL)
	suite.NotEqual(originalEmoji.ImageURL, dbEmoji.ImageURL)
	suite.NotEqual(originalEmoji.ImageStaticURL, dbEmoji.ImageStaticURL)
	suite.NotEqual(originalEmoji.ImageFileSize, dbEmoji.ImageFileSize)
	suite.NotEqual(originalEmoji.ImageStaticFileSize, dbEmoji.ImageStaticFileSize)
	suite.NotEqual(originalEmoji.ImagePath, dbEmoji.ImagePath)
	suite.NotEqual(originalEmoji.ImageStaticPath, dbEmoji.ImageStaticPath)
	suite.NotEqual(originalEmoji.ImageStaticPath, dbEmoji.ImageStaticPath)
	suite.NotEqual(originalEmoji.UpdatedAt, dbEmoji.UpdatedAt)
	suite.NotEqual(originalEmoji.ImageUpdatedAt, dbEmoji.ImageUpdatedAt)

	// the old image files should no longer be in storage
	_, err = suite.storage.Get(ctx, oldEmojiImagePath)
	suite.ErrorIs(err, storage.ErrNotFound)
	_, err = suite.storage.Get(ctx, oldEmojiImageStaticPath)
	suite.ErrorIs(err, storage.ErrNotFound)
}

func (suite *ManagerTestSuite) TestEmojiProcessBlockingTooLarge() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/big-panda.gif")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	emojiID := "01GDQ9G782X42BAMFASKP64343"
	emojiURI := "http://localhost:8080/emoji/01GDQ9G782X42BAMFASKP64343"

	processingEmoji, err := suite.manager.ProcessEmoji(ctx, data, "big_panda", emojiID, emojiURI, nil, false)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processingEmoji.LoadEmoji(ctx)
	suite.EqualError(err, "store: given emoji size 630kiB greater than max allowed 50.0kiB")
	suite.Nil(emoji)
}

func (suite *ManagerTestSuite) TestEmojiProcessBlockingTooLargeNoSizeGiven() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/big-panda.gif")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), -1, nil
	}

	emojiID := "01GDQ9G782X42BAMFASKP64343"
	emojiURI := "http://localhost:8080/emoji/01GDQ9G782X42BAMFASKP64343"

	processingEmoji, err := suite.manager.ProcessEmoji(ctx, data, "big_panda", emojiID, emojiURI, nil, false)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processingEmoji.LoadEmoji(ctx)
	suite.EqualError(err, "store: calculated emoji size 630kiB greater than max allowed 50.0kiB")
	suite.Nil(emoji)
}

func (suite *ManagerTestSuite) TestEmojiProcessBlockingNoFileSizeGiven() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/rainbow-original.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), -1, nil
	}

	emojiID := "01GDQ9G782X42BAMFASKP64343"
	emojiURI := "http://localhost:8080/emoji/01GDQ9G782X42BAMFASKP64343"

	// process the media with no additional info provided
	processingEmoji, err := suite.manager.ProcessEmoji(ctx, data, "rainbow_test", emojiID, emojiURI, nil, false)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processingEmoji.LoadEmoji(ctx)
	suite.NoError(err)
	suite.NotNil(emoji)

	// make sure it's got the stuff set on it that we expect
	suite.Equal(emojiID, emoji.ID)

	// file meta should be correctly derived from the image
	suite.Equal("image/png", emoji.ImageContentType)
	suite.Equal("image/png", emoji.ImageStaticContentType)
	suite.Equal(36702, emoji.ImageFileSize)

	// now make sure the emoji is in the database
	dbEmoji, err := suite.db.GetEmojiByID(ctx, emojiID)
	suite.NoError(err)
	suite.NotNil(dbEmoji)

	// make sure the processed emoji file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, emoji.ImagePath)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/rainbow-original.png")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedStaticBytes, err := suite.storage.Get(ctx, emoji.ImageStaticPath)
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytes)

	processedStaticBytesExpected, err := os.ReadFile("./test/rainbow-static.png")
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytesExpected)

	suite.Equal(processedStaticBytesExpected, processedStaticBytes)
}

func (suite *ManagerTestSuite) TestEmojiWebpProcess() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/nb-flag-original.webp")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	emojiID := "01GDQ9G782X42BAMFASKP64343"
	emojiURI := "http://localhost:8080/emoji/01GDQ9G782X42BAMFASKP64343"

	processingEmoji, err := suite.manager.ProcessEmoji(ctx, data, "nb-flag", emojiID, emojiURI, nil, false)
	suite.NoError(err)

	// do a blocking call to fetch the emoji
	emoji, err := processingEmoji.LoadEmoji(ctx)
	suite.NoError(err)
	suite.NotNil(emoji)

	// make sure it's got the stuff set on it that we expect
	suite.Equal(emojiID, emoji.ID)

	// file meta should be correctly derived from the image
	suite.Equal("image/webp", emoji.ImageContentType)
	suite.Equal("image/png", emoji.ImageStaticContentType)
	suite.Equal(294, emoji.ImageFileSize)

	// now make sure the emoji is in the database
	dbEmoji, err := suite.db.GetEmojiByID(ctx, emojiID)
	suite.NoError(err)
	suite.NotNil(dbEmoji)

	// make sure the processed emoji file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, emoji.ImagePath)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/nb-flag-original.webp")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedStaticBytes, err := suite.storage.Get(ctx, emoji.ImageStaticPath)
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytes)

	processedStaticBytesExpected, err := os.ReadFile("./test/nb-flag-static.png")
	suite.NoError(err)
	suite.NotEmpty(processedStaticBytesExpected)

	suite.Equal(processedStaticBytesExpected, processedStaticBytes)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 1920, Height: 1080, Size: 2073600, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 288, Size: 147456, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Small)
	suite.Equal("image/jpeg", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(269739, attachment.File.FileSize)
	suite.Equal("LiBzRk#6V[WF_NvzV@WY_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessPartial() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}

		// Fuck up the bytes a bit by cutting
		// off the second half, tee hee!
		b = b[:len(b)/2]

		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)

	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)

	// Since we're cutting off the byte stream
	// halfway through, we should get an error here.
	suite.EqualError(err, "store: error writing media to storage: scan-data is unbounded; EOI not encountered before EOF")
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.Zero(attachment.FileMeta)
	suite.Equal("image/jpeg", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Empty(attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// Attachment should have type unknown
	suite.Equal(gtsmodel.FileTypeUnknown, dbAttachment.Type)

	// Nothing should be in storage for this attachment.
	stored, err := suite.storage.Has(ctx, attachment.File.Path)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.False(stored)

	stored, err = suite.storage.Has(ctx, attachment.Thumbnail.Path)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.False(stored)
}

func (suite *ManagerTestSuite) TestPDFProcess() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from Frantz
		b, err := os.ReadFile("./test/Frantz-Fanon-The-Wretched-of-the-Earth-1965.pdf")
		if err != nil {
			panic(err)
		}

		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)

	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.Zero(attachment.FileMeta)
	suite.Equal("application/pdf", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Empty(attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// Attachment should have type unknown
	suite.Equal(gtsmodel.FileTypeUnknown, dbAttachment.Type)

	// Nothing should be in storage for this attachment.
	stored, err := suite.storage.Has(ctx, attachment.File.Path)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.False(stored)

	stored, err = suite.storage.Has(ctx, attachment.Thumbnail.Path)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.False(stored)
}

func (suite *ManagerTestSuite) TestSlothVineProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test video
		b, err := os.ReadFile("./test/test-mp4-original.mp4")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the video
	suite.Equal(338, attachment.FileMeta.Original.Width)
	suite.Equal(240, attachment.FileMeta.Original.Height)
	suite.Equal(81120, attachment.FileMeta.Original.Size)
	suite.EqualValues(1.4083333, attachment.FileMeta.Original.Aspect)
	suite.EqualValues(6.640907, *attachment.FileMeta.Original.Duration)
	suite.EqualValues(29.000029, *attachment.FileMeta.Original.Framerate)
	suite.EqualValues(0x59e74, *attachment.FileMeta.Original.Bitrate)
	suite.EqualValues(gtsmodel.Small{
		Width: 338, Height: 240, Size: 81120, Aspect: 1.4083333333333334,
	}, attachment.FileMeta.Small)
	suite.Equal("video/mp4", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(312413, attachment.File.FileSize)
	suite.Equal("L00000fQfQfQfQfQfQfQfQfQfQfQ", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-mp4-processed.mp4")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-mp4-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestLongerMp4ProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test video
		b, err := os.ReadFile("./test/longer-mp4-original.mp4")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the video
	suite.Equal(600, attachment.FileMeta.Original.Width)
	suite.Equal(330, attachment.FileMeta.Original.Height)
	suite.Equal(198000, attachment.FileMeta.Original.Size)
	suite.EqualValues(1.8181819, attachment.FileMeta.Original.Aspect)
	suite.EqualValues(16.6, *attachment.FileMeta.Original.Duration)
	suite.EqualValues(10, *attachment.FileMeta.Original.Framerate)
	suite.EqualValues(0xc8fb, *attachment.FileMeta.Original.Bitrate)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 281, Size: 143872, Aspect: 1.822064,
	}, attachment.FileMeta.Small)
	suite.Equal("video/mp4", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(109549, attachment.File.FileSize)
	suite.Equal("L00000fQfQfQfQfQfQfQfQfQfQfQ", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/longer-mp4-processed.mp4")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/longer-mp4-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestBirdnestMp4ProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test video
		b, err := os.ReadFile("./test/birdnest-original.mp4")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the video
	suite.Equal(404, attachment.FileMeta.Original.Width)
	suite.Equal(720, attachment.FileMeta.Original.Height)
	suite.Equal(290880, attachment.FileMeta.Original.Size)
	suite.EqualValues(0.5611111, attachment.FileMeta.Original.Aspect)
	suite.EqualValues(9.822041, *attachment.FileMeta.Original.Duration)
	suite.EqualValues(30, *attachment.FileMeta.Original.Framerate)
	suite.EqualValues(0x117c79, *attachment.FileMeta.Original.Bitrate)
	suite.EqualValues(gtsmodel.Small{
		Width: 287, Height: 512, Size: 146944, Aspect: 0.5605469,
	}, attachment.FileMeta.Small)
	suite.Equal("video/mp4", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(1409577, attachment.File.FileSize)
	suite.Equal("L00000fQfQfQfQfQfQfQfQfQfQfQ", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/birdnest-processed.mp4")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/birdnest-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestNotAnMp4ProcessBlocking() {
	// try to load an 'mp4' that's actually an mkv in disguise

	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test video
		b, err := os.ReadFile("./test/not-an.mp4")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// pre processing should go fine but...
	processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)

	// we should get an error while loading
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.EqualError(err, "finish: error decoding video: error determining video metadata: [width height framerate]")

	// partial attachment should be
	// returned, with 'unknown' type.
	suite.NotNil(attachment)
	suite.Equal(gtsmodel.FileTypeUnknown, attachment.Type)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessBlockingNoContentLengthGiven() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		// give length as -1 to indicate unknown
		return io.NopCloser(bytes.NewBuffer(b)), -1, nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 1920, Height: 1080, Size: 2073600, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 288, Size: 147456, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Small)
	suite.Equal("image/jpeg", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(269739, attachment.File.FileSize)
	suite.Equal("LiBzRk#6V[WF_NvzV@WY_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessBlockingReadCloser() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// open test image as a file
		f, err := os.Open("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		// give length as -1 to indicate unknown
		return f, -1, nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 1920, Height: 1080, Size: 2073600, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 288, Size: 147456, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Small)
	suite.Equal("image/jpeg", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(269739, attachment.File.FileSize)
	suite.Equal("LiBzRk#6V[WF_NvzV@WY_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestPngNoAlphaChannelProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-png-noalphachannel.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 186, Height: 187, Size: 34782, Aspect: 0.9946524064171123,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 186, Height: 187, Size: 34782, Aspect: 0.9946524064171123,
	}, attachment.FileMeta.Small)
	suite.Equal("image/png", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(17471, attachment.File.FileSize)
	suite.Equal("LFQT7e.A%O%4?co$M}M{_1W9~TxV", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-png-noalphachannel-processed.png")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-png-noalphachannel-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestPngAlphaChannelProcessBlocking() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-png-alphachannel.png")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 186, Height: 187, Size: 34782, Aspect: 0.9946524064171123,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 186, Height: 187, Size: 34782, Aspect: 0.9946524064171123,
	}, attachment.FileMeta.Small)
	suite.Equal("image/png", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(18904, attachment.File.FileSize)
	suite.Equal("LFQT7e.A%O%4?co$M}M{_1W9~TxV", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-png-alphachannel-processed.png")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-png-alphachannel-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessBlockingWithCallback() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	// process the media with no additional info provided
	processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 1920, Height: 1080, Size: 2073600, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 288, Size: 147456, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Small)
	suite.Equal("image/jpeg", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(269739, attachment.File.FileSize)
	suite.Equal("LiBzRk#6V[WF_NvzV@WY_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := suite.storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := suite.storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestSimpleJpegProcessBlockingWithDiskStorage() {
	ctx := context.Background()

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// load bytes from a test image
		b, err := os.ReadFile("./test/test-jpeg.jpg")
		if err != nil {
			panic(err)
		}
		return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
	}

	accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

	temp := fmt.Sprintf("%s/gotosocial-test", os.TempDir())
	defer os.RemoveAll(temp)

	disk, err := storage.OpenDisk(temp, &storage.DiskConfig{
		LockFile: path.Join(temp, "store.lock"),
	})
	if err != nil {
		panic(err)
	}

	var state state.State

	state.Workers.Start()
	defer state.Workers.Stop()

	storage := &gtsstorage.Driver{
		Storage: disk,
	}
	state.Storage = storage
	state.DB = suite.db

	diskManager := media.NewManager(&state)
	suite.manager = diskManager

	// process the media with no additional info provided
	processingMedia := diskManager.PreProcessMedia(data, accountID, nil)
	// fetch the attachment id from the processing media
	attachmentID := processingMedia.AttachmentID()

	// do a blocking call to fetch the attachment
	attachment, err := processingMedia.LoadAttachment(ctx)
	suite.NoError(err)
	suite.NotNil(attachment)

	// make sure it's got the stuff set on it that we expect
	// the attachment ID and accountID we expect
	suite.Equal(attachmentID, attachment.ID)
	suite.Equal(accountID, attachment.AccountID)

	// file meta should be correctly derived from the image
	suite.EqualValues(gtsmodel.Original{
		Width: 1920, Height: 1080, Size: 2073600, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Original)
	suite.EqualValues(gtsmodel.Small{
		Width: 512, Height: 288, Size: 147456, Aspect: 1.7777777777777777,
	}, attachment.FileMeta.Small)
	suite.Equal("image/jpeg", attachment.File.ContentType)
	suite.Equal("image/jpeg", attachment.Thumbnail.ContentType)
	suite.Equal(269739, attachment.File.FileSize)
	suite.Equal("LiBzRk#6V[WF_NvzV@WY_3rqV@a$", attachment.Blurhash)

	// now make sure the attachment is in the database
	dbAttachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
	suite.NoError(err)
	suite.NotNil(dbAttachment)

	// make sure the processed file is in storage
	processedFullBytes, err := storage.Get(ctx, attachment.File.Path)
	suite.NoError(err)
	suite.NotEmpty(processedFullBytes)

	// load the processed bytes from our test folder, to compare
	processedFullBytesExpected, err := os.ReadFile("./test/test-jpeg-processed.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedFullBytesExpected)

	// the bytes in storage should be what we expected
	suite.Equal(processedFullBytesExpected, processedFullBytes)

	// now do the same for the thumbnail and make sure it's what we expected
	processedThumbnailBytes, err := storage.Get(ctx, attachment.Thumbnail.Path)
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytes)

	processedThumbnailBytesExpected, err := os.ReadFile("./test/test-jpeg-thumbnail.jpg")
	suite.NoError(err)
	suite.NotEmpty(processedThumbnailBytesExpected)

	suite.Equal(processedThumbnailBytesExpected, processedThumbnailBytes)
}

func (suite *ManagerTestSuite) TestSmallSizedMediaTypeDetection_issue2263() {
	for index, test := range []struct {
		name     string // Test title
		path     string // File path
		expected string // Expected ContentType
	}{
		{
			name:     "big size JPEG",
			path:     "./test/test-jpeg.jpg",
			expected: "image/jpeg",
		},
		{
			name:     "big size PNG",
			path:     "./test/test-png-noalphachannel.png",
			expected: "image/png",
		},
		{
			name:     "small size JPEG",
			path:     "./test/test-jpeg-1x1px-white.jpg",
			expected: "image/jpeg",
		},
		{
			name:     "golden case PNG (big size)",
			path:     "./test/test-png-alphachannel-1x1px.png",
			expected: "image/png",
		},
	} {
		suite.Run(test.name, func() {
			ctx, cncl := context.WithTimeout(context.Background(), time.Second*60)
			defer cncl()

			data := func(_ context.Context) (io.ReadCloser, int64, error) {
				// load bytes from a test image
				b, err := os.ReadFile(test.path)
				suite.NoError(err, "Test %d: failed during test setup", index+1)

				return io.NopCloser(bytes.NewBuffer(b)), int64(len(b)), nil
			}

			accountID := "01FS1X72SK9ZPW0J1QQ68BD264"

			// process the media with no additional info provided
			processingMedia := suite.manager.PreProcessMedia(data, accountID, nil)
			if _, err := processingMedia.LoadAttachment(ctx); err != nil {
				suite.FailNow(err.Error())
			}

			attachmentID := processingMedia.AttachmentID()

			// fetch the attachment id from the processing media
			attachment, err := suite.db.GetAttachmentByID(ctx, attachmentID)
			if err != nil {
				suite.FailNow(err.Error())
			}

			// make sure it's got the stuff set on it that we expect
			// the attachment ID and accountID we expect
			suite.Equal(attachmentID, attachment.ID)
			suite.Equal(accountID, attachment.AccountID)

			actual := attachment.File.ContentType
			expect := test.expected

			suite.Equal(expect, actual, "Test %d: %s", index+1, test.name)
		})
	}
}

func (suite *ManagerTestSuite) TestMisreportedSmallMedia() {
	const accountID = "01FS1X72SK9ZPW0J1QQ68BD264"
	var actualSize int

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// Load bytes from small png.
		b, err := os.ReadFile("./test/test-png-alphachannel-1x1px.png")
		if err != nil {
			suite.FailNow(err.Error())
		}

		actualSize = len(b)

		// Report media as twice its actual size. This should be corrected.
		return io.NopCloser(bytes.NewBuffer(b)), int64(2 * actualSize), nil
	}

	// Process the media with no additional info provided.
	attachment, err := suite.manager.
		PreProcessMedia(data, accountID, nil).
		LoadAttachment(context.Background())
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(actualSize, attachment.File.FileSize)
}

func (suite *ManagerTestSuite) TestNoReportedSizeSmallMedia() {
	const accountID = "01FS1X72SK9ZPW0J1QQ68BD264"
	var actualSize int

	data := func(_ context.Context) (io.ReadCloser, int64, error) {
		// Load bytes from small png.
		b, err := os.ReadFile("./test/test-png-alphachannel-1x1px.png")
		if err != nil {
			suite.FailNow(err.Error())
		}

		actualSize = len(b)

		// Return zero for media size. This should be detected.
		return io.NopCloser(bytes.NewBuffer(b)), 0, nil
	}

	// Process the media with no additional info provided.
	attachment, err := suite.manager.
		PreProcessMedia(data, accountID, nil).
		LoadAttachment(context.Background())
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(actualSize, attachment.File.FileSize)
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}
