package playwrightgen

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/zikani03/basi"
)

var ErrNotAPlaywrightAction = errors.New("not a playwright action")

type PlawrightGenFunc func(w io.Writer, action *basi.Action) error

var SupportedActions = actionMap

func notSupportedByPW(w io.Writer, action *basi.Action) error {
	return ErrNotAPlaywrightAction
}

func screenshotFunc(w io.Writer, action *basi.Action) error {
	args := action.Arguments.String
	_, err := fmt.Fprintf(w, "\tawait page.screenshot({ path:  %q });\n", args)
	return err
}

func actionFunc(hasargs ...bool) PlawrightGenFunc {
	includeArgs := len(hasargs) == 1 && hasargs[0]
	return func(w io.Writer, action *basi.Action) error {
		var err error
		if includeArgs && action.Arguments != nil {
			args := action.Arguments.String
			_, err = fmt.Fprintf(w, "\tawait page.%s(%q, %q);\n", strings.ToLower(action.Action), action.Selector.Selector, args)
		} else {
			_, err = fmt.Fprintf(w, "\tawait page.%s(%q);\n", strings.ToLower(action.Action), action.Selector.Selector)
		}
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}
		return nil
	}
}

type ExpectFunc func(matcher string) PlawrightGenFunc

func assertFunc(matcher string, hasargs ...bool) PlawrightGenFunc {
	includeArgs := len(hasargs) == 1 && hasargs[0]

	return func(w io.Writer, action *basi.Action) error {
		expectExpr := matcher
		args := ""
		var err error
		if includeArgs && action.Arguments != nil {
			args = action.Arguments.String
			_, err = fmt.Fprintf(w, "\tawait expect(page.locator(%q)).%s(%q);\n", action.Selector.Selector, expectExpr, args)
		} else {
			_, err = fmt.Fprintf(w, "\tawait expect(page.locator(%q)).%s();\n", action.Selector.Selector, expectExpr)
		}
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}
		return nil
	}
}

var actionMap = map[string]PlawrightGenFunc{
	"Use":                   actionFunc(true),
	"Click":                 actionFunc(),
	"DoubleClick":           actionFunc(),
	"Doubleclick":           actionFunc(),
	"Tap":                   actionFunc(),
	"Focus":                 actionFunc(),
	"Blur":                  actionFunc(),
	"Clear":                 actionFunc(),
	"Fill":                  actionFunc(true),
	"Find":                  notSupportedByPW,
	"Check":                 actionFunc(),
	"Uncheck":               actionFunc(),
	"FillCheckbox":          actionFunc(), // alias for Check
	"Press":                 actionFunc(true),
	"PressSequentially":     actionFunc(true),
	"Select":                actionFunc(true), // alias for SelectOption
	"SelectOption":          actionFunc(true),
	"SelectMultipleOptions": actionFunc(true),
	"Type":                  actionFunc(true), // alias for PressSequentially
	"WaitFor":               actionFunc(true),
	"WaitForSelector":       actionFunc(true),
	"WaitForURL":            actionFunc(true),
	"Goto":                  actionFunc(),
	"GoBack":                actionFunc(),
	"GoForward":             actionFunc(),
	"Refresh":               actionFunc(),
	"Screenshot":            screenshotFunc,
	"Upload":                actionFunc(true), // alias for UploadFile
	"UploadFile":            actionFunc(true),
	"UploadFiles":           actionFunc(true), // alias for UploadMultipleFiles
	"UploadMultipleFiles":   actionFunc(true),
	"ExpectText":            assertFunc("toContainText", true),
	"ExpectAttr":            assertFunc("toHaveAttribute", true),
	"ExpectAttribute":       assertFunc("toHaveAttribute", true),
	"ExpectValue":           assertFunc("toHaveValue", true),
	"ExpectValues":          assertFunc("toHaveValues", true),
	"ExpectAttached":        assertFunc("toBeAttached"),
	"ExpectChecked":         assertFunc("toBeChecked"),
	"ExpectDisabled":        assertFunc("toBeDisabled"),
	"ExpectEditable":        assertFunc("toBeEditable"),
	"ExpectEmpty":           assertFunc("toBeEmpty"),
	"ExpectEnabled":         assertFunc("toBeEnabled"),
	"ExpectFocused":         assertFunc("toBeFocused"),
	"ExpectHidden":          assertFunc("toBeHidden"),
	"ExpectInViewport":      assertFunc("toBeInViewport"),
	"ExpectVisible":         assertFunc("toBeVisible"),
	"ExpectToContainClass":  assertFunc("toContainClass", true),
	"ExpectToContainText":   assertFunc("toContainText", true),
	// "ExpectAccessibleDescription":  assertFunc("toBeAccessibleDescription"),
	// "ExpectAccessibleErrorMessage": assertFunc("toBeAccessibleErrorMessage"),
	// "ExpectAccessibleName":      assertFunc("toBeAccessibleName"),
	// "ExpectToMatchAriaSnapshot": assertFunc("toBeToMatchAriaSnapshot"),
	"ExpectClass":      assertFunc("toHaveClass", true),
	"ExpectCSS":        assertFunc("toHaveCSS", true),
	"ExpectId":         assertFunc("toHaveId", true),
	"ExpectJSProperty": assertFunc("toHaveJSProperty", true),
}
