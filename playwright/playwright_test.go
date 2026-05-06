package playwright

import (
	"context"
	"fmt"
	"os"
	"testing"

	playwrightgo "github.com/playwright-community/playwright-go"
)

const testPage = `
<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Test Page</title>
</head>

<body>
    <form method="post" action="https://example.com/submit" id="example-form">
        <h1>Example form</h1>
        <div class="control"><input type="text" name="firstName" id="firstName"></div>
        <div class="control"><input type="text" name="lastName" id="lastName"></div>
        <div class="control"><input type="text" name="age" id="age"></div>
        <div class="control"><input type="email" name="email" id="email"></div>
        <div class="control"><textarea name="bio" id="biography"></textarea></div>
        <button id="submit-button" type="submit">Submit</button>
    </form>
    <div id="age-shown-when-input" style="display: none;">
        <h4>AGE</h4>
        <span id="inputted-age"></span>
    </div>

    <script>
        document.addEventListener('DOMContentLoaded', function () {

            const el = document.getElementById('age');
            age.addEventListener('input', function () {
                const ageValue = el.value;
                const ageShown = document.getElementById('inputted-age');
                ageShown.textContent = ageValue;
                ageShown.parentElement.style.display = ageValue ? 'block' : 'none';
            });
        });
    </script>
</body>
</html>
`

func TestPerformActions(t *testing.T) {

	testActions := []ExecutorAction{
		{Action: "Fill", Selector: "#firstName", Content: "John"},
		{Action: "Fill", Selector: "[name=lastName]", Content: "John"},
		{Action: "Fill", Selector: "#age", Content: "24"},
		{Action: "Focus", Selector: "#email"},
		{Action: "WaitFor", Selector: "#inputted-age", Content: "24"},
		{Action: "Click", Selector: "#submit-button"},
	}

	pw, err := playwrightgo.Run()
	if err != nil {
		t.Fail()
	}
	browser, err := pw.Chromium.Launch(playwrightgo.BrowserTypeLaunchOptions{
		Headless: playwrightgo.Bool(true),
	})
	if err != nil {
		t.Fail()
	}
	browserCtx, err := browser.NewContext()
	if err != nil {
		t.Fail()
	}
	page, err := browserCtx.NewPage()
	if err != nil {
		t.Fail()
	}

	err = page.SetContent(testPage, playwrightgo.PageSetContentOptions{})
	if err != nil {
		t.Error("failed to set testPage content")
	}

	err = performActions(context.Background(), New(), page, testActions, NewExecutionContext())
	if err != nil {
		t.Errorf("failed to test actions %v", err)
	}

	t.Cleanup(func() {
		err = browser.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to close browser properly %v", err)
		}
		err = pw.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to close browser properly %v", err)
		}
	})
}

func TestPerformFindOnEmptyElement(t *testing.T) {

	testActions := []ExecutorAction{
		{Action: "Find", Selector: "does-not-exist", Content: "does-not-exist"},
	}

	pw, err := playwrightgo.Run()
	if err != nil {
		t.Fail()
	}
	browser, err := pw.Chromium.Launch(playwrightgo.BrowserTypeLaunchOptions{
		Headless: playwrightgo.Bool(true),
	})
	if err != nil {
		t.Fail()
	}
	browserCtx, err := browser.NewContext()
	if err != nil {
		t.Fail()
	}
	page, err := browserCtx.NewPage()
	if err != nil {
		t.Fail()
	}

	err = page.SetContent(testPage, playwrightgo.PageSetContentOptions{})
	if err != nil {
		t.Error("failed to set testPage content")
	}

	err = performActions(context.Background(), New(), page, testActions, NewExecutionContext())
	if err == nil {
		t.Errorf("expected error while performing actions")
	}

	t.Cleanup(func() {
		err = browser.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to close browser properly %v", err)
		}
		err = pw.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to close browser properly %v", err)
		}
	})
}

func TestGenerateCAT(t *testing.T) {
	pw, err := playwrightgo.Run()
	if err != nil {
		t.Fatalf("could not start playwright: %v", err)
	}
	t.Cleanup(func() { pw.Stop() })

	browser, err := pw.Chromium.Launch(playwrightgo.BrowserTypeLaunchOptions{
		Headless: playwrightgo.Bool(true),
	})
	if err != nil {
		t.Fatalf("could not launch browser: %v", err)
	}
	t.Cleanup(func() { browser.Close() })

	context, err := browser.NewContext()
	if err != nil {
		t.Fatalf("could not create context: %v", err)
	}
	page, err := context.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "Basic structure with landmarks and interactive elements",
			html: `
				<nav><button id="b1" aria-label="Login">Sign In</button></nav>
				<main>
					<div class="ignore-me">
						<input name="username" type="text" placeholder="User">
					</div>
				</main>
			`,
			expected: `nav>button#b1[aria-label="Login"]+main>input[name="username"][type="text"][placeholder="User"]`,
		},
		{
			name: "Pruning of non-visual tags",
			html: `
				<form>
					<script>alert(1)</script>
					<style>.css { color: blue; }</style>
					<svg><path d="M10 10"/></svg>
					<button type="submit">Send</button>
				</form>
			`,
			expected: `form>button[type="submit"]`,
		},
		{
			name: "Nested interesting elements",
			html: `
				<section title="Container">
					<article>
						<a href="/post/1" title="Read more">Title</a>
					</article>
				</section>
			`,
			expected: `section[title="Container"]>article>a[title="Read more"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := page.SetContent(tt.html, playwrightgo.PageSetContentOptions{})
			if err != nil {
				t.Fatalf("failed to set content: %v", err)
			}

			cat, err := GenerateCAT(page)
			if err != nil {
				t.Fatalf("GenerateCAT failed: %v", err)
			}

			if cat != tt.expected {
				t.Errorf("\nexpected: %s\ngot:      %s", tt.expected, cat)
			}
		})
	}
}
