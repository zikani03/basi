package main

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/zikani03/basi"
	"github.com/zikani03/basi/core"
	"github.com/zikani03/basi/playwright"
	"gopkg.in/yaml.v2"
)

type RunCmd struct {
	File        string `arg:"" help:"filename for file to run"`
	Directory   string `short:"d" help:"directory containing .basi files to be run"`
	URL         string `short:"u" help:"which url to run the test against"`
	CDPEndpoint string `help:"URL to Chrome Developer Protocol compliant browser/server to run tests on"`
	Remote      bool   `help:"whether to run remote test"`
	Docker      bool   `help:"whether to run tests inside docker"`
	Local       bool   `help:"whether to install playwright locally and run tests"`
	OutputDir   string `short:"o" help:"Where to write test output and screenshots"`
	Timeout     string `short:"t" help:"Timeout e.g. 30s"`
	// Progress / noise control
	Quiet  bool `help:"Suppress JSON Lines progress output"`
	Silent bool `help:"Alias for --quiet"`
	// Repeated runs
	Repeat   int  `help:"Run the spec N times sequentially" default:"1"`
	FailFast bool `help:"Stop after the first failing run (use with --repeat)"`
	// Output format
	Output     string `help:"Final output format: json, html, text" default:""`
	OutputFile string `help:"File path for html output (default: basi-report.html)" default:"basi-report.html"`
}

func (r *RunCmd) getParsedTimeout() float64 {
	if r.Timeout == "" {
		return 0.0
	}
	d, err := time.ParseDuration(r.Timeout)
	if err != nil {
		return 0.0
	}
	return float64(d.Milliseconds())
}

func (r *RunCmd) Run(globals *Globals) error {
	fileData, err := os.ReadFile(r.File)
	if err != nil {
		return err
	}

	executor := &playwright.Executor{}
	actions := make([]playwright.ExecutorAction, 0)

	if strings.HasSuffix(r.File, ".basi") {
		parsed, err := basi.Parse(r.File, bytes.NewBuffer(fileData))
		if err != nil {
			return err
		}

		for _, p := range parsed.Actions {
			action := *playwright.NewExecutorAction(p)
			if r.Timeout != "" {
				action.Timeout = r.getParsedTimeout()
			}
			actions = append(actions, action)
		}

		cdpEndpoint := cmp.Or(parsed.GetMetaFieldString("CDPEndpoint"), r.CDPEndpoint)
		if cdpEndpoint != "" {
			if !strings.HasPrefix(cdpEndpoint, "wss://") {
				return fmt.Errorf("invalid CDP Endpoint provided: '%s'", cdpEndpoint)
			}
		}

		headless := parsed.GetMetaFieldString("Headless") == "yes" || globals.Headless
		executor = &playwright.Executor{
			Name:        parsed.GetMetaFieldString("Title"),
			Description: parsed.GetMetaFieldString("Description"),
			URL:         cmp.Or(parsed.GetMetaFieldString("URL"), r.URL),
			Browser:     cmp.Or(parsed.GetMetaFieldString("Browsers"), globals.Browser),
			CDPEndpoint: cdpEndpoint,
			Headless:    headless,
			Actions:     actions,
			Context:     playwright.NewExecutionContext(),
		}

	} else if strings.HasSuffix(r.File, ".yaml") || strings.HasSuffix(r.File, ".yml") {
		if err := yaml.Unmarshal(fileData, executor); err != nil {
			return fmt.Errorf("unable to parse step got: %v", err)
		}

	} else {
		return fmt.Errorf("failed to run, invalid file specified: %s", r.File)
	}

	// Build emitter chain.
	// JSONL progress events go to stderr so stdout stays clean for structured output.
	var baseEmitter core.Emitter
	if r.Quiet || r.Silent {
		baseEmitter = core.NoopEmitter{}
	} else {
		baseEmitter = core.NewJSONLinesEmitter(os.Stderr)
	}

	// When an output mode is requested, wrap with a collector so we can build the report.
	var collector *core.CollectingEmitter
	if r.Output != "" {
		collector = core.NewCollectingEmitter(baseEmitter)
		executor.Emitter = collector
	} else {
		executor.Emitter = baseEmitter
	}

	repeat := max(r.Repeat, 1)

	passed := 0
	failed := 0
	totalStart := time.Now()
	var lastResult playwright.Result

	for i := range repeat {
		executor.RunNumber = i + 1
		iterStart := time.Now()
		raw, runErr := executor.Run(context.Background())
		elapsed := time.Since(iterStart)

		if runErr != nil {
			failed++
			if repeat > 1 {
				fmt.Fprintf(os.Stderr, "[run %d/%d] FAIL (%s) — %v\n", i+1, repeat, elapsed.Round(time.Millisecond), runErr)
			}
			if r.FailFast {
				break
			}
		} else {
			passed++
			if repeat > 1 {
				fmt.Fprintf(os.Stderr, "[run %d/%d] PASS (%s)\n", i+1, repeat, elapsed.Round(time.Millisecond))
			}
			if res, ok := raw.(playwright.Result); ok {
				lastResult = res
			}
		}
	}

	totalElapsed := time.Since(totalStart)

	if repeat > 1 {
		fmt.Fprintf(os.Stderr, "%d runs: %d passed, %d failed\n", repeat, passed, failed)
	}

	// Write structured output if requested.
	if r.Output != "" && collector != nil {
		report := playwright.BuildReport(r.File, passed, failed, collector.Events, totalElapsed)
		switch r.Output {
		case "json":
			if err := playwright.WriteJSONReport(os.Stdout, report); err != nil {
				return fmt.Errorf("failed to write JSON report: %w", err)
			}
		case "html":
			f, err := os.Create(r.OutputFile)
			if err != nil {
				return fmt.Errorf("failed to create HTML report file: %w", err)
			}
			defer f.Close()
			if err := playwright.WriteHTMLReport(f, report); err != nil {
				return fmt.Errorf("failed to write HTML report: %w", err)
			}
			fmt.Fprintf(os.Stderr, "HTML report written to %s\n", r.OutputFile)
		case "text":
			if lastResult.Page != nil {
				fmt.Print(lastResult.Page.Text)
			}
		default:
			return fmt.Errorf("unknown output format %q; valid values: json, html, text", r.Output)
		}
	}

	slog.Debug("running the executor", "url", executor.URL)
	slog.Debug("executed successfully")

	if failed > 0 {
		return fmt.Errorf("%d of %d run(s) failed", failed, repeat)
	}
	return nil
}
