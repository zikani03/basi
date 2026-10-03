# Basi GitHub Issues — Evaluation & Implementation Spec

> Updated: 2026-10-03  
> Repo: github.com/zikani03/basi  
> Issues pulled from: `gh issue list`

---

## Summary

| # | Title | Effort | Priority | Verdict |
|---|-------|--------|----------|---------|
| [#16](#16) | Robust CI pipeline | M | High | Implement |
| [#15](#15) | Inherit config from another file | S | Medium | Implement |
| [#9](#9)  | Tab support (NewTab/OpenTab) | M | Medium | Implement |
| [#29](#29) | Lock browser/playwright version | M | Medium | Implement |
| [#26](#26) | Faker support for inputs | M | Low | Implement |
| [#36](#36) | Support LightPanda browser | M | Medium | Implement |
| [#18](#18) | Check source files (dev subcommand) | L | Low | Defer |

Effort: S = small (< 1 day), M = medium (1–3 days), L = large (> 3 days)

### Implemented

| # | Title | Commit |
|---|-------|--------|
| #34 | Emit progress as JSON Lines | 2ff565a |
| #17 | Repeated runs (N times) | 2ff565a |
| #8  | Output modes (JSON/HTML/Text) | 2ff565a |

---

## Issue Details

---

### #16 — Add robust CI pipeline {#16}

**State:** Open  
**Effort:** Medium  
**Priority:** High

#### What it asks for

A GitHub Actions CI pipeline that runs basi's own tests and exercises real actions/assertions against a test web app. The suggested target is `cypress-realworld-app`.

#### Evaluation

Basi currently has no CI. Given that it's a browser automation tool, tests must run in a real browser, which requires Playwright browsers to be installed in the CI environment. GitHub Actions with Ubuntu runners is the natural fit.

#### Implementation Plan

1. **`.github/workflows/ci.yml`**:
   ```yaml
   name: CI
   on: [push, pull_request]
   jobs:
     test:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with: { go-version: '1.22' }
         - run: go build ./...
         - run: go test ./...
         - name: Install Playwright browsers
           run: go run playwright/install/main.go  # or equivalent
         - name: Start cypress-realworld-app
           run: |
             npx --yes create-react-app@latest . --template ... # or use the Docker image
             # Start in background, wait for port
         - name: Run basi e2e specs
           run: basi run ./testdata/*.basi
   ```
2. **Testdata `.basi` files** covering the cypress-realworld-app login, transaction creation, and notification flows.
3. **Matrix build** — test on `ubuntu-latest` and `macos-latest`; `windows-latest` optional.
4. **Caching** — cache `~/.cache/ms-playwright` across runs to avoid re-downloading browsers.

#### Files to touch

- `.github/workflows/ci.yml` (new)
- `testdata/` (new directory with `.basi` specs targeting cypress-realworld-app)

---

### #15 — Inherit configuration from another `.basi` file {#15}

**State:** Open  
**Effort:** Small  
**Priority:** Medium

#### What it asks for

Allow a `.basi` file's frontmatter to extend another file, importing selected configuration sections:
```
Extends : "./base.basi;Browsers,ScreenSizes,Devices,NetworkSpeed"
```

#### Evaluation

Similar to `tsconfig extends` or YAML anchors. Useful for sharing browser matrices, viewport configs, and network profiles across many spec files. The frontmatter parser already exists; this is an extension of the config merge step.

#### Implementation Plan

1. **`Extends` frontmatter key** — add to the frontmatter struct in `lib.go`.
2. **Config merge** — after parsing the current file's frontmatter, if `Extends` is set:
   - Parse the referenced file's frontmatter only (stop at `---`).
   - Extract the named sections (`Browsers`, `ScreenSizes`, etc.).
   - Merge into the current config, with local values taking precedence (local wins).
3. **Circular reference detection** — maintain a `seen` set of resolved paths; error if a cycle is detected.
4. **Relative path resolution** — resolve the `Extends` path relative to the directory of the current `.basi` file.

#### Files to touch

- `lib.go` — `Extends` field in frontmatter struct, merge logic
- `core/types.go` — extend the config struct if needed

---

### #9 — Tab support (NewTab / OpenTab) {#9}

**State:** Open  
**Effort:** Medium  
**Priority:** Medium

#### What it asks for

New actions to open and switch between browser tabs:
- `NewTab "<optional url>"` — open a new tab, optionally navigate to a URL.
- `OpenTab "<index>|<name>|<regex>"` — switch to a tab by index, title, or title regex.

#### Evaluation

Playwright (and `playwright-go`) supports multiple pages via `browser.NewPage()` and `context.Pages()`. The primary challenge is tracking the "active page" through the action runner, which currently operates on a single `playwright.Page`.

#### Implementation Plan

1. **`RunContext` struct** (or extend `core/types.go`) — add a `Pages []playwright.Page` slice and `ActivePage int` index.
2. **`NewTab` action** in `lib.go` grammar:
   ```
   NewTab [ <url> ]
   ```
   Implementation: call `context.NewPage()`, append to `Pages`, set `ActivePage` to the new page, optionally navigate.
3. **`OpenTab` action**:
   ```
   OpenTab <index|name|regex>
   ```
   Implementation: match against `page.Title()` or use integer index; update `ActivePage`.
4. **Update all existing actions** in `playwright/actions.go` to dispatch on `ctx.Pages[ctx.ActivePage]` instead of a bare `page` variable.
5. **`CloseTab`** — close the current tab, switch back to the previous one.

#### Files to touch

- `lib.go` — `NewTab`, `OpenTab`, `CloseTab` grammar nodes
- `playwright/actions.go` — multi-page dispatch
- `core/types.go` — `RunContext` with page list

---

### #29 — Lock browser/playwright version {#29}

**State:** Open  
**Effort:** Medium  
**Priority:** Medium

#### What it asks for

Allow a project-level lock (e.g., `basi.lock` or frontmatter keys) to pin the playwright-go version and specific browser binary version used. This ensures reproducible runs and avoids unnecessary downloads.

#### Evaluation

`playwright-go` downloads browser binaries via `playwright.Install()`. Version pinning requires either:
- Embedding the playwright-go version in a lock file and checking it matches the running binary.
- Storing expected browser channel / revision in the lock, and skipping download if the cached binary matches.

The simplest effective approach: a `basi.lock` JSON file that records playwright driver version + browser channel + revision hash, committed to VCS.

#### Implementation Plan

1. **`basi.lock` file format** (JSON):
   ```json
   {
     "playwright": "0.6100.0",
     "browsers": {
       "chromium": { "revision": "1234", "channel": "stable" },
       "firefox":  { "revision": "5678" }
     }
   }
   ```
2. **`basi lock` subcommand** — writes current playwright version and installed browser revisions to `basi.lock`.
3. **Lock enforcement on `basi run`** — if `basi.lock` exists in the working directory, validate it against the running playwright driver version and cached browser revisions. Fail fast with a clear message if there is a mismatch.
4. **Frontmatter alternative** — allow `BrowserVersion : "stable"` in `.basi` file frontmatter for single-file pinning.
5. **`--ignore-lock` flag** — bypass lock check for CI environments that manage browsers separately.

#### Files to touch

- `cmd/main.go` — `lock` subcommand, `--ignore-lock` flag
- `playwright/playwright.go` — revision detection, lock validation
- `core/lock.go` (new) — lock file read/write

---

### #26 — Faker support for generating input values {#26}

**State:** Open  
**Effort:** Medium  
**Priority:** Low

#### What it asks for

Allow `.basi` scripts to use faker-generated values (random names, emails, phone numbers, dates, etc.) as arguments to `Fill` and similar actions, rather than hardcoded strings.

#### Evaluation

This removes the need for external data generation scripts and enables property-style tests. A template expression approach (e.g., `Fill "input[name=email]" "{{faker.email}}"`) integrates cleanly with the existing DSL parser without grammar changes.

A suitable Go faker library: `github.com/brianvoe/gofakeit/v6` (zero external deps, rich API).

#### Implementation Plan

1. **Add `gofakeit` dependency.**
2. **Argument interpolation** — in `core/vars.go` (which already handles variable substitution), add a `interpolateFaker(s string) string` function that replaces `{{faker.<type>}}` tokens:
   - `{{faker.email}}`, `{{faker.name}}`, `{{faker.phone}}`, `{{faker.uuid}}`, `{{faker.date}}`, `{{faker.int}}`, `{{faker.word}}`, etc.
3. **Seed support** — add optional `FakerSeed : "12345"` frontmatter key so runs can be made deterministic for debugging.
4. **Run interpolation** before actions execute — call `interpolateFaker` on all string arguments at action dispatch time.

#### Files to touch

- `core/vars.go` — `interpolateFaker` function
- `lib.go` — pass interpolated values through to actions
- `go.mod` / `go.sum` — add `gofakeit`

---

### #36 — Support LightPanda out of the box {#36}

**State:** Open  
**Effort:** Medium  
**Priority:** Medium

#### What it asks for

Allow `--browser=lightpanda` to work the same way `--browser=firefox` or `--browser=chromium` works. Basi should auto-download the correct OS build of LightPanda if it is not present, run the script using LightPanda, and emit warnings for unsupported features.

#### Evaluation

LightPanda is a Zig-built lightweight browser with Playwright-compatible CDP support, aimed at fast, low-overhead headless testing. The main challenge is that `playwright-go` does not include a LightPanda browser type natively — the integration must go through the CDP (Chrome DevTools Protocol) remote connection path (`playwright.ConnectOverCDP`).

Key constraints noted in the linked blog post:
- LightPanda exposes a CDP endpoint but does not support the full Playwright feature set (e.g., screenshots may be limited).
- The binary must be downloaded separately per OS/arch (no official Playwright install).

#### Implementation Plan

1. **Add `lightpanda` as a recognised browser name** in `cmd/main.go` alongside the existing `--browser` flag handling.
2. **Download helper** — in `playwright/playwright.go` or a new `playwright/lightpanda.go`:
   - Resolve the correct GitHub release asset URL for the host OS/arch from LightPanda's release feed.
   - Download the binary to a cache directory (e.g., `~/.basi/browsers/lightpanda`).
   - Make it executable.
3. **Launch via CDP** — use `playwright.ConnectOverCDP("http://localhost:<port>")` after starting the LightPanda process as a subprocess. Wire the resulting `Browser` into the existing action runner.
4. **Feature warnings** — maintain a static list of actions/assertions known to be unsupported (e.g., `Screenshot`) and emit `WARN` log lines before executing them.
5. **`--browser=lightpanda` flag docs** — update help text and README.

#### Files to touch

- `cmd/main.go` — flag handling
- `playwright/playwright.go` — browser selection switch
- `playwright/lightpanda.go` (new) — download + CDP launch logic

---

### #18 — Subcommand to check source files in dev environments {#18}

**State:** Open  
**Effort:** Large  
**Priority:** Low  
**Verdict: Defer**

#### What it asks for

A dev-mode subcommand that parses source files (HTML, JSX, Vue, etc.) and cross-checks that selectors used in `.basi` scripts actually exist in the source, catching selector drift before running the browser.

#### Evaluation

This is a genuinely useful DX feature but is significantly more complex than other issues:
- Requires parsing multiple source formats (HTML, React JSX, Vue SFCs).
- Selectors can be dynamic (generated at runtime) — false positives are likely.
- Tight coupling to frontend framework conventions.

The value is real but the implementation surface is large, and basi's core value is in runtime browser automation. This is better served by a separate lint plugin or integration with existing tools (ESLint, Biome) once the core feature set is stable.

If implemented in future:
- `basi check --src=./src <file.basi>` subcommand.
- Parse HTML with `golang.org/x/net/html`; parse JSX/Vue as text with regex heuristics for `id=`, `className=`, `data-testid=`.
- Report selectors in `.basi` that have no matching element in source.

---

## Implementation Order

1. **#16 CI pipeline** — establishes a safety net for all subsequent changes.
2. **#15 Config inheritance** — small, no dependencies.
3. **#9 Tab support** — medium, requires RunContext refactor.
4. **#29 Version locking** — medium, operational value.
5. **#26 Faker support** — medium, low priority.
6. **#36 LightPanda** — medium, experimental browser.
7. **#18 Source file checker** — large, defer to later milestone.
