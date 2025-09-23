package docgen

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/repr"
	"github.com/zikani03/basi"
	"github.com/zikani03/basi/playwright"
)

func GenerateStepsFromActions(filename, url string, spec *basi.PlaywrightAction, debug bool) (string, error) {
	actions, err := cloneAndaddScreenShotSteps(filename, spec)
	if err != nil {
		return "", err
	}
	spec.Actions = actions
	if debug {
		repr.Println(actions, repr.Indent("  "), repr.OmitEmpty(true))
	}

	executableActions := make([]playwright.ExecutorAction, 0)
	for _, p := range spec.Actions {
		executableActions = append(executableActions, *playwright.NewExecutorAction(p))
	}
	headless := spec.GetMetaFieldString("Headless") == "yes" // || globals.Headless

	executor := &playwright.Executor{
		Name:        spec.GetMetaFieldString("Title"),
		Description: spec.GetMetaFieldString("Description"),
		URL:         cmp.Or(spec.GetMetaFieldString("URL"), url),
		// Browser:     cmp.Or(spec.GetMetaFieldString("Browsers"), globals.Browser),
		Headless: headless,
		Actions:  executableActions,
	}

	_, err = executor.Run(context.Background())
	if err != nil {
		return "", err
	}
	// create a test file
	f, err := os.Create(filename + "doc.md")
	if err != nil {
		return "", err
	}
	defer f.Close()
	fmt.Fprintf(f, "# Test Document for %s \n\n", spec.GetMetaFieldString("Title"))
	f.WriteString("Documented by Basi\n\n")
	stepNo := 1
	for _, action := range actions {
		if action.Action == "Screenshot" {
			screenshotFileName := action.Arguments.String
			_, screenshotFileRel := filepath.Split(screenshotFileName)
			fmt.Fprintf(f, "![[Figure %d]](%s)\n\n", stepNo, screenshotFileRel)
			continue
		}
		fmt.Fprintf(f, "%d. %s %s\n\n", stepNo, action.Action, selectorToHuman(action.Selector.Selector))
		stepNo += 1
	}
	return filename, nil
}

func selectorToHuman(selector string) string {
	// TODO: parse selector and use heuristics for best case of human representation of it
	return selector
}

func cloneAndaddScreenShotSteps(sourcefile string, spec *basi.PlaywrightAction) ([]*basi.Action, error) {
	newActions := make([]*basi.Action, 0)
	var err error
	dir, filename := filepath.Split(sourcefile)
	specFilenameLower := strings.ReplaceAll(strings.ToLower(filename), ".basi", "")
	if !filepath.IsAbs(dir) {
		dir, err = filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to find directory to place doc and screenshot: %v", err)
		}
	}

	for idx, action := range spec.Actions {
		if action.Action == "" {
			return nil, fmt.Errorf("action cannot be empty or nil at: %d", idx)
		}

		newActions = append(newActions, action)
		// no need to add a screenshot when we are already doing a screenshot step
		if action.Action == "Screenshot" {
			continue
		}
		// just add a screenshot step
		newActions = append(newActions, &basi.Action{
			Action: "Screenshot",
			Selector: &basi.Selector{
				Selector: "body", // how to get the best screenshot element?
			},
			Arguments: &basi.String{
				String: filepath.Join(dir, fmt.Sprintf("%s-step-%d.png", specFilenameLower, idx)),
			},
		})
	}
	return newActions, nil
}
