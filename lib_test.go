package basi

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	content := `
Fill "x" "something something here" 
Fill "#x" "something something here"
Fill ".x" "something something here"
Fill "[name=email]" "something something here"
Fill "some element with whitespace" "something something here"
Fill ".input[type=\"text\"]" "something something here"
Click "existent" "something"
Click "GetByRole( \"existent\" )"
Click    "existent"   "this should be accepted" # "not this"
	`

	_, err := Parse("test.yaml", strings.NewReader(content))
	if err != nil {
		t.Fail()
	}
}

func TestParseFail(t *testing.T) {
	cases := []string{
		`Fill "x" "something something here" "extra"`,
		`InvalidAction "#x" "something something here"`,
		`Fill`,
		`Fill "" "something something here" "something else"`,
		`Fill some element with whitespace "something something here"`,
		`Fill .input[type=\"text\"] "something something here"`,
		`Click 'single quote' "something"`,
		`Click GetByRole( \"existent\" )`,
		`Click    "existent"   "this should be accepted" "not this"`,
	}

	for _, content := range cases {
		_, err := Parse("test.yaml", strings.NewReader(content))
		if err == nil {
			t.Errorf("expected case to fail with content: %s", content)
		}
	}
}

func TestParseMetadata(t *testing.T) {
	cases := []string{
		`ID            : "Some random ID"
---
Goto "https://nndi.cloud"
`,
		`ID            : "Some random ID without separator"

Goto "https://nndi.cloud"
`,
		`ID            : "Some random ID"
URL           : "https://nndi.cloud"
Title         : "Navigate to home on nndi"
Headless      : "yes"
Description   : "Navigates to the NNDI website and clicks the Home link"
Browser       : "chromium"
---
Goto "https://nndi.cloud"
Click "#navbar > ul > li.active > a"
ExpectAttr "data-nav-section" "home"
Screenshot "body" "./test-nndi.png"
`,
	}

	for _, content := range cases {
		_, err := Parse("test.yaml", strings.NewReader(content))
		if err != nil {
			t.Errorf("test failed with error: %v on \n\t%s", err, content)
		}
	}
}
