package basi

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
)

type ActionFunc int

const StubExpectAction = 0

var actionMap = map[string]ActionFunc{
	"Use": StubExpectAction, // Use allows users to import from other files
	// Playwright Actions
	"Click":                        StubExpectAction,
	"DoubleClick":                  StubExpectAction,
	"Doubleclick":                  StubExpectAction,
	"Tap":                          StubExpectAction,
	"Focus":                        StubExpectAction,
	"Blur":                         StubExpectAction,
	"Clear":                        StubExpectAction,
	"Fill":                         StubExpectAction,
	"Find":                         StubExpectAction,
	"Check":                        StubExpectAction,
	"Uncheck":                      StubExpectAction,
	"FillCheckbox":                 StubExpectAction,
	"Press":                        StubExpectAction,
	"PressSequentially":            StubExpectAction,
	"Select":                       StubExpectAction,
	"SelectOption":                 StubExpectAction,
	"SelectMultipleOptions":        StubExpectAction,
	"Type":                         StubExpectAction,
	"WaitFor":                      StubExpectAction,
	"WaitForSelector":              StubExpectAction,
	"WaitForURL":                   StubExpectAction,
	"Goto":                         StubExpectAction,
	"GoBack":                       StubExpectAction,
	"GoForward":                    StubExpectAction,
	"Refresh":                      StubExpectAction,
	"Screenshot":                   StubExpectAction,
	"Upload":                       StubExpectAction,
	"UploadFile":                   StubExpectAction,
	"UploadFiles":                  StubExpectAction,
	"UploadMultipleFiles":          StubExpectAction,
	"ExpectText":                   StubExpectAction,
	"ExpectAttr":                   StubExpectAction,
	"ExpectAttribute":              StubExpectAction,
	"ExpectValue":                  StubExpectAction,
	"ExpectValues":                 StubExpectAction,
	"ExpectAttached":               StubExpectAction,
	"ExpectChecked":                StubExpectAction,
	"ExpectDisabled":               StubExpectAction,
	"ExpectEditable":               StubExpectAction,
	"ExpectEmpty":                  StubExpectAction,
	"ExpectEnabled":                StubExpectAction,
	"ExpectFocused":                StubExpectAction,
	"ExpectHidden":                 StubExpectAction,
	"ExpectInViewport":             StubExpectAction,
	"ExpectVisible":                StubExpectAction,
	"ExpectToContainClass":         StubExpectAction,
	"ExpectToContainText":          StubExpectAction,
	"ExpectAccessibleDescription":  StubExpectAction,
	"ExpectAccessibleErrorMessage": StubExpectAction,
	"ExpectAccessibleName":         StubExpectAction,
	"ExpectClass":                  StubExpectAction,
	"ExpectCSS":                    StubExpectAction,
	"ExpectId":                     StubExpectAction,
	"ExpectJSProperty":             StubExpectAction,
	"ExpectToMatchAriaSnapshot":    StubExpectAction,
}

func lexerActionsFromMap() string {
	names := make([]string, 0)
	for name, _ := range actionMap {
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

	return pwAction, nil
}
