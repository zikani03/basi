package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/zikani03/basi"
)

type TestCmd struct {
	File string `arg:"" help:"File to test"`
}

func (c *TestCmd) Run(globals *Globals) error {
	fileData, err := os.ReadFile(c.File)
	if err != nil {
		return err
	}
	parsed, err := basi.DebugParse(c.File, bytes.NewBuffer(fileData))
	if err != nil {
		return err
	}
	fmt.Printf("parsed %v", parsed)
	return nil
}
