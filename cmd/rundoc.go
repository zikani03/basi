package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/zikani03/basi"
	"github.com/zikani03/basi/internal/experimental/docgen"
)

type RunDocCmd struct {
	RunCmd
}

func (c *RunDocCmd) Run(globals *Globals) error {
	fileData, err := os.ReadFile(c.File)
	if err != nil {
		return err
	}
	parsed, err := basi.Parse(c.File, bytes.NewBuffer(fileData))
	if err != nil {
		return err
	}
	filename, err := docgen.GenerateStepsFromActions(c.File, c.URL, parsed, globals.Debug)
	if err != nil {
		return fmt.Errorf("failed to generate document: %v", err)
	}
	fmt.Printf("Generated file at %s", filename)
	return nil
}
