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

package router

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

const (
	justTime     = "15:04"
	dateYear     = "Jan 02, 2006"
	dateTime     = "Jan 02, 15:04"
	dateYearTime = "Jan 02, 2006, 15:04"
	monthYear    = "Jan, 2006"
	badTimestamp = "bad timestamp"
)

// LoadTemplates loads html templates for use by the given engine
func LoadTemplates(engine *gin.Engine) error {
	templateBaseDir := config.GetWebTemplateBaseDir()
	if templateBaseDir == "" {
		return fmt.Errorf("%s cannot be empty and must be a relative or absolute path", config.WebTemplateBaseDirFlag())
	}

	templateBaseDir, err := filepath.Abs(templateBaseDir)
	if err != nil {
		return fmt.Errorf("error getting absolute path of %s: %s", templateBaseDir, err)
	}

	if _, err := os.Stat(filepath.Join(templateBaseDir, "index.tmpl")); err != nil {
		return fmt.Errorf("%s doesn't seem to contain the templates; index.tmpl is missing: %w", templateBaseDir, err)
	}

	engine.LoadHTMLGlob(filepath.Join(templateBaseDir, "*"))
	return nil
}

func oddOrEven(n int) string {
	if n%2 == 0 {
		return "even"
	}
	return "odd"
}

func escape(str string) template.HTML {
	/* #nosec G203 */
	return template.HTML(template.HTMLEscapeString(str))
}

func noescape(str string) template.HTML {
	/* #nosec G203 */
	return template.HTML(str)
}

func noescapeAttr(str string) template.HTMLAttr {
	/* #nosec G203 */
	return template.HTMLAttr(str)
}

func timestamp(stamp string) string {
	t, err := util.ParseISO8601(stamp)
	if err != nil {
		log.Errorf(nil, "error parsing timestamp %s: %s", stamp, err)
		return badTimestamp
	}

	t = t.Local()

	tYear, tMonth, tDay := t.Date()
	now := time.Now()
	currentYear, currentMonth, currentDay := now.Date()

	switch {
	case tYear == currentYear && tMonth == currentMonth && tDay == currentDay:
		return "Today, " + t.Format(justTime)
	case tYear == currentYear:
		return t.Format(dateTime)
	default:
		return t.Format(dateYear)
	}
}

func timestampPrecise(stamp string) string {
	t, err := util.ParseISO8601(stamp)
	if err != nil {
		log.Errorf(nil, "error parsing timestamp %s: %s", stamp, err)
		return badTimestamp
	}
	return t.Local().Format(dateYearTime)
}

func timestampVague(stamp string) string {
	t, err := util.ParseISO8601(stamp)
	if err != nil {
		log.Errorf(nil, "error parsing timestamp %s: %s", stamp, err)
		return badTimestamp
	}
	return t.Format(monthYear)
}

type iconWithLabel struct {
	faIcon string
	label  string
}

func visibilityIcon(visibility apimodel.Visibility) template.HTML {
	var icon iconWithLabel

	switch visibility {
	case apimodel.VisibilityPublic:
		icon = iconWithLabel{"globe", "public"}
	case apimodel.VisibilityUnlisted:
		icon = iconWithLabel{"unlock", "unlisted"}
	case apimodel.VisibilityPrivate:
		icon = iconWithLabel{"lock", "private"}
	case apimodel.VisibilityMutualsOnly:
		icon = iconWithLabel{"handshake-o", "mutuals only"}
	case apimodel.VisibilityDirect:
		icon = iconWithLabel{"envelope", "direct"}
	}

	/* #nosec G203 */
	return template.HTML(fmt.Sprintf(`<i aria-label="Visibility: %v" class="fa fa-%v"></i>`, icon.label, icon.faIcon))
}

// text is a template.HTML to affirm that the input of this function is already escaped
func emojify(emojis []apimodel.Emoji, inputText template.HTML) template.HTML {
	out := text.Emojify(emojis, string(inputText))

	/* #nosec G203 */
	// (this is escaped above)
	return template.HTML(out)
}

func acctInstance(acct string) string {
	parts := strings.Split(acct, "@")
	if len(parts) > 1 {
		return "@" + parts[1]
	}

	return ""
}

func increment(i int) int {
	return i + 1
}

// isNil will safely check if 'v' is nil without
// dealing with weird Go interface nil bullshit.
func isNil(i interface{}) bool {
	type eface struct{ _, data unsafe.Pointer }
	return (*eface)(unsafe.Pointer(&i)).data == nil
}

// deref returns the dereferenced value of
// its input. To ensure you don't pass nil
// pointers into this func, use isNil first.
func deref(i any) any {
	vOf := reflect.ValueOf(i)
	if vOf.Kind() != reflect.Pointer {
		// Not a pointer.
		return i
	}

	return vOf.Elem()
}

func LoadTemplateFunctions(engine *gin.Engine) {
	engine.SetFuncMap(template.FuncMap{
		"escape":           escape,
		"noescape":         noescape,
		"noescapeAttr":     noescapeAttr,
		"oddOrEven":        oddOrEven,
		"visibilityIcon":   visibilityIcon,
		"timestamp":        timestamp,
		"timestampVague":   timestampVague,
		"timestampPrecise": timestampPrecise,
		"emojify":          emojify,
		"acctInstance":     acctInstance,
		"increment":        increment,
		"isNil":            isNil,
		"deref":            deref,
	})
}
