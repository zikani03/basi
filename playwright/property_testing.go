package playwright

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"

	playwrightgo "github.com/playwright-community/playwright-go"
)

// InvariantRegistry holds global invariants that must be checked after every mutating action
type InvariantRegistry struct {
	Invariants []Invariant
}

// Invariant represents a global assertion
type Invariant struct {
	Action   string
	Selector string
	Content  string
}

// VariableStore holds extracted variables during test execution
type VariableStore struct {
	Variables map[string]interface{}
}

// NewVariableStore creates a new variable store
func NewVariableStore() *VariableStore {
	return &VariableStore{
		Variables: make(map[string]interface{}),
	}
}

// Set sets a variable value
func (vs *VariableStore) Set(name string, value interface{}) {
	vs.Variables[name] = value
}

// Get gets a variable value
func (vs *VariableStore) Get(name string) interface{} {
	return vs.Variables[name]
}

// InterpolateVariables replaces variable references in strings
func (vs *VariableStore) InterpolateVariables(input string) string {
	re := regexp.MustCompile(`\$[a-zA-Z][a-zA-Z_\d]*`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		varName := strings.TrimPrefix(match, "$")
		if value := vs.Get(varName); value != nil {
			return fmt.Sprintf("%v", value)
		}
		return match // Keep original if not found
	})
}

// ExecutionContext holds the state during test execution
type ExecutionContext struct {
	Invariants    *InvariantRegistry
	Variables     *VariableStore
	CurrentOrigin string
}

// NewExecutionContext creates a new execution context
func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		Invariants: &InvariantRegistry{},
		Variables:  NewVariableStore(),
	}
}

// IsMutatingAction checks if an action modifies state
func IsMutatingAction(action string) bool {
	mutatingActions := []string{
		"Click", "DoubleClick", "Doubleclick", "Tap", "Clear", "Fill",
		"Check", "Uncheck", "FillCheckbox", "Press", "PressSequentially",
		"Select", "SelectOption", "SelectMultipleOptions", "Type",
		"Goto", "GoBack", "GoForward", "Refresh",
	}
	for _, ma := range mutatingActions {
		if action == ma {
			return true
		}
	}
	return false
}

// CheckInvariants runs all registered invariants
func (ec *ExecutionContext) CheckInvariants(page playwrightgo.Page, assertions playwrightgo.PlaywrightAssertions) error {
	for _, invariant := range ec.Invariants.Invariants {
		action := ExecutorAction{
			Action:   invariant.Action,
			Selector: invariant.Selector,
			Content:  ec.Variables.InterpolateVariables(invariant.Content),
		}
		err := performAssertion(assertions, page.Locator(invariant.Selector), &action)
		if err != nil {
			return fmt.Errorf("invariant failed: %s %s %s: %w", invariant.Action, invariant.Selector, invariant.Content, err)
		}
	}
	return nil
}

// performFuzz executes bounded random interactions
func performFuzz(ctx context.Context, page playwrightgo.Page, action *ExecutorAction, execCtx *ExecutionContext) error {
	stepCount := action.Number
	if stepCount <= 0 {
		stepCount = 10 // default
	}

	scopeSelector := action.Selector
	ignoreSelector := action.Content

	// Create a timeout context for the entire fuzz operation
	fuzzCtx, cancel := context.WithTimeout(ctx, 30*time.Second) // 30 second max
	defer cancel()

	for i := 0; i < stepCount; i++ {
		select {
		case <-fuzzCtx.Done():
			return fmt.Errorf("fuzz operation timed out")
		default:
		}

		// Find actionable elements within scope
		scopeLocator := page.Locator(scopeSelector)
		actionableElements := scopeLocator.Locator("input, button, a, select, textarea")

		// Exclude ignored elements
		if ignoreSelector != "" {
			actionableElements = actionableElements.Locator(fmt.Sprintf(":not(%s)", ignoreSelector))
		}

		count, err := actionableElements.Count()
		if err != nil || count == 0 {
			continue // No actionable elements found
		}

		// Select random element
		randomIndex := rand.Intn(count)
		randomElement := actionableElements.Nth(randomIndex)

		// Check if element is visible and enabled
		isVisible, err := randomElement.IsVisible()
		if err != nil || !isVisible {
			continue
		}

		isEnabled, err := randomElement.IsEnabled()
		if err != nil || !isEnabled {
			continue
		}

		// Determine element type and perform appropriate action
		tagName, err := randomElement.Evaluate("el => el.tagName.toLowerCase()", nil)
		if err != nil {
			continue
		}
		tagNameStr, ok := tagName.(string)
		if !ok {
			continue
		}

		switch tagNameStr {
		case "input":
			inputType, _ := randomElement.GetAttribute("type")
			switch inputType {
			case "checkbox", "radio":
				randomElement.Check()
			case "text", "password", "email", "search", "tel", "url":
				randomText := generateRandomString(10)
				randomElement.Fill(randomText)
			default:
				randomElement.Click()
			}
		case "button", "a":
			randomElement.Click()
		case "select":
			randomElement.SelectOption(playwrightgo.SelectOptionValues{Values: &[]string{""}})
		case "textarea":
			randomText := generateRandomString(20)
			randomElement.Fill(randomText)
		}

		// Check invariants after each random action
		assertions := playwrightgo.NewPlaywrightAssertions()
		err = execCtx.CheckInvariants(page, assertions)
		if err != nil {
			return fmt.Errorf("invariant failed during fuzz step %d: %w", i+1, err)
		}

		// Check if we left the origin (security boundary)
		currentURL := page.URL()
		if currentURL != "" {
			if u, err := url.Parse(currentURL); err == nil {
				currentOrigin := u.Scheme + "://" + u.Host
				if currentOrigin != execCtx.CurrentOrigin {
					// Navigate back to original origin
					page.GoBack()
					break
				}
			}
		}
	}

	return nil
}

// performEventually polls until a condition is met
func performEventually(ctx context.Context, page playwrightgo.Page, action *ExecutorAction, execCtx *ExecutionContext, assertions playwrightgo.PlaywrightAssertions) error {
	eventuallyCtx, cancel := context.WithTimeout(ctx, 5*time.Second) // 5 second timeout
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-eventuallyCtx.Done():
			return fmt.Errorf("eventually condition not met within timeout")
		case <-ticker.C:
			// Interpolate variables in the action content
			interpolatedContent := execCtx.Variables.InterpolateVariables(action.Content)

			// Create a temporary action with interpolated content
			tempAction := *action
			tempAction.Content = interpolatedContent

			locator := page.Locator(action.Selector)
			err := performAssertion(assertions, locator, &tempAction)
			if err == nil {
				return nil // Condition met
			}
		}
	}
}

// performNext asserts the immediate next state
func performNext(page playwrightgo.Page, action *ExecutorAction, execCtx *ExecutionContext, assertions playwrightgo.PlaywrightAssertions) error {
	// Interpolate variables in the action content
	interpolatedContent := execCtx.Variables.InterpolateVariables(action.Content)

	// Create a temporary action with interpolated content
	tempAction := *action
	tempAction.Content = interpolatedContent

	locator := page.Locator(action.Selector)
	return performAssertion(assertions, locator, &tempAction)
}

// generateRandomString creates a random string of given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
