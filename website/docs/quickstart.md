# Quick Start

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

