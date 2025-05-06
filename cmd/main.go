package main

import (
	"fmt"
	"os"
	p "v4playwrightactionparser"
)

func main() {
	_, err := p.DebugParse(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse action: %v\n", err)
	}
}
