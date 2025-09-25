package main

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/zikani03/basi"
	"github.com/zikani03/basi/playwright"
	"gopkg.in/yaml.v2"
)

type RunCmd struct {
	File      string `arg:"" help:"filename for file to run"`
	Directory string `short:"d" help:"directory containing .basi files to be run"`
	URL       string `short:"u" help:"which url to run the test against"`
	Remote    bool   `help:"whether to run remote test"`
	Docker    bool   `help:"whether to run tests inside docker"`
	Local     bool   `help:"whether to install playwright locally and run tests"`
	OutputDir string `short:"o" help:"Where to write test output and screenshots"`
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
			actions = append(actions, *playwright.NewExecutorAction(p))
		}

		headless := parsed.GetMetaFieldString("Headless") == "yes" || globals.Headless
		executor = &playwright.Executor{
			Name:        parsed.GetMetaFieldString("Title"),
			Description: parsed.GetMetaFieldString("Description"),
			URL:         cmp.Or(parsed.GetMetaFieldString("URL"), r.URL),
			Browser:     cmp.Or(parsed.GetMetaFieldString("Browsers"), globals.Browser),
			Headless:    headless,
			Actions:     actions,
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
