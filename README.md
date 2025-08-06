# basi

`basi` allow users and developers to author and run Playwright actions using a simple
configuration file with less code. Browser automation steps are written in `.basi` files. 

The goal is for the .basi file DSL to be simple enough to hand over to non-technical users.

> NOTE: `basi` is still in very early development. There are no guarantees about API or feature stability.

## Installation 

Download a binary from the [GitHub Releases](https://github.com/zikani03/basi/releases) and place it on your $PATH.

> NOTE: `basi` depends on Playwright and needs to download some
> browsers and tools if playwright if it is not already installed.
> You will notice this the first time you run the test/files

If you want to contribute or build from the source code see the [Building](#building) section

Once installed you can then run it :

```sh
$ basi --help
```

## Example usage

Create a file named `example-hn.basi` file with the following content:

```
Goto "https://news.ycombinator.com/login"
Fill "input[name=acct]" "throwaway-username" 
Fill "input[name=pw]" "fakepassword"
Click "input[value=login]"
Screenshot "body" "./test-screenshot.png"
```

You can now run the file using basi, like so:

```sh 
$ basi run example-hn.basi
```

**You can use `Find` to select an element to run assertions on it**

```
Goto "https://github.com/"
Find "Build and ship software on a single, collaborative platform"
ExpectId "hero-section-brand-heading"
Screenshot "body" "./data/test-github.png"
```

**You can setup metadata/configuration for each run in a `frontmatter` section**

```
ID            : "A random ID to identify the run"
URL           : "https://nndi.cloud/"
Title         : "Navigate to home on nndi"
Headless      : "yes"
Description   : "Navigates to the NNDI website and clicks the Home link"
---
Goto "/"
Click "#navbar > ul > li.active > a"
ExpectAttr "data-nav-section" "home"
Screenshot "body" "./test-nndi.png"
```

## Available actions

|Action|Arguments|Example|
|------|---------|-------|
|Click                 |**querySelector**| Click "#element" |
|DoubleClick           |**querySelector**| DoubleClick "#element" |
|Tap                   |**querySelector**| Tap "#element" |
|Focus                 |**querySelector**| Focus "#element" |
|Blur                  |**querySelector**| Blur "#element" |
|Fill                  |**querySelector** TEXT| Fill "#element" "my input text" |
|Find                  |**textContent or querySelector**| Find "My Account" |
|Clear                 |**querySelector**| Clear "#element" |
|Check                 |**querySelector**| Check "#element" |
|Uncheck               |**querySelector**| Uncheck "#element" |
|FillCheckbox          |**querySelector**| FillCheckbox "#element" |
|Press                 |**querySelector** TEXT| Press "#element" "some text"|
|PressSequentially     |**querySelector** TEXT | PressSequentially "#element" "some input"|
|Type                  |**querySelector** TEXT | Type "#element" |
|Screenshot            |**querySelector** TEXT | Screenshot "#selector" "filename.png"|
|Select                |**querySelector** TEXT | Select "#someSelect" "Value or Label"|
|SelectOption          |**querySelector** TEXT | Select "#someSelect" "Value or Label"|
|SelectMultipleOptions |**querySelector** TEXT | SelectMultipleOptions "#someSelect" "Value or Label 1,Value or Label 2,..., Value or Label N"|
|WaitFor               |**querySelector**| WaitFor "#element" |
|WaitForSelector       |**querySelector**| WaitForSelector "#element" |
|Goto                  |**REGEX**| Goto "^some-page" |
|WaitForURL            |**REGEX**| WaitForURL "^some-page" |
|GoBack                |N/A| GoBack |
|GoForward             |N/A| GoForward |
|Refresh               |N/A| Refresh |

## Expects

`basi` implements most of [Playwright's Assertions](https://playwright.dev/docs/test-assertions) via Expect actions. The following Expect actions are currently supported


|Action|Arguments|Example|
|------|---------|-------|
|ExpectText| **argument** | ExpectText "Click Here" |
|ExpectAttr| **attributeName** **argument** | ExpectAttr "name" "some-name" |
|ExpectAttribute| **argument** | ExpectAttribute "name" "some-name" |
|ExpectValue| **argument** | ExpectValue "something" |
|ExpectValues| **argument** | ExpectValues "something,something" |
|ExpectAttached| None | ExpectAttached  |
|ExpectChecked| None | ExpectChecked |
|ExpectDisabled| None ExpectDisabled |
|ExpectEditable| None | ExpectEditable |
|ExpectEmpty| None | ExpectEmpty  |
|ExpectEnabled| None | ExpectEnabled  |
|ExpectFocused| None | ExpectFocused |
|ExpectHidden| None| ExpectHidden |
|ExpectInViewport| None | ExpectInViewport |
|ExpectVisible| None| ExpectVisible |
|ExpectToContainClass| **argument** | ExpectToContainClass "class-name" |
|ExpectToContainText| **argument** | ExpectToContainText "something" |
|ExpectAccessibleDescription| **argument** | ExpectAccessibleDescription "description"  |
|ExpectAccessibleErrorMessage| **argument** | ExpectAccessibleErrorMessage "An error message" |
|ExpectAccessibleName| **argument** | ExpectAccessibleName "An accessible name" |
|ExpectClass| **className** | ExpectClass "a-class-name" |
|ExpectCSS| **css-property** **argument** | ExpectCSS "display" "flex" |
|ExpectId| **argument** | ExpectId "an-id" |

## Building 

```sh
$ git clone https://github.com/zikani03/basi

$ cd basi

$ go build -o basi ./cmd/main.go

$ ./basi --help

# Test with the example file in the repo

$ ./basi run example-hn.basi --url "https://news.ycombinator.com"
```

## LICENSE

Apache 2.0 LICENSE