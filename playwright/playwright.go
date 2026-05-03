package playwright

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strconv"
	"strings"

	playwrightgo "github.com/playwright-community/playwright-go"
	"github.com/zikani03/basi"
)

const Name = "playwright"

type Executor struct {
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string            `json:"url" yaml:"url"`
	Browser     string            `json:"browser" yaml:"browser"`
	Device      string            `json:"device" yaml:"device"`
	Actions     []ExecutorAction  `json:"actions" yaml:"actions"`
	Headless    bool              `json:"headless" yaml:"headless"`
	Context     *ExecutionContext `json:"-" yaml:"-"` // Execution context for property-based testing
}

type ExecutorAction struct {
	Action   string `json:"action"`                             // The action to perform, must be a valid/supported action
	Selector string `json:"selector" yaml:"selector"`           // DOM selector or expression
	Content  string `json:"content,omitempty" yaml:"content"`   // Content for actions that require it
	Variable string `json:"variable,omitempty" yaml:"variable"` // Variable name for Extract
	Number   int    `json:"number,omitempty" yaml:"number"`     // Number for Fuzz step count
	Options  any    `json:"options,omitempty" yaml:"options"`   // Options applicable to the given action
}

// Target represents a locator target which can be either a semantic intent
// or a pinned (deterministic) selector.
type Target struct {
	Raw      string
	IsPinned bool
}

func NewTarget(raw string) Target {
	isPinned := strings.HasPrefix(raw, "@")
	return Target{
		Raw:      strings.TrimPrefix(raw, "@"),
		IsPinned: isPinned,
	}
}

