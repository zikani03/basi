# Skill: basi E2E Testing

Use this skill when the user asks you to write, run, debug, or generate basi end-to-end tests,
or when they mention a `.basi` file.

## What to do

1. **Understand the flow** — Ask which user flows or pages need coverage if not specified.
2. **Inspect the app** — If a local dev server is running, use the browser or read source to learn selectors.
3. **Write the `.basi` file** using the syntax below.
4. **Validate**: `basi test <file.basi>` (no browser launched — fast syntax check).
5. **Run**: `basi run <file.basi> --headless`.
6. **Fix failures** by reading error output and adjusting selectors, waits, or assertions.

---

## .basi File Syntax

### Frontmatter (optional)

```
ID          : "unique-id"
Title       : "Human-readable name"
URL         : "https://app.example.com"
Headless    : "yes"
Browsers    : "chromium"
ScreenSizes : "1280x800"
---
```

Frontmatter ends at `---`. Skip it entirely if not needed.

### Comments

`#` or `;` at the start of a line. **No inline comments** — a `#` after an action breaks parsing.

---

## Actions Quick Reference

### Navigation
```
Goto "https://example.com"
Goto "^/dashboard"              # regex — match URL pattern
WaitForURL "^/dashboard"
GoBack
GoForward
Refresh
```

### Interaction
```
Click "#selector"
Fill "#input" "text value"
Clear "#input"
Type "#input" "types like a human"
Press "#input" "Enter"
Check "#checkbox"
Uncheck "#checkbox"
Select "#dropdown" "Option Label"
Tap "#mobile-btn"
Focus "#field"
Blur "#field"
```

### Finding elements
```
Find "Visible text on page"      # select by text content; subsequent actions use this element
FindNth "button" "3"
FindFirst ".card"
FindLast ".card"
FindMatching "h2" "^Hello"       # regex on text
```

### Waiting
```
WaitFor "#element"
WaitForSelector ".spinner-gone"
```

### Assertions (use after Find or with explicit selector)
```
ExpectText "exact text"
ExpectValue "input value"
ExpectAttr "data-status" "active"
ExpectClass "is-active"
ExpectToContainClass "badge"
ExpectCSS "display" "flex"
ExpectVisible
ExpectHidden
ExpectEnabled
ExpectDisabled
ExpectChecked
ExpectEditable
ExpectEmpty
ExpectAttached
ExpectInViewport
ExpectFocused
ExpectAccessibleName "nav label"
```

### Screenshots
```
Screenshot "body" "./screenshots/step-01.png"
Screenshot "#form" "./screenshots/form.png"
```

### Reuse
```
Use "./shared/login.basi"
```

### Property-based testing
```
Always "ExpectVisible #header"       # invariant checked after every mutating action
Extract "$varName" "h1"              # capture element text into variable
Click "#button"
Next "ExpectVisible .toast"          # immediate next-state assertion
Fuzz "5 .form-field"                 # random interactions (30 s hard cap)
Eventually "ExpectText $varName h1"  # eventual-state assertion with variable
```

---

## CLI Commands

```sh
basi test file.basi                            # validate syntax only
basi run file.basi                             # run test
basi run file.basi --headless                  # no browser window
basi run file.basi -u "https://staging.app"   # override URL
basi run file.basi -t "30s"                   # per-action timeout
basi run file.basi --repeat 5 --fail-fast     # reliability run
basi run file.basi --output html              # generate HTML report
basi run file.basi --output json              # JSON output for parsing
basi run -d ./e2e/                            # run entire directory
basi run file.basi --cdp-endpoint "wss://…"  # remote browser via CDP
```

---

## Patterns to Apply

### Login helper (shared)

```
# shared/login.basi
Fill "input[name=email]" "user@example.com"
Fill "input[name=password]" "password"
Click "button[type=submit]"
WaitFor "#dashboard"
```

### Page that requires auth

```
Title : "Dashboard smoke"
URL   : "https://app.example.com"
Headless : "yes"
---
Use "./shared/login.basi"

Always "ExpectVisible #app-shell"

Goto "/dashboard"
WaitFor ".content-loaded"
Screenshot "body" "./screenshots/dashboard.png"

Find "Recent activity"
ExpectVisible
```

### Form submission + success check

```
Goto "/contact"
Fill "#name" "Test User"
Fill "#email" "test@example.com"
Fill "#message" "Hello from basi"
Screenshot "body" "./screenshots/before-submit.png"
Click "button[type=submit]"
WaitFor ".success-message"
ExpectText "Message sent"
Screenshot "body" "./screenshots/after-submit.png"
```

---

## Limitations — Communicate These to the User

- **No multi-tab support** — actions only target the active page; new tabs lose context.
- **No if/else or loops** — DSL is purely sequential.
- **No dynamic test data** — hardcode values or parameterise via `-u` URL override.
- **`$variable` interpolation** only works in `Always`, `Eventually`, `Extract` — not in `Fill` or `Click`.
- **Full-line comments only** — `Click "#btn" # comment` breaks parsing.
- **Circular `Use` imports** are not detected; avoid them.
- **Default action timeout is 10 s** — use `-t "30s"` for slow pages (max 300 s).
- **`Fuzz` hard cap is 30 s** regardless of count.
- **Early-stage tool** — no API stability guarantees between versions.

---

## Troubleshooting

| Error | Cause | Fix |
|---|---|---|
| `parse error` | Inline comment after action | Move `#` comment to its own line |
| `timeout waiting for selector` | Wrong selector or element not rendered | Fix selector; add `WaitFor` |
| Test fails in CI, passes locally | Missing `--headless` | Add `--headless` flag or `Headless: "yes"` |
| `Use` file not found | Relative path wrong | Paths are relative to the **calling file's directory** |
| Screenshot blank | Page not loaded | Add `WaitFor` before `Screenshot` |
| `ExpectText` flaky | Content loads asynchronously | `WaitFor` a stable anchor element first |

---

## When you're done

- Report which `.basi` file(s) were created or modified.
- Show the run command the user should execute.
- If tests passed, show the relevant screenshot paths.
- If tests failed, show the failing action and your diagnosis.
