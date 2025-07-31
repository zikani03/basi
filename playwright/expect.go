package playwright

import (
	"fmt"
	"strings"

	playwrightgo "github.com/playwright-community/playwright-go"
)

func StubExpectAction(page playwrightgo.Page, action *ExecutorAction) error {
	return nil
}

type AssertionFunc func(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error

func performAssertion(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction) error {
	assertionKind := strings.ToLower(strings.Replace(action.Action, "Expect", "", 1))
	switch assertionKind {
	case "":
		fallthrough
	case "text":
		return ExpectText(assertions, locator, action)
	case "value":
		return ExpectValue(assertions, locator, action)
	case "values":
		return ExpectValues(assertions, locator, action)
	case "attr":
		return ExpectAttribute(assertions, locator, action)
	case "attribute":
		return ExpectAttribute(assertions, locator, action)
	default:
	}
	return fmt.Errorf("unsupported assertion provided: %s", action.Action)
}

func ExpectText(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveText(action.Content)
}

func ExpectAttribute(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveAttribute(action.Selector, action.Content)
}

func ExpectValue(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveValue(action.Content) // TODO: handle options
}

func ExpectValues(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveValues([]any{action.Content}) // TODO: handle options and array
}