func NewExecutorAction(act *basi.Action) *ExecutorAction {
	args := ""
	if act.Arguments != nil {
		args = act.Arguments.String
	}
	variable := ""
	if act.Variable != nil {
		variable = act.Variable.Variable
	}
	number := 0
	if act.Number != nil {
		number = act.Number.Number
	}
	return &ExecutorAction{
		Action:   act.Action,
		Selector: act.Selector.Selector,
		Content:  args,
		Variable: variable,
		Number:   number,
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

	if e.Name != "" {
		fmt.Printf("Running: \033[10;1;1m%s\033[0m on\033[32;1;4m(%s)\033[0m\n", e.Name, e.Browser)
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
		// Set current origin for fuzzing boundary checks
		if e.Context == nil {
			e.Context = NewExecutionContext()
		}
		if currentURL := page.URL(); currentURL != "" {
			if u, err := url.Parse(currentURL); err == nil {
				e.Context.CurrentOrigin = u.Scheme + "://" + u.Host
			}
		}
	}

	err = performActions(ctx, page, e.Actions, e.Context)
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

// Resolve evaluates a Target against the current page state.
func Resolve(ctx context.Context, t Target, page playwrightgo.Page) (string, error) {
	if t.IsPinned {
		return t.Raw, nil
	}
	cat, err := GenerateCAT(page)
	if err != nil {
		return "", fmt.Errorf("failed to generate CAT: %w", err)
	}
	slog.Debug("generated CAT", "cat", cat)

	// TODO: Phase 3 - Resolution via Tiered Brain (Local SLM/Cloud VLM)
	return t.Raw, nil
}

func performActions(ctx context.Context, page playwrightgo.Page, actions []ExecutorAction, execCtx *ExecutionContext) error {
	if execCtx == nil {
		execCtx = NewExecutionContext()
	}
	assertions := playwrightgo.NewPlaywrightAssertions()
	var lastLocator playwrightgo.Locator
	for i := range actions {
		action := &actions[i]
		if action.Action == "" {
			return fmt.Errorf("action cannot be empty, please specify an action")
		}

		actionName := action.Action

		// Handle Always action - register invariant
		if actionName == "Always" {
			// The embedded expect action is in the Selector field
			invariant := Invariant{
				Action:   action.Selector,
				Selector: "",
				Content:  "",
			}
			execCtx.Invariants.Invariants = append(execCtx.Invariants.Invariants, invariant)
			continue
		}

		// Handle Extract action
		if actionName == "Extract" {
			target := NewTarget(action.Content)
			resolved, err := Resolve(ctx, target, page)
			if err != nil {
				return fmt.Errorf("failed to resolve extract target: %w", err)
			}
			locator := page.Locator(resolved)
			textContent, err := locator.TextContent()
			if err != nil {
				return fmt.Errorf("failed to extract text from selector %s: %w", action.Content, err)
			}
			execCtx.Variables.Set(action.Selector, textContent)
			continue
		}

		// Handle Eventually action
		if actionName == "Eventually" {
			// The embedded expect action is in the Selector field
			embeddedActionObj := ExecutorAction{
				Action:   action.Selector,
				Selector: "",
				Content:  "",
			}
			err := performEventually(ctx, page, &embeddedActionObj, execCtx, assertions)
			if err != nil {
				return fmt.Errorf("eventually action failed: %w", err)
			}
			continue
		}

		// Handle Next action
		if actionName == "Next" {
			// The embedded expect action is in the Selector field
			embeddedActionObj := ExecutorAction{
				Action:   action.Selector,
				Selector: "",
				Content:  "",
			}
			err := performNext(page, &embeddedActionObj, execCtx, assertions)
			if err != nil {
				return fmt.Errorf("next action failed: %w", err)
			}
			continue
		}

		if action.Selector == "" && actionName != "Extract" && actionName != "Fuzz" {
			return fmt.Errorf("selector cannot be empty for action %s, please specify a selector", actionName)
		}

		// Determine if the action uses a target selector that needs resolution
		isTarget := true
		nonTargetActions := map[string]bool{
			"Always": true, "Eventually": true, "Next": true, "Extract": true, "Fuzz": true,
			"Goto": true, "WaitForURL": true, "Screenshot": true,
			"Refresh": true, "GoBack": true, "GoForward": true,
		}
		if nonTargetActions[actionName] {
			isTarget = false
		}

		if isTarget && action.Selector != "" {
			target := NewTarget(action.Selector)
			resolved, err := Resolve(ctx, target, page)
			if err != nil {
				return fmt.Errorf("failed to resolve selector for action %s: %w", actionName, err)
			}
			action.Selector = resolved
		}

		if strings.HasPrefix(actionName, "Find") {
			loc, err := tryFindLocator(page, *action)
			if err != nil {
				return fmt.Errorf("failed to find a element on the page using: '%s'", cmp.Or(action.Selector, action.Content))
			}
			lastLocator = loc
			numMatched, err := lastLocator.Count()
			if numMatched <= 0 || err != nil {
				return fmt.Errorf("failed to find a element on the page using: '%s'", cmp.Or(action.Selector, action.Content))
			}
			continue
		}

		// Handle ExpectEach action
		if actionName == "ExpectEach" {
			prev := actions[i-1]

			locator := lastLocator
			if !strings.HasPrefix(prev.Action, "Expect") && !strings.HasPrefix(prev.Action, "Find") {
				locator = page.Locator(prev.Selector)
			}
			if locator == nil {
				return fmt.Errorf("cannot perform assertion without a locator / selector")
			}
			err := performAssertion(assertions, locator, action)
			if err != nil {
				return err
			}
			lastLocator = locator
			continue
		}

		actionFunc, ok := actionMap[actionName]
		if !ok {
			return fmt.Errorf("invalid or unsupported action: '%s'", actionName)
		}

		slog.Debug(fmt.Sprintf("performing action '%s'", *action))

		actErr := actionFunc(page, action)
		if actErr != nil {
			return actErr
		}

		// Check invariants after mutating actions
		if IsMutatingAction(actionName) {
			err := execCtx.CheckInvariants(page, assertions)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func tryFindLocator(page playwrightgo.Page, action ExecutorAction) (playwrightgo.Locator, error) {
	selectorOrContent := cmp.Or(action.Selector, action.Content)
	var loc playwrightgo.Locator
	type SelectorFunc func(string) playwrightgo.Locator
	selectorFuncs := []SelectorFunc{
		func(s string) playwrightgo.Locator { return page.GetByText(s) },
		func(s string) playwrightgo.Locator { return page.GetByPlaceholder(s) },
		func(s string) playwrightgo.Locator { return page.GetByLabel(s) },
		func(s string) playwrightgo.Locator { return page.GetByAltText(s) },
		func(s string) playwrightgo.Locator { return page.Locator(s) },
	}

	for _, f := range selectorFuncs {
		loc = f(selectorOrContent)
		if loc == nil {
			return nil, fmt.Errorf("could not find element using %s", selectorOrContent)
		}
	}

	switch action.Action {
	case "FindNth":
		nth, err := strconv.Atoi(action.Content)
		if err != nil {
			return nil, fmt.Errorf(`the parameter N must be a number for FindNth e.g. FindNth "%s" "5"`, selectorOrContent)
		}
		return loc.Nth(nth), nil
	case "FindMatching", "FindRegex":
		notExact := false
		pattern := action.Content
		// check if the content is a regular expression or make it into one
		if !strings.HasSuffix(pattern, "/") && strings.HasPrefix(pattern, "/") {
			pattern = "/" + action.Content + "/"
		}
		loc = page.GetByText(pattern, playwrightgo.PageGetByTextOptions{
			Exact: &notExact,
		})
		return loc.First(), nil
	case "FindFirst":
		return loc.First(), nil
	case "FindLast":
		return loc.Last(), nil
	}
	if loc == nil {
		return nil, fmt.Errorf("could not find element using %s", selectorOrContent)
	}
	return loc, nil
}
