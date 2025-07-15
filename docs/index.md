# pact

Some notes about pact

- It implements a DSL 
- It uses [playwright-go](https://github.com/playwright-community/playwright-go) under the hood.

## Objectives

- Support YAML and .pact files
- Generate structured output files for integration into (CLIs, AI Agents, general system integrations/connects)
- Support 95% of Playwright actions that can be done on programmatic SDKs
- Support running playwright `--local`, `--docker` and `--remote`
- Support generating documents from/in-between steps
- Figure out how to incorporate the playwright test generator
- Design import capability for code reause
- Suppor faker and fuzztesting
- Support variables across the session/tets

## playwright Actions DSL

> NOTE: This is just an experiment for an idea I got while working on a PR for [ovh/venom](https://github.com/ovh/venom/pull/843), YMMV

Playwright Actions (PACT) is a small DSL (domain specific language) for interacting with [playwright](https://playwright.dev).

It allows users to perform actions via Playwright without having to use 
actual programmatic SDKs or syntax - opening Playwright up to less technical
users and faster authoring of end to end UI tests.

This repository implements a basic Lexer and Parser for PACT using [participle](https://github.com/alecthomas/participle)

It looks like this, each action is specified on its own line:

```
Fill "x" "something something here" 
Fill "#x" "something something here"
Fill ".x" "something something here"
Fill "[name=email]" "something something here"
Fill "some element with whitespace" "something something here"
Fill ".input[type=\"text\"]" "something something here"
Click "existent" "something"
Click "GetByRole( \"existent\" )"
```

To play with this repo you can:

```sh 
$ git clone https://github.com/zikani03/pact

$ cd pact

$ cat actions.pact | go run ./cmd/main.go
```

## IDEAS

```
[Config]
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
