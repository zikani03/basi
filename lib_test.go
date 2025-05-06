package v4playwrightactionparser

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

	_, err := Parse(strings.NewReader(content))
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
		_, err := Parse(strings.NewReader(content))
		if err == nil {
			t.Errorf("expected case to fail with content: %s", content)
		}
	}
}
