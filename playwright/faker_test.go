package playwright

import (
	"strconv"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
)

func TestInterpolateFakerKnownTokens(t *testing.T) {
	email := interpolateFaker("{{faker.email}}")
	if !strings.Contains(email, "@") {
		t.Errorf("expected faker.email to contain '@', got %q", email)
	}

	uuid := interpolateFaker("{{faker.uuid}}")
	if uuid == "" || uuid == "{{faker.uuid}}" {
		t.Errorf("expected faker.uuid to be a non-empty resolved value, got %q", uuid)
	}

	intVal := interpolateFaker("{{faker.int}}")
	if _, err := strconv.Atoi(intVal); err != nil {
		t.Errorf("expected faker.int to parse as int, got %q (%v)", intVal, err)
	}

	// Token embedded in a larger string with surrounding whitespace variant.
	name := interpolateFaker("Hello {{ faker.name }}!")
	if name == "Hello {{ faker.name }}!" || !strings.HasPrefix(name, "Hello ") || !strings.HasSuffix(name, "!") {
		t.Errorf("expected faker.name token to be replaced in context, got %q", name)
	}
}

func TestInterpolateFakerUnknownTokenUnchanged(t *testing.T) {
	in := "{{faker.notareal}}"
	out := interpolateFaker(in)
	if out != in {
		t.Errorf("expected unknown faker token to be left unchanged, got %q", out)
	}
}

func TestInterpolateFakerNoTokens(t *testing.T) {
	in := "input[name=email]"
	if out := interpolateFaker(in); out != in {
		t.Errorf("expected string without tokens to be unchanged, got %q", out)
	}
}

func TestInterpolateFakerDeterministicWithSeed(t *testing.T) {
	gofakeit.Seed(1)
	first := interpolateFaker("{{faker.email}}")

	gofakeit.Seed(1)
	second := interpolateFaker("{{faker.email}}")

	if first != second {
		t.Errorf("expected deterministic output with same seed, got %q and %q", first, second)
	}

	// A different seed should (very likely) produce a different value.
	gofakeit.Seed(2)
	third := interpolateFaker("{{faker.email}}")
	if third == first {
		t.Errorf("expected different seed to produce a different value, both were %q", third)
	}
}

func TestInterpolateCombinesFakerAndVars(t *testing.T) {
	ec := NewExecutionContext()
	ec.Variables.Set("username", "alice")

	out := ec.Interpolate("user=$username mail={{faker.email}}")
	if !strings.Contains(out, "user=alice ") {
		t.Errorf("expected $var to be resolved, got %q", out)
	}
	if !strings.Contains(out, "@") {
		t.Errorf("expected faker.email to be resolved, got %q", out)
	}
	if strings.Contains(out, "{{") || strings.Contains(out, "$username") {
		t.Errorf("expected all tokens resolved, got %q", out)
	}
}

func TestInterpolateUnknownVarUnchanged(t *testing.T) {
	ec := NewExecutionContext()
	out := ec.Interpolate("$missing")
	if out != "$missing" {
		t.Errorf("expected unknown $var to be left unchanged, got %q", out)
	}
}

func TestInterpolateNilSafe(t *testing.T) {
	var ec *ExecutionContext
	// Nil context should still resolve faker tokens and not panic on the $var pass.
	out := ec.Interpolate("{{faker.uuid}} $x")
	if strings.Contains(out, "{{") {
		t.Errorf("expected faker token resolved even with nil context, got %q", out)
	}
	if !strings.Contains(out, "$x") {
		t.Errorf("expected $x left unchanged with nil Variables, got %q", out)
	}
}
