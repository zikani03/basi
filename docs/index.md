# pact

Some notes about pact

- It implements a DSL 
- It uses [playwright-go](https://github.com/playwright-community/playwright-go) under the hood.
- Inspired/triggered by a PR for [ovh/venom](https://github.com/ovh/venom/pull/843)

## Objectives

- Support YAML and .pact files
- Support Assertions for full end-to-end testing flow
- Generate structured output files for integration into (CLIs, AI Agents, general system integrations/connects)
- Support 95% of Playwright actions that can be done on programmatic SDKs
- Support running playwright `--local`, `--docker` and `--remote`
- Support generating documents from/in-between steps
- Figure out how to incorporate the playwright test generator
- Design import capability for code reause
- Suppor faker and fuzztesting
- Support variables across the session/tets

## playwright Actions DSL

Playwright Actions (PACT) is a small DSL (domain specific language) for interacting with [playwright](https://playwright.dev).

It allows users to perform actions via Playwright without having to use 
actual programmatic SDKs or syntax - opening Playwright up to less technical
users and faster authoring of end to end UI tests.

This repository implements a basic Lexer and Parser for PACT using [participle](https://github.com/alecthomas/participle). Each action is specified on its own line.


## IDEAS

```
[Config]
  name: 
  url: "https://localhost:5173"
  headless: true 
  device: "Samsung Galaxy"

Use "./login.pact"

Fill       "#email" "change@example.com" 
Fill       "#email" "zikani@example.com" 
Fill       "#password"  "zikani123" 
Screenshot "body" "./test-screenshot.png" 
Click      "#loginButton" 
WaitFor    ".second-dashboard-user-name"

Use "./generate-document.pact"

[Asserts]
page body is "something"
```
