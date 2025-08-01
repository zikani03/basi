package playwright

import (
	"cmp"
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
	case "attached":
		return ExpectToBeAttached(assertions, locator, action)
	case "checked":
		return ExpectToBeChecked(assertions, locator, action)
	case "disabled":
		return ExpectToBeDisabled(assertions, locator, action)
	case "editable":
		return ExpectToBeEditable(assertions, locator, action)
	case "empty":
		return ExpectToBeEmpty(assertions, locator, action)
	case "enabled":
		return ExpectToBeEnabled(assertions, locator, action)
	case "focused":
		return ExpectToBeFocused(assertions, locator, action)
	case "hidden":
		return ExpectToBeHidden(assertions, locator, action)
	case "inviewport":
		return ExpectToBeInViewport(assertions, locator, action)
	case "visible":
		return ExpectToBeVisible(assertions, locator, action)
	case "containclass":
		return ExpectToContainClass(assertions, locator, action)
	case "containtext":
		return ExpectToContainText(assertions, locator, action)
	case "accessibledescription":
		return ExpectToHaveAccessibleDescription(assertions, locator, action)
	case "accessibleerrormessage":
		return ExpectToHaveAccessibleErrorMessage(assertions, locator, action)
	case "accessiblename":
		return ExpectToHaveAccessibleName(assertions, locator, action)
	case "attribute":
		return ExpectToHaveAttribute(assertions, locator, action)
	case "class":
		return ExpectToHaveClass(assertions, locator, action)
	case "css":
		return ExpectToHaveCSS(assertions, locator, action)
	case "id":
		return ExpectToHaveId(assertions, locator, action)
	case "jsproperty":
		return ExpectToHaveJSProperty(assertions, locator, action)
	// case "count":
	// 	return ExpectToHaveCount(assertions, locator, action)
	// case "role":
	// 	return ExpectToHaveRole(assertions, locator, action)
	case "":
		fallthrough
	case "text":
		return ExpectText(assertions, locator, action)
	case "value":
		return ExpectToHaveValue(assertions, locator, action)
	case "values":
		return ExpectToHaveValues(assertions, locator, action)
	// case "matchariasnapshot":
	// 	return ToMatchAriaSnapshot(assertions, locator, action)
	default:
	}
	return fmt.Errorf("unsupported assertion provided: %s", action.Action)
}

func ExpectText(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveText(action.Content)
}

func ExpectToHaveAttribute(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveAttribute(action.Selector, action.Content)
}

func ExpectToHaveValue(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveValue(action.Content) // TODO: handle options
}

func ExpectToHaveValues(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveValues([]any{action.Content}) // TODO: handle options and array
}

func ExpectToBeAttached(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToBeAttached() // TODO: support options - (options ...LocatorAssertionsToBeAttachedOptions) error
}

func ExpectToBeChecked(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToBeChecked() // TODO: support options - (options ...LocatorAssertionsToBeCheckedOptions) error
}

func ExpectToBeDisabled(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToBeDisabled() // TODO: support options - (options ...LocatorAssertionsToBeDisabledOptions) error
}

func ExpectToBeEditable(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToBeEditable() // TODO: support options - (options ...LocatorAssertionsToBeEditableOptions) error
}

func ExpectToBeEmpty(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToBeEmpty() // TODO: support options - (options ...LocatorAssertionsToBeEmptyOptions) error
}

func ExpectToBeEnabled(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToBeEnabled() // TODO: support options - (options ...LocatorAssertionsToBeEnabledOptions) error
}

func ExpectToBeFocused(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToBeFocused() // TODO: support options - (options ...LocatorAssertionsToBeFocusedOptions) error
}

func ExpectToBeHidden(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToBeHidden() // TODO: support options - (options ...LocatorAssertionsToBeHiddenOptions) error
}

func ExpectToBeInViewport(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToBeInViewport() // TODO: support options - (options ...LocatorAssertionsToBeInViewportOptions) error
}

func ExpectToBeVisible(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToBeVisible() // TODO: support options - (options ...LocatorAssertionsToBeVisibleOptions) error
}

func ExpectToContainClass(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToContainClass(action.Content) // TODO: support options - (expected interface{}, options ...LocatorAssertionsToContainClassOptions) error
}

func ExpectToContainText(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToContainText(action.Content) // TODO: support options - (expected interface{}, options ...LocatorAssertionsToContainTextOptions) error
}

func ExpectToHaveAccessibleDescription(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveAccessibleDescription(action.Content) // TODO: support options - (description interface{}, options ...LocatorAssertionsToHaveAccessibleDescriptionOptions) error
}

func ExpectToHaveAccessibleErrorMessage(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveAccessibleErrorMessage(action.Content) // TODO: support options - (errorMessage interface{}, options ...LocatorAssertionsToHaveAccessibleErrorMessageOptions) error
}

func ExpectToHaveAccessibleName(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveAccessibleName(action.Content) // TODO: support options - (name interface{}, options ...LocatorAssertionsToHaveAccessibleNameOptions) error
}

func ExpectToHaveClass(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveClass(cmp.Or(action.Selector, action.Content)) // TODO: support options - (expected interface{}, options ...LocatorAssertionsToHaveClassOptions) error
}

// TODO: we cannot yet support count because we don't have a Find action
// func ExpectToHaveCount(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
// 	assert := assertions.Locator(locator)
// 	return assert.ToHaveCount() // TODO: support options - (count int, options ...LocatorAssertionsToHaveCountOptions) error
// }

func ExpectToHaveCSS(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveCSS(action.Selector, action.Content) // TODO: support options - (name string, value interface{}, options ...LocatorAssertionsToHaveCSSOptions) error
}

func ExpectToHaveId(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveId(cmp.Or(action.Selector, action.Content)) // TODO: support options - (id interface{}, options ...LocatorAssertionsToHaveIdOptions) error
}

func ExpectToHaveJSProperty(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToHaveJSProperty(action.Selector, action.Content) // TODO: support options - (name string, value interface{}, options ...LocatorAssertionsToHaveJSPropertyOptions) error
}

// TODO: support aria roles
// func ExpectToHaveRole(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
// 	assert := assertions.Locator(locator)
// 	return assert.ToHaveRole(cmp.Or(action.Selector, action.Content)) // TODO: support options - (role AriaRole, options ...LocatorAssertionsToHaveRoleOptions) error
// }

func ExpectToMatchAriaSnapshot(assertions playwrightgo.PlaywrightAssertions, locator playwrightgo.Locator, action *ExecutorAction, page ...*playwrightgo.Page) error {
	assert := assertions.Locator(locator)
	return assert.ToMatchAriaSnapshot(cmp.Or(action.Selector, action.Content)) // TODO: support options - (expected string, options ...LocatorAssertionsToMatchAriaSnapshotOptions) error
}
