package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

var version string = "0.0.0"

type Globals struct {
	Headless bool        `help:"Run in headless mode or not"`
	Browser  string      `help:"Browser to use for the tests, defaults to chromuium"`
	Debug    bool        `help:"Enable debug mode"`
	Version  VersionFlag `name:"version" help:"Show version and quit"`
}

type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

type CLI struct {
	Globals
	Run           RunCmd                `cmd:"" help:"Run tests using basi"`
	RunDoc        RunDocCmd             `cmd:"" help:"Experimental feature to generate docs from steps"`
	Test          TestCmd               `cmd:"" help:"Test a .basi file for syntax"`
	GenPlaywright GeneratePlaywrightCmd `cmd:"" help:"Generate playwright equivalent JavaScript file"`
}

func main() {
	cli := CLI{
		Globals: Globals{
			Version: VersionFlag(version),
		},
	}

	ctx := kong.Parse(&cli,
		kong.Name("basi"),
		kong.Description("Run tests via playwright"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": version,
		})

	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
