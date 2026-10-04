---
name: basi
homepage: "https://github.com/zikani03/basi"
description: "**basi** is a browser automation tool that runs end-to-end tests via a plain-text `.basi` DSL. It wraps Playwright so testers can write automation scripts without JavaScript. Tests are
declarative, read like English, and are stored as simple text files."
---

## What basi Is

**basi** is a browser automation tool that runs end-to-end tests via a plain-text `.basi` DSL.
It wraps Playwright so testers can write automation scripts without JavaScript. Tests are
declarative, read like English, and are stored as simple text files.

Install: `go install github.com/zikani03/basi@latest`

Binary releases: <https://github.com/zikani03/basi/releases>

---

## Your Job as an Agent

When a developer asks you to write, fix, or run basi e2e tests:

1. Understand what UI flows need coverage (ask if unclear).
2. Write `.basi` file(s) that test those flows.
3. Validate syntax: `basi test <file.basi>`.
4. Run the test: `basi run <file.basi>`.
5. Interpret results and fix failures.

---

## .basi File Format

### Minimal file

```
Goto "https://example.com"
Find "Sign in"
Click
```

### With frontmatter

Frontmatter is optional YAML-like config at the top, terminated by `---`.

```
ID          : "checkout-flow"
Title       : "Full checkout smoke test"
URL         : "https://staging.example.com"
Headless    : "yes"
Browsers    : "chromium"
ScreenSizes : "1280x800"
---
Goto "/shop"
Click "#add-to-cart"
Goto "/checkout"
Fill "#email" "buyer@example.com"
Click "#place-order"
WaitFor ".confirmation-banner"
ExpectText "Order confirmed"
```

**All frontmatter fields are optional.** If you skip frontmatter, start the file directly with actions.

### Comments

Lines starting with `#` or `;` are ignored.

```
# Navigate to login page
Goto "/login"
```

No inline comments — the comment must be on its own line.

---

## Action Reference

### Navigation

```
Goto "https://example.com"          # absolute URL
Goto "^/dashboard"                  # regex — matches any URL matching pattern
WaitForURL "^/dashboard"            # wait until URL matches regex
GoBack
GoForward
Refresh
```

### Interacting with elements

```
Click "#submit-btn"
DoubleClick ".item-card"
Tap "#mobile-button"                # mobile tap
Fill "#email" "user@example.com"
Clear "#search-input"
Type "#search" "query text"         # simulates keystrokes
Press "#input" "Enter"
PressSequentially "#pin" "1234"
Focus "#username"
Blur "#username"
```

### Checkboxes & selects

```
Check "#agree"
Uncheck "#newsletter"
Select "#country" "Malawi"
SelectMultipleOptions "#tags" "Go,Python,Rust"
```

### Finding elements

`Find` selects an element by **visible text content**. Subsequent actions / assertions operate on
the found element without needing to repeat the selector.

```
Find "Add to cart"
Click                               # clicks the found element

FindNth "button" "2"                # 2nd matching button
FindFirst ".product-card"
FindLast ".product-card"
FindMatching "h2" "^Latest"         # regex match on text
```

### Waiting

```
WaitFor "#modal"                    # wait for selector to appear
WaitForSelector ".loading-done"
```

### File uploads

```
Upload "#avatar-input" "./avatar.png"
UploadFiles "#attachments" "./file1.pdf,./file2.pdf"
```

### Screenshots

```
Screenshot "body" "./screenshots/checkout.png"
Screenshot "#payment-form" "./screenshots/form.png"
```

### Reusing test files

```
Use "./shared/login.basi"           # inline another .basi file
```

---

## Assertion Reference

Assertions follow a `Find` (or stand alone with an explicit selector) to verify page state.

### Text & value

```
Find "Username"
ExpectText "johndoe"                # element text equals
ExpectValue "johndoe"               # input value equals
```

### Attributes & CSS

```
Find "Status badge"
ExpectAttr "data-status" "active"
ExpectClass "is-active"
ExpectToContainClass "badge"
ExpectCSS "color" "rgb(0, 128, 0)"
```

### Visibility & state

```
Find "Error message"
ExpectVisible
ExpectHidden
ExpectEnabled
ExpectDisabled
ExpectEditable
ExpectChecked
ExpectEmpty
ExpectAttached                      # element exists in DOM
ExpectInViewport
ExpectFocused
```

### Accessibility

```
Find "Main nav"
ExpectAccessibleName "primary navigation"
ExpectAccessibleDescription "Site navigation links"
```

---

## Property-Based Testing

For robustness testing beyond simple scripts:

```
Always "ExpectVisible #main-nav"    # global invariant — checked after every mutating action

Extract "$title" "h1"               # capture text into a variable

Click "#edit-toggle"
Next "ExpectVisible .edit-form"     # assert immediate next state

Fuzz "5 .form-field"                # perform 5 random interactions on matching elements

Eventually "ExpectText $title h1"  # assert eventual state (with variable interpolation)
```

**Fuzz** has a hard 30-second timeout. Only interacts with visible, enabled elements.

---

## Running Tests

