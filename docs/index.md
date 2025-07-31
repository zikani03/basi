# basi

Some notes about basi

- It implements a DSL 
- It uses [playwright-go](https://github.com/playwright-community/playwright-go) under the hood.
- Inspired/triggered by a PR for [ovh/venom](https://github.com/ovh/venom/pull/843)

## Objectives

- Support YAML and .basi files
- Support Assertions for full end-to-end testing flow
- Generate structured output files for integration into (CLIs, AI Agents, general system integrations/connects)
- Support 95% of Playwright actions that can be done on programmatic SDKs
- Support running playwright `--local`, `--docker` and `--remote`
- Support generating documents from/in-between steps
- Figure out how to incorporate the playwright test generator
- Design import capability for code reuse
- Suppor faker and fuzztesting
- Support variables across the session/tests

## Playwright Actions DSL (`.basi` file)

The project implements a DSL (domain specific language) for interacting with [playwright](https://playwright.dev). The language is called pact, for [historic reasons](#2)

It allows users to perform actions via Playwright without having to use  actual programmatic SDKs or syntax - opening Playwright up to less technical users and faster authoring of end to end UI tests.
Each action is specified on its own line.

The Lexer and Parser for the DSL is implemented using [participle](https://github.com/alecthomas/participle). 

## IDEAS

```
[Config]
  name: 
  url: "https://localhost:5173"
  headless: true 
  device: "Samsung Galaxy"

Use "./login.basi"

Fill       "#email" "change@example.com" 
Fill       "#email" "zikani@example.com" 
Fill       "#password"  "zikani123" 
Screenshot "body" "./test-screenshot.png" 
Click      "#loginButton" 
WaitFor    ".second-dashboard-user-name"

Use "./generate-document.basi"

[Asserts]
page body is "something"
```
