package playwright

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"

	playwrightgo "github.com/playwright-community/playwright-go"
	"github.com/zikani03/basi"
	"github.com/zikani03/basi/core"
)

const Name = "playwright"

type Executor struct {
	Name        string           `json:"name,omitempty" yaml:"name,omitempty"`
	Description string           `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string           `json:"url" yaml:"url"`
	Browser     string           `json:"browser" yaml:"browser"`
	Device      string           `json:"device" yaml:"device"`
	Actions     []ExecutorAction `json:"actions" yaml:"actions"`
	Headless    bool             `json:"headless" yaml:"headless"`
	Assertions  []core.Assertion `json:"assertions,omitempty" yaml:"assertions,omitempty"`
}

type ExecutorAction struct {
	Action   string `json:"action"`                           // The action to perform, must be a valid/supported action
	Selector string `json:"selector" yaml:"selector"`         // DOM selector or expression
	Content  string `json:"content,omitempty" yaml:"content"` // Content for actions that require it
	Options  any    `json:"options,omitempty" yaml:"options"` // Options applicable to the given action
}

func NewExecutorAction(act *basi.Action) *ExecutorAction {
	args := ""
	if act.Arguments != nil {
		args = act.Arguments.String
	}
	return &ExecutorAction{
		Action:   act.Action,
		Selector: act.Selector.Selector,
		Content:  args,
		Options:  nil,
	}
}

func (a ExecutorAction) String() string {
	return fmt.Sprintf("action: %s selector: %s content: %s, options: %v", a.Action, a.Selector, a.Content, a.Options)
}

func New() *Executor {
	return &Executor{
		Headless: true,
	}
}

type Result struct {
	Page     *Page `json:"page" yaml:"page"`
	Document *Page `json:"document" yaml:"document"` // alias to Page
}

type Page struct {
	Location *url.URL   `json:"location" yaml:"location"`
	Body     string     `json:"body" yaml:"body"`
	Query    *PageQuery `json:"query" yaml:"query"`
	Scripts  []string   `json:"scripts" yaml:"scripts"`
	CSSFiles []string   `json:"css_files" yaml:"css_files"`
}

// PageQuery allows users to assert the page.Body using css selectrors
type PageQuery struct {
}

// ZeroValueResult return an empty implementation of this executor result
func (Executor) ZeroValueResult() interface{} {
	return Result{}
}

// // GetDefaultAssertions return default assertions for type exec
// func (Executor) GetDefaultAssertions() *venom.StepAssertions {
// 	return &venom.StepAssertions{Assertions: []Assertion{"page.body ShouldNotBeEmpty"}}
// }

// Run execute TestStep of type playwright
func (e *Executor) Run(ctx context.Context) (interface{}, error) {
	browsers := make([]string, 0)
	if e.Browser != "" && slices.Contains[[]string, string]([]string{"chromium", "firefox"}, e.Browser) {
		browsers = append(browsers, e.Browser)
	} else {
		browsers = append(browsers, "chromium")
	}
	err := playwrightgo.Install(&playwrightgo.RunOptions{
		Browsers: browsers,
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch playwright: %w", err)
	}

	pw, err := playwrightgo.Run()
	if err != nil {
		return nil, fmt.Errorf("could not launch playwright: %w", err)
	}
	browser, err := pw.Chromium.Launch(playwrightgo.BrowserTypeLaunchOptions{
		Headless: playwrightgo.Bool(e.Headless), // should we expose this option?
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch Chromium: %w", err)
	}
	context, err := browser.NewContext()
	if err != nil {
		return nil, fmt.Errorf("could not create context: %w", err)
	}
	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	if e.URL != "" {
		_, err = page.Goto(e.URL)
		if err != nil {
			return nil, fmt.Errorf("could not goto: %w", err)
		}
	}

	err = performActions(ctx, page, e.Actions)
	if err != nil {
		return nil, err
	}

	pageBodyBytes, err := page.Content()
	if err != nil {
		return nil, fmt.Errorf("could not goto: %w", err)
	}

	err = browser.Close()
	if err != nil {
		return nil, fmt.Errorf("could not close browser: %w", err)
	}
	err = pw.Stop()
	if err != nil {
		return nil, fmt.Errorf("could not stop Playwright: %w", err)
	}

	pageURL, err := url.Parse(page.URL())
	if err != nil {
		slog.Debug("failed to parse page URL from *playwright.Page object", "error", err)
	}
	pageResult := &Page{
		Location: pageURL,
		Body:     string(pageBodyBytes),
		Query:    nil,
	}

	return Result{
		Page:     pageResult,
		Document: pageResult,
	}, nil
}

func performActions(ctx context.Context, page playwrightgo.Page, actions []ExecutorAction) error {
	assertions := playwrightgo.NewPlaywrightAssertions()
	var lastLocator playwrightgo.Locator
	for i, action := range actions {
		if action.Action == "" {
			return fmt.Errorf("action cannot be empty, please specify an action")
		}
		if action.Selector == "" {
			return fmt.Errorf("selector cannot be empty, please specify a selector")
		}

		actionName := action.Action

		if strings.HasPrefix(actionName, "Expect") {
			// we need to perform an assertion
			if i == 0 {
				return fmt.Errorf("cannot start with an Assertion")
			}
			prev := actions[i-1]

			locator := lastLocator
			if strings.HasPrefix(prev.Action, "Expect") == false {
				locator = page.Locator(prev.Selector)
			}
			if locator == nil {
				return fmt.Errorf("cannot perform assertion without a locator / selector")
			}
			err := performAssertion(assertions, locator, &action)
			if err != nil {
				return err
			}
			lastLocator = locator
		}

		actionFunc, ok := actionMap[actionName]
		if !ok {
			return fmt.Errorf("invalid or unsupported action: '%s'", actionName)
		}

		slog.Debug(fmt.Sprintf("performing action '%s'", action))

		var actErr error
		if len(action.Content) <= 1 {
			actErr = actionFunc(page, &action)
		} else {
			actErr = actionFunc(page, &action)
		}
		if actErr != nil {
			return actErr
		}
	}
	return nil
}
