package playwrightgen

import (
	"cmp"
	"errors"
	"fmt"
	"os"

	"github.com/zikani03/basi"
)

func Create(filename string, spec *basi.PlaywrightAction) (string, error) {
	file, err := os.Create(filename + ".js")
	if err != nil {
		return "", fmt.Errorf("could not create .js file: %v", err)
	}
	defer file.Close()
	_, err = file.WriteString("import { test } from '@playwright/test';\n")
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %v", err)
	}

	_, err = fmt.Fprintf(file, "test( '%s', async ({ page }) => {\n", cmp.Or(spec.GetMetaFieldString("Title"), filename))
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %v", err)
	}

	for _, action := range spec.Actions {
		actionFunc, ok := actionMap[action.Action]
		if !ok {
			return "", fmt.Errorf("action %s not supported", action.Action)
		}
		err = actionFunc(file, action)
		if errors.Is(err, ErrNotAPlaywrightAction) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("failed to write to file: %v", err)
		}
	}

	_, err = file.WriteString("\n});\n")
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %v", err)
	}
	return file.Name(), nil
}
