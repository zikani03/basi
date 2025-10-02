package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/zikani03/basi"
	"github.com/zikani03/basi/internal/experimental/playwrightgen"
)

type GeneratePlaywrightCmd struct {
	File string `arg:"" help:"filename for file to run"`
}

func (r *GeneratePlaywrightCmd) Run(globals *Globals) error {
	fileData, err := os.ReadFile(r.File)
	if err != nil {
		return err
	}
	parsed, err := basi.Parse(r.File, bytes.NewBuffer(fileData))
	if err != nil {
		return err
	}
	filename, err := playwrightgen.Create(r.File, parsed)
	if err != nil {
		return fmt.Errorf("failed to generate file: %v", err)
	}
	fmt.Printf("Generated file at %s", filename)
	return nil
}
