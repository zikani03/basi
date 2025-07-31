package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/zikani03/basi"
	"github.com/zikani03/basi/playwright"
	"gopkg.in/yaml.v2"
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

type RunCmd struct {
	File      string `arg:"" help:"filename for file to run"`
	Directory string `short:"d" help:"directory containing .basi files to be run"`
	URL       string `short:"u" help:"which url to run the test against"`
	Remote    bool   `help:"whether to run remote test"`
	Docker    bool   `help:"whether to run tests inside docker"`
	Local     bool   `help:"whether to install playwright locally and run tests"`
	OutputDir string `short:"o" help:"Where to write test output and screenshots"`
}

type TestCmd struct {
	File string `help:"File to test"`
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func (r *RunCmd) Run(globals *Globals) error {
	fileData, err := os.ReadFile(r.File)
	if err != nil {
		return err
	}

	executor := &playwright.Executor{}
	actions := make([]playwright.ExecutorAction, 0)

	if strings.HasSuffix(r.File, ".basi") {
		parsedActions, err := basi.Parse(r.File, bytes.NewBuffer(fileData))
		if err != nil {
			return err
		}

		for _, p := range parsedActions.Actions {
			actions = append(actions, *playwright.NewExecutorAction(p))
		}

		executor = &playwright.Executor{
			URL:      r.URL,
			Browser:  globals.Browser,
			Actions:  actions,
			Headless: globals.Headless,
		}

	} else if strings.HasSuffix(r.File, ".yaml") || strings.HasSuffix(r.File, ".yml") {
		if err := yaml.Unmarshal(fileData, executor); err != nil {
			return fmt.Errorf("unable to parse step got: %v", err)
		}

	} else {
		return fmt.Errorf("failed to run, invalid file specified: %s", r.File)
	}

	slog.Debug("running the executor", "url", executor.URL)
	_, err = executor.Run(context.Background())
	if err != nil {
		return err
	}
	slog.Debug("executed sucessfully")
	return nil
}

type CLI struct {
	Globals
	Run RunCmd `cmd:"" help:"Run tests using basi"`
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
