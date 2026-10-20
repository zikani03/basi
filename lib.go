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
	"FindRegex", // alias of FindMatching
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
	// Property-based testing actions
	"Always",
	"Fuzz",
	"Extract",
	"Eventually",
	"Next",
}

// operandlessActions are actions that are legitimately allowed to appear
// without any operands (Selector, Arguments, Variable or Number). These are
// navigation-only actions. Every other action requires at least one operand.
var operandlessActions = map[string]bool{
	"GoBack":    true,
	"GoForward": true,
	"Refresh":   true,
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
		{`MetaField`, `ID|Title|URL|Description|Headless|Browsers|Browser|CDPEndpoint|Extends|FakerSeed`}, // TODO: support ScreenSizes
		{`Ident`, `[a-zA-Z][a-zA-Z_\d]*`},
		{`Variable`, `\$[a-zA-Z][a-zA-Z_\d]*`},
		{`Number`, `\d+`},
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
	Selector  *Selector `[ @@ ]`
	Arguments *String   `[ @@ ]`
	Variable  *Variable `[ @@ ]`
	Number    *Number   `[ @@ ]`
	comment   *string   `@comment`
}

type Variable struct {
	// Pos     lexer.Position
	Variable string `@Variable`
}

func (Variable) value() {}

type Number struct {
	// Pos    lexer.Position
	Number int `@Number`
}

func (Number) value() {}

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

	// Post-parse validation: every action that is not explicitly navigation-only
	// must carry at least one operand (Selector, Arguments, Variable or Number).
	for _, action := range pwAction.Actions {
		if operandlessActions[action.Action] {
			continue
		}
		if action.Selector == nil && action.Arguments == nil && action.Variable == nil && action.Number == nil {
			return nil, fmt.Errorf("action %q requires at least one argument", action.Action)
		}
	}

	return pwAction, nil
}

// ResolveExtends resolves the optional "Extends" frontmatter field of parsed,
// inheriting frontmatter fields from the referenced .basi file(s).
//
// The Extends value grammar is:
//
//	"<path>"               -> inherit ALL fields the referenced file declares
//	"<path>;Field1,Field2" -> inherit only the named fields (trimmed)
//
// The value is split on the FIRST ';'. The left side is a file path resolved
// relative to the directory of basePath (the current .basi file); an absolute
// path is used as-is. The referenced file may itself declare Extends, which is
// resolved recursively first so that grandparent values flow down.
//
// Merge semantics are "local wins": a parent field is only appended when the
// current file does not already declare a field with the same name; a field the
// current file already declares is never overwritten.
//
// If parsed.Meta is nil or there is no Extends field, this is a no-op.
func ResolveExtends(parsed *PlaywrightAction, basePath string) error {
	return resolveExtends(parsed, basePath, map[string]bool{})
}

// resolveExtends is the recursive helper behind ResolveExtends. The seen set
// holds the absolute paths of .basi files already visited on the current chain
// so circular references can be detected.
func resolveExtends(parsed *PlaywrightAction, basePath string, seen map[string]bool) error {
	if parsed == nil || parsed.Meta == nil {
		return nil
	}

	extendsValue := parsed.GetMetaFieldString("Extends")
	if extendsValue == "" {
		return nil
	}

	// Split on the first ';' to separate the path from the optional section list.
	pathPart := extendsValue
	var sections []string
	if idx := strings.Index(extendsValue, ";"); idx >= 0 {
		pathPart = extendsValue[:idx]
		for _, s := range strings.Split(extendsValue[idx+1:], ",") {
			if s = strings.TrimSpace(s); s != "" {
				sections = append(sections, s)
			}
		}
	}
	pathPart = strings.TrimSpace(pathPart)
	if pathPart == "" {
		return fmt.Errorf("Extends field has an empty path: %q", extendsValue)
	}

	// Resolve the referenced path relative to the current file's directory.
	refPath := pathPart
	if !filepath.IsAbs(refPath) {
		refPath = filepath.Join(filepath.Dir(basePath), refPath)
	}
	absPath, err := filepath.Abs(refPath)
	if err != nil {
		return fmt.Errorf("failed to resolve Extends path %q: %w", pathPart, err)
	}
	absPath = filepath.Clean(absPath)

	// Circular-reference detection: if this file is already on the chain, bail.
	if seen[absPath] {
		return fmt.Errorf("circular Extends reference detected: %s", absPath)
	}
	seen[absPath] = true

	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read Extends file %s: %w", absPath, err)
	}

	parent, err := Parse(absPath, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to parse Extends file %s: %w", absPath, err)
	}

	// Resolve the parent's own Extends first so grandparent values flow down,
	// resolving its relative path against its own directory.
	if err := resolveExtends(parent, absPath, seen); err != nil {
		return err
	}

	if parent.Meta == nil {
		return nil
	}

	// Determine which parent fields to inherit.
	var inherit []*MetaField
	if len(sections) == 0 {
		// No ';' section list: inherit every field the parent declares.
		inherit = parent.Meta.Fields
	} else {
		// Only inherit the named sections. A named section that the parent does
		// not declare is silently skipped (chosen for ergonomics so filters can
		// name fields that only some bases provide) rather than raising an error.
		for _, name := range sections {
			for _, f := range parent.Meta.Fields {
				if f.Name == name {
					inherit = append(inherit, f)
				}
			}
		}
	}

	// Merge with local-wins semantics: never overwrite a field the current file
	// already declares. The current file's own Extends field is not inherited.
	for _, pf := range inherit {
		if pf.Name == "Extends" {
			continue
		}
		if parsed.GetMetaFieldString(pf.Name) != "" || hasField(parsed.Meta, pf.Name) {
			continue
		}
		parsed.Meta.Fields = append(parsed.Meta.Fields, &MetaField{Name: pf.Name, Value: pf.Value})
	}

	return nil
}

// hasField reports whether fm declares a field with the given name, including
// fields whose value is the empty string (which GetMetaFieldString cannot
// distinguish from an absent field).
func hasField(fm *FrontMatter, name string) bool {
	if fm == nil {
		return false
	}
	for _, f := range fm.Fields {
		if f.Name == name {
			return true
		}
	}
	return false
}
