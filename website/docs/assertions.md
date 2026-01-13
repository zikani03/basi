# Assertions

Assertions are tests you can perform to verify some state in the automation process. Assertions, named Expects in `basi` help you to check expected conditions which is useful for verifying different things for example that an element or component appears on the page or that a form field contains a certain value.

`basi` implements most of [Playwright's Assertions](https://playwright.dev/docs/test-assertions) via
Expect actions. 

The following Asserions / Expect actions are currently supported:

| Action                       | Arguments                      | Example                                         |
| ---------------------------- | ------------------------------ | ----------------------------------------------- |
| ExpectText                   | **argument**                   | ExpectText "Click Here"                         |
| ExpectAttr                   | **attributeName** **argument** | ExpectAttr "name" "some-name"                   |
| ExpectAttribute              | **argument**                   | ExpectAttribute "name" "some-name"              |
| ExpectValue                  | **argument**                   | ExpectValue "something"                         |
| ExpectValues                 | **argument**                   | ExpectValues "something,something"              |
| ExpectAttached               | None                           | ExpectAttached                                  |
| ExpectChecked                | None                           | ExpectChecked                                   |
| ExpectDisabled               | None ExpectDisabled            |                                                 |
| ExpectEditable               | None                           | ExpectEditable                                  |
| ExpectEmpty                  | None                           | ExpectEmpty                                     |
| ExpectEnabled                | None                           | ExpectEnabled                                   |
| ExpectFocused                | None                           | ExpectFocused                                   |
| ExpectHidden                 | None                           | ExpectHidden                                    |
| ExpectInViewport             | None                           | ExpectInViewport                                |
| ExpectVisible                | None                           | ExpectVisible                                   |
| ExpectToContainClass         | **argument**                   | ExpectToContainClass "class-name"               |
| ExpectToContainText          | **argument**                   | ExpectToContainText "something"                 |
| ExpectAccessibleDescription  | **argument**                   | ExpectAccessibleDescription "description"       |
| ExpectAccessibleErrorMessage | **argument**                   | ExpectAccessibleErrorMessage "An error message" |
| ExpectAccessibleName         | **argument**                   | ExpectAccessibleName "An accessible name"       |
| ExpectClass                  | **className**                  | ExpectClass "a-class-name"                      |
| ExpectCSS                    | **css-property** **argument**  | ExpectCSS "display" "flex"                      |
| ExpectId                     | **argument**                   | ExpectId "an-id"                                |

