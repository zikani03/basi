package playwright

import (
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
)

// fakerTokenRe matches tokens like {{faker.email}} (with optional surrounding
// whitespace), capturing the faker type name.
var fakerTokenRe = regexp.MustCompile(`\{\{\s*faker\.([a-zA-Z_]+)\s*\}\}`)

// fakerFuncs maps a (lower-cased) faker type name to a function that produces a
// fresh fake value as a string. Each call regenerates a new value, which is
// what callers rely on for repeated runs.
var fakerFuncs = map[string]func() string{
	"email":     func() string { return gofakeit.Email() },
	"name":      func() string { return gofakeit.Name() },
	"firstname": func() string { return gofakeit.FirstName() },
	"lastname":  func() string { return gofakeit.LastName() },
	"username":  func() string { return gofakeit.Username() },
	"phone":     func() string { return gofakeit.Phone() },
	"uuid":      func() string { return gofakeit.UUID() },
	"word":      func() string { return gofakeit.Word() },
	"sentence":  func() string { return gofakeit.Sentence(6) },
	"int":       func() string { return strconv.Itoa(gofakeit.Number(0, 1000000)) },
	"number":    func() string { return strconv.Itoa(gofakeit.Number(0, 1000000)) },
	"digit":     func() string { return gofakeit.Digit() },
	"date":      func() string { return gofakeit.Date().Format("2006-01-02") },
	"url":       func() string { return gofakeit.URL() },
	"bool":      func() string { return strconv.FormatBool(gofakeit.Bool()) },
	"company":   func() string { return gofakeit.Company() },
	"city":      func() string { return gofakeit.City() },
	"country":   func() string { return gofakeit.Country() },
	"zip":       func() string { return gofakeit.Zip() },
	"street":    func() string { return gofakeit.Street() },
	"price":     func() string { return strconv.FormatFloat(gofakeit.Price(1, 1000), 'f', 2, 64) },
}

// interpolateFaker replaces {{faker.<type>}} tokens in s with freshly generated
// fake values. Unknown faker types are left untouched (same convention as
// unknown $var references) and a warning is logged.
func interpolateFaker(s string) string {
	if !strings.Contains(s, "{{") {
		return s
	}
	return fakerTokenRe.ReplaceAllStringFunc(s, func(match string) string {
		sub := fakerTokenRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		typeName := strings.ToLower(sub[1])
		if fn, ok := fakerFuncs[typeName]; ok {
			return fn()
		}
		slog.Warn("unknown faker type, leaving token unchanged", "token", match)
		return match
	})
}