```sh
# Basic run
basi run test.basi

# Override URL (useful for staging/prod targeting)
basi run test.basi -u "https://staging.example.com"

# Headless (no browser window)
basi run test.basi --headless

# Custom timeout per action
basi run test.basi -t "30s"

# Run all .basi files in a directory
basi run -d ./e2e/

# Repeat N times (reliability testing)
basi run test.basi --repeat 5 --fail-fast

# Generate HTML report
basi run test.basi --output html --output-file report.html

# JSON output (for CI parsing)
basi run test.basi --output json

# Validate syntax only (no browser launched)
basi test test.basi

# Remote browser via CDP (e.g. Cloudflare Browser Rendering)
basi run test.basi --cdp-endpoint "wss://api.cloudflare.com/..."
```

Exit code `0` = success, `1` = failure or error.

---

## Best Practices

### 1. Prefer semantic selectors

```
# Good — stable, intention-revealing
Click "[data-testid='submit-order']"
Fill "#email-input" "user@example.com"

# Avoid — brittle against DOM restructure
Click "div.container > div:nth-child(2) > button"
```

### 2. Wait before asserting

```
# Good — explicit wait before assertion
Click "#load-data"
WaitFor ".data-table"
Find "Total"
ExpectText "42"

# Risky — may assert before content renders
Click "#load-data"
ExpectText "42"
```

### 3. Modularise with Use

```
# shared/login.basi
Fill "input[name=email]" "user@example.com"
Fill "input[name=password]" "secret"
Click "button[type=submit]"
WaitFor "#dashboard"
```

```
# checkout.basi
Use "./shared/login.basi"
Goto "/shop"
Click "#add-to-cart"
# ...
```

### 4. Screenshot on key state transitions

```
Click "#place-order"
WaitFor ".confirmation-banner"
Screenshot "body" "./screenshots/confirmation.png"
```

### 5. Use frontmatter for environment separation

Set `URL` in frontmatter and override at runtime with `-u` to target different environments
without changing test files:

```sh
basi run checkout.basi -u "https://staging.example.com"
basi run checkout.basi -u "https://prod.example.com"
```

### 6. Register global invariants early

```
Always "ExpectVisible #main-nav"
Always "ExpectVisible footer"

# ... rest of test actions
```

### 7. Name screenshots with context

```
Screenshot "body" "./screenshots/01-login-page.png"
Screenshot "body" "./screenshots/02-after-submit.png"
```

---

## Limitations to Know Before Writing Tests

| Limitation | Detail |
|---|---|
| **No multi-tab support** | All actions run on one page. Tests that open new tabs will lose context. |
| **No conditions / loops** | The DSL has no `if`, `for`, or branching logic. |
| **No test data generation** | No Faker-like data generation built in; hardcode or parameterise via URL override. |
| **Comments must be full-line** | `Click "#btn" # comment` is not valid; the comment breaks parsing. |
| **Variable interpolation is limited** | `$var` works only in `Always`, `Eventually`, and `Extract`. Not in `Fill` or `Click`. |
| **Circular `Use` imports not detected** | Avoid `A.basi` using `B.basi` which uses `A.basi`. |
| **Fuzz has a 30 s hard cap** | Long fuzz counts on complex pages will be cut short. |
| **Default action timeout is 10 s** | For slow pages, override globally with `-t "30s"`. Max is 300 s. |
| **Early-stage software** | No API stability guarantees. Check the changelog when upgrading. |
| **`--output text` needs page body** | Text-mode reports only work when page body was captured during the run. |

---

## CI/CD Integration Examples

### GitHub Actions

```yaml
- name: Install basi
  run: go install github.com/zikani03/basi@latest

- name: Run e2e tests
  run: basi run -d ./e2e/ --headless --output html --output-file e2e-report.html

- uses: actions/upload-artifact@v4
  with:
    name: e2e-report
    path: e2e-report.html
```

### GitLab CI

```yaml
e2e:
  script:
    - go install github.com/zikani03/basi@latest
    - basi run -d ./e2e/ --headless --output json
  artifacts:
    paths:
      - "*.png"
```

---

## Troubleshooting Guide for Agents

| Symptom | Likely cause | Fix |
|---|---|---|
| `parse error` on run | Inline comment or bad frontmatter field | Check for `# comment` after an action; move comment to its own line |
| `timeout waiting for selector` | Element never appears / wrong selector | Add `WaitFor` before assertion; verify selector with `basi test` |
| Test passes locally, fails in CI | Missing `--headless` flag | Add `--headless` or set `Headless: "yes"` in frontmatter |
| `Use` file not found | Relative path resolution | Paths are relative to the **calling** file's directory |
| Screenshot is blank | Page not loaded yet | Add `WaitFor` for a stable element before `Screenshot` |
| `ExpectText` fails on dynamic content | Content changes or loads async | Use `WaitFor` on a stable anchor element first |

---

## Example: Complete Login + Dashboard Test

```
ID       : "login-dashboard-smoke"
Title    : "Login and dashboard smoke test"
URL      : "https://app.example.com"
Headless : "yes"
---

# Register invariants
Always "ExpectVisible #app-shell"

# Navigate to login
Goto "/login"
Screenshot "body" "./screenshots/01-login.png"

# Fill credentials
Fill "#email" "qa@example.com"
Fill "#password" "test-password-123"

# Submit and wait for redirect
Click "button[type=submit]"
WaitForURL "^/dashboard"

Screenshot "body" "./screenshots/02-dashboard.png"

# Verify dashboard elements
Find "Welcome back"
ExpectVisible

Find "Recent activity"
ExpectVisible

# Check nav links are present
Find "Settings"
ExpectAttr "href" "/settings"
```
