package playwright

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	playwrightgo "github.com/mxschmitt/playwright-go"
	"github.com/zikani03/basi"
	"github.com/zikani03/basi/core"
)

const Name = "playwright"

type Executor struct {
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string            `json:"url" yaml:"url"`
	CDPEndpoint string            `json:"cdpEndpoint,omitempty" yaml:"cdpEndpoint,omitempty"`
	Browser     string            `json:"browser" yaml:"browser"`
	Device      string            `json:"device" yaml:"device"`
	Actions     []ExecutorAction  `json:"actions" yaml:"actions"`
	Headless    bool              `json:"headless" yaml:"headless"`
	Context     *ExecutionContext `json:"-" yaml:"-"` // Execution context for property-based testing
	Emitter     core.Emitter     `json:"-" yaml:"-"` // Progress emitter; nil means no-op
	RunNumber   int              `json:"-" yaml:"-"` // Current run index (1-based) for repeated runs
}

type ExecutorAction struct {
	Action   string  `json:"action"`                                     // The action to perform, must be a valid/supported action
	Selector string  `json:"selector" yaml:"selector"`                   // DOM selector or expression
	Content  string  `json:"content,omitempty" yaml:"content"`           // Content for actions that require it
	Variable string  `json:"variable,omitempty" yaml:"variable"`         // Variable name for Extract
	Number   int     `json:"number,omitempty" yaml:"number"`             // Number for Fuzz step count
	Options  any     `json:"options,omitempty" yaml:"options"`           // Options applicable to the given action
	Timeout  float64 `json:"timeout,omitempty" yaml:"timeout,omitempty"` // timeout
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
	Text     string     `json:"text" yaml:"text"` // plain-text content of the page body
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
	emitter := e.Emitter
	if emitter == nil {
		emitter = core.NoopEmitter{}
	}
	runNumber := e.RunNumber
	if runNumber < 1 {
		runNumber = 1
	}

	browsers := make([]string, 0)
	if e.Browser != "" && slices.Contains([]string{"chromium", "firefox"}, e.Browser) {
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

	var browser playwrightgo.Browser
	var browserEngine playwrightgo.BrowserType

	if e.Browser == "firefox" {
		browserEngine = pw.Firefox
	} else { // defaults to chromium on any other case
		browserEngine = pw.Chromium
	}

	if e.CDPEndpoint != "" {
		// In order to support platforms like cloudflare, we need to allow users to be able to pass
		// options like headers to the CDP options. The best way is probably to expose this somehowm but...
		cdpHeadersFromEnv := map[string]string{}
		headersJSON := os.Getenv("CDP_HEADERS")
		if headersJSON != "" {
			if err := json.Unmarshal([]byte(headersJSON), &cdpHeadersFromEnv); err != nil {
				return nil, fmt.Errorf("failed to parse CDP_HEADERS: %v", err)
			}
		}

		browser, err = browserEngine.ConnectOverCDP(e.CDPEndpoint, playwrightgo.BrowserTypeConnectOverCDPOptions{
			Headers: cdpHeadersFromEnv,
		})
	} else {
		browser, err = browserEngine.Launch(playwrightgo.BrowserTypeLaunchOptions{
			Headless: playwrightgo.Bool(e.Headless),
		})
	}

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

	err = performActions(ctx, page, e.Actions, e.Context, emitter, runNumber)
	if err != nil {
		return nil, err
	}

	pageBodyBytes, err := page.Content()
	if err != nil {
		return nil, fmt.Errorf("could not get page content: %w", err)
	}
	pageText, err := page.Locator("body").InnerText()
	if err != nil {
		pageText = "" // non-fatal; some pages may not have a body
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
		Text:     pageText,
		Query:    nil,
	}

	return Result{
		Page:     pageResult,
		Document: pageResult,
	}, nil
}

func performActions(ctx context.Context, page playwrightgo.Page, actions []ExecutorAction, execCtx *ExecutionContext, emitter core.Emitter, runNumber int) error {
	if execCtx == nil {
		execCtx = NewExecutionContext()
	}
	assertions := playwrightgo.NewPlaywrightAssertions()
	var lastLocator playwrightgo.Locator
	for i, action := range actions {
		start := time.Now()
		emitter.Emit(core.ProgressEvent{
			Step:      i + 1,
			Run:       runNumber,
			Action:    action.Action,
			Selector:  action.Selector,
			Status:    core.StepRunning,
			Timestamp: start.Format(time.RFC3339),
		})

		err := dispatchSingleAction(ctx, page, i, &action, actions, &lastLocator, execCtx, assertions)

		elapsed := time.Since(start)
		if err != nil {
			emitter.Emit(core.ProgressEvent{
				Step:       i + 1,
				Run:        runNumber,
				Action:     action.Action,
				Selector:   action.Selector,
				Status:     core.StepFail,
				Message:    err.Error(),
				Timestamp:  time.Now().Format(time.RFC3339),
				DurationMs: elapsed.Milliseconds(),
			})
			return err
		}
		emitter.Emit(core.ProgressEvent{
			Step:       i + 1,
			Run:        runNumber,
			Action:     action.Action,
			Selector:   action.Selector,
			Status:     core.StepOK,
			Timestamp:  time.Now().Format(time.RFC3339),
			DurationMs: elapsed.Milliseconds(),
		})
	}
	return nil
}

func dispatchSingleAction(ctx context.Context, page playwrightgo.Page, i int, action *ExecutorAction, actions []ExecutorAction, lastLocator *playwrightgo.Locator, execCtx *ExecutionContext, assertions playwrightgo.PlaywrightAssertions) error {
	if action.Action == "" {
		return fmt.Errorf("action cannot be empty, please specify an action")
	}

	actionName := action.Action

	// Handle Always action - register invariant
	if actionName == "Always" {
		invariant := Invariant{
			Action:   action.Selector,
			Selector: "",
			Content:  "",
		}
		execCtx.Invariants.Invariants = append(execCtx.Invariants.Invariants, invariant)
		return nil
	}

	// Handle Extract action
	if actionName == "Extract" {
		locator := page.Locator(action.Content)
		textContent, err := locator.TextContent()
		if err != nil {
			return fmt.Errorf("failed to extract text from selector %s: %w", action.Content, err)
		}
		execCtx.Variables.Set(action.Selector, textContent)
		return nil
	}

	// Handle Fuzz action
	if actionName == "Fuzz" {
		parts := strings.Fields(action.Selector)
		if len(parts) < 1 {
			return fmt.Errorf("Fuzz action requires at least a step count")
		}
		stepCount := 10
		if parsedCount, err := strconv.Atoi(parts[0]); err == nil {
			stepCount = parsedCount
		}
		scopeSelector := ".body"
		if len(parts) >= 2 {
			scopeSelector = parts[1]
		}
		ignoreSelector := ""
		if len(parts) >= 3 {
			ignoreSelector = parts[2]
		}

		fuzzAction := ExecutorAction{
			Action:   "Fuzz",
			Selector: scopeSelector,
			Content:  ignoreSelector,
			Number:   stepCount,
		}
		if err := performFuzz(ctx, page, &fuzzAction, execCtx); err != nil {
			return fmt.Errorf("fuzz action failed: %w", err)
		}
		return nil
	}

	// Handle Eventually action
	if actionName == "Eventually" {
		embeddedActionObj := ExecutorAction{
			Action:   action.Selector,
			Selector: "",
			Content:  "",
		}
		if err := performEventually(ctx, page, &embeddedActionObj, execCtx, assertions); err != nil {
			return fmt.Errorf("eventually action failed: %w", err)
		}
		return nil
	}

	// Handle Next action
	if actionName == "Next" {
		embeddedActionObj := ExecutorAction{
			Action:   action.Selector,
			Selector: "",
			Content:  "",
		}
		if err := performNext(page, &embeddedActionObj, execCtx, assertions); err != nil {
			return fmt.Errorf("next action failed: %w", err)
		}
		return nil
	}

	if action.Selector == "" && actionName != "Extract" && actionName != "Fuzz" {
		return fmt.Errorf("selector cannot be empty for action %s, please specify a selector", actionName)
	}

	if strings.HasPrefix(actionName, "Find") {
		loc, err := tryFindLocator(page, *action)
		if err != nil {
			return fmt.Errorf("failed to find a element on the page using: '%s'", cmp.Or(action.Selector, action.Content))
		}
		*lastLocator = loc
		numMatched, err := (*lastLocator).Count()
		if numMatched <= 0 || err != nil {
			return fmt.Errorf("failed to find a element on the page using: '%s'", cmp.Or(action.Selector, action.Content))
		}
		return nil
	}

	if strings.HasPrefix(actionName, "Expect") {
		if i == 0 {
			return fmt.Errorf("cannot start with an Assertion")
		}
		prev := actions[i-1]

		locator := *lastLocator
		if !strings.HasPrefix(prev.Action, "Expect") && !strings.HasPrefix(prev.Action, "Find") {
			locator = page.Locator(prev.Selector)
		}
		if locator == nil {
			return fmt.Errorf("cannot perform assertion without a locator / selector")
		}
		if err := performAssertion(assertions, locator, action); err != nil {
			return err
		}
		*lastLocator = locator
	}

	actionFunc, ok := actionMap[actionName]
	if !ok {
		return fmt.Errorf("invalid or unsupported action: '%s'", actionName)
	}

	slog.Debug(fmt.Sprintf("performing action '%s'", action))

	if actErr := actionFunc(page, action); actErr != nil {
		return actErr
	}

	// Check invariants after mutating actions
	if IsMutatingAction(actionName) {
		if err := execCtx.CheckInvariants(page, assertions); err != nil {
			return err
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
