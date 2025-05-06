package v4playwrightactionparser

import (
	"io"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
)

type ActionFunc int

var actionMap = map[string]ActionFunc{
	"Click":                 0,
	"DoubleClick":           0,
	"Doubleclick":           0,
	"Tap":                   0,
	"Focus":                 0,
	"Blur":                  0,
	"Clear":                 0,
	"Fill":                  0,
	"Check":                 0,
	"Uncheck":               0,
	"FillCheckbox":          0, // alias for Check
	"Press":                 0,
	"PressSequentially":     0,
	"Select":                0, // alias for SelectOption
	"SelectOption":          0,
	"SelectMultipleOptions": 0,
	"Type":                  0, // alias for PressSequentially
	"WaitFor":               0,
	"WaitForSelector":       0,
	"WaitForURL":            0,
	"Goto":                  0,
	"GoBack":                0,
	"GoForward":             0,
	"Refresh":               0,
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
		{`Ident`, `[a-zA-Z][a-zA-Z_\d]*`},
		{`String`, `"(?:\\.|[^"])*"`},
		{`Selector`, `"(?:\\.|[^"])*"`},
		{"comment", `[#;][^\n]*`},
		{"Whitespace", `[ \s]+`},
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
	Actions []*Action `@@*`
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

func DebugParse(r io.Reader) (any, error) {
	actions, err := parser.Parse("tests.yaml", r)
	repr.Println(actions, repr.Indent("  "), repr.OmitEmpty(true))
	if err != nil {
		return nil, err
	}
	return actions, nil
}

func Parse(r io.Reader) (*PlaywrightAction, error) {
	actions, err := parser.Parse("tests.yaml", r)
	if err != nil {
		return nil, err
	}
	return actions, nil
}
