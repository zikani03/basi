package basi

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
)

var Version = "devel"

type ActionFunc int

var actions = []string{
	"Use", // Use allows users to import from other files
	// Playwright Actions
	"Click",
	"DoubleClick",
	"Doubleclick",
	"Tap",
	"Focus",
	"Blur",
	"Clear",
	"Fill",
	"FindNth",
	"FindMatching",
	"FindFirst",
	"FindLast",
	"Find",
	"Check",
	"Uncheck",
	"FillCheckbox",
	"Press",
	"PressSequentially",
	"Select",
	"SelectOption",
	"SelectMultipleOptions",
	"Type",
	"WaitFor",
	"WaitForSelector",
	"WaitForURL",
	"Goto",
	"GoBack",
	"GoForward",
	"Refresh",
	"Screenshot",
	"Upload",
	"UploadFile",
	"UploadFiles",
	"UploadMultipleFiles",
	"ExpectText",
	"ExpectAttr",
	"ExpectAttribute",
	"ExpectValue",
	"ExpectValues",
	"ExpectAttached",
	"ExpectChecked",
	"ExpectDisabled",
	"ExpectEditable",
	"ExpectEmpty",
	"ExpectEnabled",
	"ExpectFocused",
	"ExpectHidden",
	"ExpectInViewport",
	"ExpectVisible",
	"ExpectToContainClass",
	"ExpectToContainText",
	"ExpectAccessibleDescription",
	"ExpectAccessibleErrorMessage",
	"ExpectAccessibleName",
	"ExpectClass",
	"ExpectCSS",
	"ExpectId",
	"ExpectJSProperty",
	"ExpectToMatchAriaSnapshot",
}

func lexerActionsFromMap() string {
	names := make([]string, 0)
	for _, name := range actions {
		names = append(names, name)
	}
	return strings.Join(names, "|")
}

var (
	actionLexer = lexer.MustSimple([]lexer.SimpleRule{
		{`Action`, lexerActionsFromMap()},
		{`MetaField`, `ID|Title|URL|Description|Headless|Browser`}, // TODO: support ScreenSizes, Extends
		{`Ident`, `[a-zA-Z][a-zA-Z_\d]*`},
		{`String`, `"(?:\\.|[^"])*"`},
		{`Selector`, `"(?:\\.|[^"])*"`},
		{"comment", `[#;][^\n]*`},
		{"Whitespace", `[ \s]+`},
		{"Colon", `[:]+`},
		{"Separator", `\-{3}`},
		{"EOL", `[\n\r]+`},
	})
	parser = participle.MustBuild[PlaywrightAction](
		participle.Lexer(actionLexer),
		participle.Unquote("String"),
		participle.Elide("Whitespace", "EOL"),
		participle.Union[Value](String{}),
	)
)

type PlaywrightAction struct {
	Meta    *FrontMatter `@@`
	Actions []*Action    `@@*`
}

func (p *PlaywrightAction) GetMetaFieldString(name string) string {
	if p.Meta == nil {
		return ""
	}
	if p.Meta.Fields == nil {
		return ""
	}
	for _, field := range p.Meta.Fields {
		if field.Name == name {
			return field.Value
		}
	}
	return ""
}

type FrontMatter struct {
	Fields    []*MetaField `@@*`
	Separator *string      `@Separator*`
}

type MetaField struct {
	// Pos   lexer.Position
	Name  string `@MetaField`
	Value string `":" @String`
}

type Action struct {
	// Pos lexer.Position
	Action    string    `@Action`
	Selector  *Selector `@@`
	Arguments *String   `[ @@ ]`
	comment   *string   `@comment`
}

type String struct {
	// Pos    lexer.Position
	String string `@String`
}

func (String) value() {}

type Selector struct {
	// Pos      lexer.Position
	Selector string `@String | "GetByRole"(@String)`
}

func (Selector) value() {}

type Value interface{ value() }

func DebugParse(filename string, r io.Reader) (*PlaywrightAction, error) {
	actions, err := Parse(filename, r)
	repr.Println(actions, repr.Indent("  "), repr.OmitEmpty(true))
	if err != nil {
		return nil, err
	}
	return actions, nil
}

func Parse(filename string, r io.Reader) (*PlaywrightAction, error) {
	pwAction, err := parser.Parse(filename, r)
	if err != nil {
		return nil, err
	}

	allActions := make([]*Action, 0)

	for _, action := range pwAction.Actions {
		if action.Action == "Use" {
			useFilename := action.Selector.Selector
			// resolve relative path files properly
			if strings.HasPrefix(useFilename, ".") {
				useFilename = filepath.Clean(filepath.Join(filepath.Dir(filename), useFilename))
			}
			data, err := os.ReadFile(useFilename)
			if err != nil {
				return nil, fmt.Errorf("failed to import file %s: %w", useFilename, err)
			}
			useAct, err := parser.Parse(useFilename, bytes.NewReader(data))
			if err != nil {
				return nil, fmt.Errorf("failed to parse actions from imported file %s: %w", useFilename, err)
			}
			allActions = append(allActions, useAct.Actions...)
		} else {
			allActions = append(allActions, action)
		}
	}

	pwAction.Actions = allActions
	return pwAction, nil
}
