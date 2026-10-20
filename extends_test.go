package basi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile writes content to path, creating parent directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

// parseAndResolve parses the file at path and resolves its Extends chain,
// returning the resolved PlaywrightAction.
func parseAndResolve(t *testing.T, path string) *PlaywrightAction {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	parsed, err := Parse(path, strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("failed to parse %s: %v", path, err)
	}
	if err := ResolveExtends(parsed, path); err != nil {
		t.Fatalf("ResolveExtends failed for %s: %v", path, err)
	}
	return parsed
}

func assertField(t *testing.T, p *PlaywrightAction, name, want string) {
	t.Helper()
	if got := p.GetMetaFieldString(name); got != want {
		t.Errorf("field %q: got %q, want %q", name, got, want)
	}
}

func TestExtendsInheritsMissingField(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base.basi")
	child := filepath.Join(dir, "child.basi")

	writeFile(t, base, `Browser     : "chromium"
Headless    : "yes"
---
Goto "https://example.com"
`)
	writeFile(t, child, `Title       : "Child"
Extends     : "./base.basi"
---
Goto "https://example.com"
`)

	p := parseAndResolve(t, child)
	assertField(t, p, "Title", "Child")
	assertField(t, p, "Browser", "chromium") // inherited
	assertField(t, p, "Headless", "yes")     // inherited
}

func TestExtendsLocalWins(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base.basi")
	child := filepath.Join(dir, "child.basi")

	writeFile(t, base, `Browser     : "firefox"
---
Goto "https://example.com"
`)
	writeFile(t, child, `Browser     : "chromium"
Extends     : "./base.basi"
---
Goto "https://example.com"
`)

	p := parseAndResolve(t, child)
	assertField(t, p, "Browser", "chromium") // local wins, not firefox
}

func TestExtendsSectionFilter(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base.basi")
	child := filepath.Join(dir, "child.basi")

	writeFile(t, base, `Browser     : "chromium"
Headless    : "yes"
Description : "base description"
---
Goto "https://example.com"
`)
	writeFile(t, child, `Title       : "Child"
Extends     : "./base.basi;Browser"
---
Goto "https://example.com"
`)

	p := parseAndResolve(t, child)
	assertField(t, p, "Browser", "chromium")  // inherited (named)
	assertField(t, p, "Headless", "")         // NOT inherited
	assertField(t, p, "Description", "")       // NOT inherited
}

func TestExtendsNoSemicolonInheritsAll(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base.basi")
	child := filepath.Join(dir, "child.basi")

	writeFile(t, base, `Browser     : "chromium"
Headless    : "yes"
Description : "base description"
---
Goto "https://example.com"
`)
	writeFile(t, child, `Title       : "Child"
Extends     : "./base.basi"
---
Goto "https://example.com"
`)

	p := parseAndResolve(t, child)
	assertField(t, p, "Browser", "chromium")
	assertField(t, p, "Headless", "yes")
	assertField(t, p, "Description", "base description")
}

func TestExtendsCircularReference(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.basi")
	b := filepath.Join(dir, "b.basi")

	writeFile(t, a, `Title       : "A"
Extends     : "./b.basi"
---
Goto "https://example.com"
`)
	writeFile(t, b, `Browser     : "chromium"
Extends     : "./a.basi"
---
Goto "https://example.com"
`)

	data, err := os.ReadFile(a)
	if err != nil {
		t.Fatalf("read a: %v", err)
	}
	parsed, err := Parse(a, strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("parse a: %v", err)
	}
	err = ResolveExtends(parsed, a)
	if err == nil {
		t.Fatalf("expected circular reference error, got nil")
	}
	if !strings.Contains(err.Error(), "circular Extends reference detected") {
		t.Errorf("expected circular reference error, got: %v", err)
	}
}

func TestExtendsRelativePathFromSubdirectory(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base.basi")
	// child lives in a nested subdirectory and references the base one level up.
	child := filepath.Join(dir, "nested", "sub", "child.basi")

	writeFile(t, base, `Browser     : "webkit"
---
Goto "https://example.com"
`)
	writeFile(t, child, `Title       : "Nested Child"
Extends     : "../../base.basi"
---
Goto "https://example.com"
`)

	p := parseAndResolve(t, child)
	assertField(t, p, "Title", "Nested Child")
	assertField(t, p, "Browser", "webkit") // inherited via relative path
}

func TestExtendsGrandparentChain(t *testing.T) {
	dir := t.TempDir()
	grand := filepath.Join(dir, "grand.basi")
	parent := filepath.Join(dir, "parent.basi")
	child := filepath.Join(dir, "child.basi")

	writeFile(t, grand, `Description : "from grandparent"
---
Goto "https://example.com"
`)
	writeFile(t, parent, `Browser     : "chromium"
Extends     : "./grand.basi"
---
Goto "https://example.com"
`)
	writeFile(t, child, `Title       : "Child"
Extends     : "./parent.basi"
---
Goto "https://example.com"
`)

	p := parseAndResolve(t, child)
	assertField(t, p, "Title", "Child")
	assertField(t, p, "Browser", "chromium")         // from parent
	assertField(t, p, "Description", "from grandparent") // flows down
}

func TestExtendsNoMetaIsNoop(t *testing.T) {
	parsed, err := Parse("x.basi", strings.NewReader(`Goto "https://example.com"`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := ResolveExtends(parsed, "x.basi"); err != nil {
		t.Errorf("expected no-op, got error: %v", err)
	}
}
