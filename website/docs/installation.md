# Installation

Download a binary from the
[GitHub Releases](https://github.com/zikani03/basi/releases) and place it on
your $PATH.

> NOTE: `basi` depends on Playwright and needs to download some browsers and
> tools if playwright if it is not already installed. You will notice this the
> first time you run the test/files

If you want to contribute or build from the source code see the
[Building](#building) section

Once installed you can then run it :

```sh
$ basi --help
```


## Building from Source

```sh
$ git clone https://github.com/zikani03/basi

$ cd basi

$ go build -o basi ./cmd/main.go

$ ./basi --help

# Test with the example file in the repo

$ ./basi run example-hn.basi --url "https://news.ycombinator.com"
```