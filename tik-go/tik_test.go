package tik_test

import (
	"errors"
	"os"
	"reflect"
	"testing"

	tik "github.com/romshark/tik/tik-go"
)

func TestParse(t *testing.T) {
	t.Parallel()

	parser := tik.NewParser(nil) // With default config.
	f := func(t *testing.T, input string, expect tik.TIK) {
		t.Helper()
		got, err := parser.Parse(input)
		requireNoErr(t, err)
		reflect.DeepEqual(expect, got)
	}

	// Empty string literal only
	f(t, "", tik.TIK{})

	// String literal only
	f(t, "hello world", tik.TIK{Tokens: tik.Tokens{
		tik.Token{
			Str:   "hello world",
			Index: 0,
			Type:  tik.TokenTypeStringLiteral,
		},
	}})

	// Number placeholder
	f(t, "{3} items", tik.TIK{Tokens: tik.Tokens{
		{Str: "{3}", Index: 0, Type: tik.TokenTypeNumber},
		{Str: " items", Index: 3, Type: tik.TokenTypeStringLiteral},
	}})

	// String placeholders
	f(t, `{"first"} {"second"}, {"_"}{"fourth"}`, tik.TIK{Tokens: tik.Tokens{
		{Str: `{"first"}`, Index: 0, Type: tik.TokenTypeStringPlaceholder},
		{Str: " ", Index: 9, Type: tik.TokenTypeStringLiteral},
		{Str: `{"second"}`, Index: 10, Type: tik.TokenTypeStringPlaceholder},
		{Str: ", ", Index: 21, Type: tik.TokenTypeStringLiteral},
		{Str: `{"_"}`, Index: 22, Type: tik.TokenTypeStringPlaceholder},
		{Str: `{"fourth"}`, Index: 27, Type: tik.TokenTypeStringPlaceholder},
	}})

	// Gender agreement
	f(t, "{he} lost {himself} in {his} thoughts", tik.TIK{Tokens: tik.Tokens{
		{Str: "{he}", Index: 0, Type: tik.TokenTypeGenderAgreement},
		{Str: " lost ", Index: 4, Type: tik.TokenTypeStringLiteral},
		{Str: "{himself}", Index: 10, Type: tik.TokenTypeGenderAgreement},
		{Str: " in ", Index: 19, Type: tik.TokenTypeStringLiteral},
		{Str: "{his}", Index: 23, Type: tik.TokenTypeGenderAgreement},
		{Str: " thoughts", Index: 28, Type: tik.TokenTypeStringLiteral},
	}})
	// Pluralization
	f(t, "You're {4th} out of {2} contenders", tik.TIK{Tokens: tik.Tokens{
		{Str: "You're ", Index: 0, Type: tik.TokenTypeStringLiteral},
		{Str: "{4th}", Index: 7, Type: tik.TokenTypeOrdinalPlural},
		{Str: " out of ", Index: 12, Type: tik.TokenTypeStringLiteral},
		{Str: "{2}", Index: 20, Type: tik.TokenTypeCardinalPlural},
		{Str: " contenders", Index: 23, Type: tik.TokenTypeStringLiteral},
	}})

	// Time
	f(t, `{3:45PM}{3:45:30PM}{April 2}{Apr 2}
		{Apr 2025}{Monday}{April 2, 3:45PM}
		{2025}{April 2, 3:45:30PM}`, tik.TIK{Tokens: tik.Tokens{
		{Str: "{3:45PM}", Index: 0, Type: tik.TokenTypeTimeShort},
		{Str: "{3:45:30PM}", Index: 8, Type: tik.TokenTypeTimeShortSeconds},
		{Str: "{April 2}", Index: 19, Type: tik.TokenTypeTimeFullMonthAndDay},
		{Str: "{Apr 2}", Index: 28, Type: tik.TokenTypeTimeShortMonthAndDay},
		{Str: "\n\t\t", Index: 35, Type: tik.TokenTypeStringLiteral},
		{Str: "{Apr 2025}", Index: 38, Type: tik.TokenTypeTimeFullMonthAndYear},
		{Str: "{Monday}", Index: 48, Type: tik.TokenTypeTimeWeekday},
		{Str: "{April 2, 3:45PM}", Index: 56, Type: tik.TokenTypeTimeDateAndShort},
		{Str: "\n\t\t", Index: 73, Type: tik.TokenTypeStringLiteral},
		{Str: "{2025}", Index: 76, Type: tik.TokenTypeTimeYear},
		{Str: "{April 2, 3:45:30PM}", Index: 82, Type: tik.TokenTypeTimeFull},
	}})

	// Escaped brace.
	f(t, `\{not a placeholder}`, tik.TIK{Tokens: tik.Tokens{
		{Str: `{not a placeholder}`, Index: 0, Type: tik.TokenTypeStringLiteral},
	}})

	// Escaped reverse solidus.
	f(t, `\\text after`, tik.TIK{Tokens: tik.Tokens{
		{Str: `\text after`, Index: 0, Type: tik.TokenTypeStringLiteral},
	}})
	// Escaped reverse solidus.
	f(t, `\ntext after\n\t\\\n`, tik.TIK{Tokens: tik.Tokens{
		{Str: `\ttext after\n\t\\\n`, Index: 0, Type: tik.TokenTypeStringLiteral},
	}})

	// Sequence of escaped reverse solidus.
	f(t, `before \\\\{Monday} after`, tik.TIK{Tokens: tik.Tokens{
		{Str: `before \\`, Index: 0, Type: tik.TokenTypeStringLiteral},
		{Str: `{Monday}`, Index: 11, Type: tik.TokenTypeTimeWeekday},
		{Str: ` after`, Index: 19, Type: tik.TokenTypeStringLiteral},
	}})

	// Case insensitivity.
	f(t, `{He}{HIMSELF}{3:45pm}{3:45:30pM}{april 2}{mOnDaY}`, tik.TIK{Tokens: tik.Tokens{
		{Str: "{He}", Index: 0, Type: tik.TokenTypeGenderAgreement},
		{Str: "{HIMSELF}", Index: 4, Type: tik.TokenTypeGenderAgreement},
		{Str: "{3:45pm}", Index: 13, Type: tik.TokenTypeTimeShort},
		{Str: "{3:45:30pM}", Index: 21, Type: tik.TokenTypeTimeShortSeconds},
		{Str: "{april 2}", Index: 32, Type: tik.TokenTypeTimeFullMonthAndDay},
		{Str: "{mOnDaY}", Index: 41, Type: tik.TokenTypeTimeWeekday},
	}})
}

func TestParseErr(t *testing.T) {
	t.Parallel()

	parser := tik.NewParser(nil) // With default config.

	f := func(t *testing.T, expectErr error, input string) {
		t.Helper()
		tk, err := parser.Parse(input)
		requireErrIs(t, expectErr, err)
		requireDeepEqual(t, tik.TIK{}, tk)
	}

	f(t, tik.ErrUnclosedPlaceholder, `unexpected EOF: {`)
	f(t, tik.ErrUnclosedStringPlaceholder, `unexpected EOF: {"`)
	f(t, tik.ErrUnclosedStringPlaceholder, `unexpected EOF: {"_`)
	f(t, tik.ErrUnclosedStringPlaceholder, `unexpected EOF: {""`)
	f(t, tik.ErrUnclosedStringPlaceholder, `unexpected EOF: {"abc }`)
	f(t, tik.ErrUnclosedPlaceholder, `unexpected EOF: {x`)
	f(t, tik.ErrUnclosedPlaceholder, `unexpected EOF: {{`)
	f(t, tik.ErrStringPlaceholderEmpty, `this is illegal: {""}`)
	f(t, tik.ErrStringPlaceholderInvSpace, `this too: {"  "}`)
	f(t, tik.ErrStringPlaceholderInvSpace, `this too: {" text "}`)
	f(t, tik.ErrStringPlaceholderInvSpace, `this too: {"  text  "}`)
	f(t, tik.ErrStringPlaceholderInvSpace, "this too: {\"\u3000text\"}")
	f(t, tik.ErrStringPlaceholderInvSpace, "this too: {\"text\u3000\"}")
	f(t, tik.ErrStringPlaceholderIllegalChars, `and this: {"\""} text after`)
	f(t, tik.ErrStringPlaceholderIllegalChars, `{"\"}`)
	f(t, tik.ErrStringPlaceholderIllegalChars, `{"abc \n def"}`)
}

func TestParseCustomPlaceholders(t *testing.T) {
	t.Parallel()

	m, err := tik.CustomPlaceholders(func(m map[string]tik.TokenType) {
		delete(m, "3")
		m["43"] = tik.TokenTypeNumber
	})
	requireNoErr(t, err)
	parser := tik.NewParser(&tik.Config{Mapping: m})

	got, err := parser.Parse("{43}{43}")
	requireNoErr(t, err)
	requireDeepEqual(t, tik.TIK{Tokens: tik.Tokens{
		tik.Token{Str: "{43}", Index: 0, Type: tik.TokenTypeNumber},
		tik.Token{Str: "{43}", Index: 4, Type: tik.TokenTypeNumber},
	}}, got)

	got, err = parser.Parse("invalid: {3}")
	requireErrIs(t, tik.ErrUknownPlaceholder, err)
	requireEqual(t, "unknown placeholder at index 9", err.Error())
	requireDeepEqual(t, tik.TIK{}, got)
}

func TestCustomPlaceholdersErrInvalid(t *testing.T) {
	t.Parallel()

	m, err := tik.CustomPlaceholders(func(m map[string]tik.TokenType) {
		delete(m, "3")
		m[`"43"`] = tik.TokenTypeNumber
	})
	requireErrIs(t, tik.ErrInvalidCustomPlaceholder, err)
	requireDeepEqual(t, nil, m)
}

func TestCustomPlaceholdersErrInvalidCustomPlaceholders(t *testing.T) {
	t.Parallel()

	m, err := tik.CustomPlaceholders(func(m map[string]tik.TokenType) {
		delete(m, "3")
	})
	requireErrIs(t, tik.ErrInvalidCustomPlaceholders, err)
	requireDeepEqual(t, nil, m)

	m, err = tik.CustomPlaceholders(func(m map[string]tik.TokenType) {
		delete(m, "3")
		m[`"3"`] = tik.TokenTypeNumber
	})
	requireErrIs(t, tik.ErrInvalidCustomPlaceholder, err)
	requireDeepEqual(t, nil, m)
}

func TestTokenizeErr(t *testing.T) {
	t.Parallel()

	var tokenizer tik.Tokenizer

	f := func(t *testing.T, input string, expectErrMsg string) {
		t.Helper()
		got, err := tokenizer.Tokenize(nil, input, nil)
		requireEqual(t, expectErrMsg, err.Error())
		requireDeepEqual(t, tik.Tokens(nil), got)
	}

	// String literal only.
	f(t, "hello world {", "unclosed placeholder at index 12")
	f(t, "{unknown}", "unknown placeholder at index 0")
}

func TestTIKPlaceholdersIter(t *testing.T) {
	t.Parallel()

	var tokenizer tik.Tokenizer

	toks, err := tokenizer.Tokenize(nil, `{3:45PM}{3:45:30PM}{April 2}{Apr 2}
		{Apr 2025}{Monday}{April 2, 3:45PM}
		{2025}{April 2, 3:45:30PM}`, nil)
	requireNoErr(t, err)
	tk := tik.TIK{Tokens: toks}

	expect := []tik.Token{
		tk.Tokens[0],
		tk.Tokens[1],
		tk.Tokens[2],
		tk.Tokens[3],
		// 4 is a string literal.
		tk.Tokens[5],
		tk.Tokens[6],
		tk.Tokens[7],
		// 8 is a string literal.
		tk.Tokens[9],
		tk.Tokens[10],
	}

	var actual []tik.Token
	for i, tok := range tk.Placeholders() {
		requireEqual(t, i, len(actual))
		actual = append(actual, tok)
	}

	requireDeepEqual(t, expect, actual)

	// Test break
	{
		counter := 0
		for range tk.Placeholders() {
			counter++
			break
		}
		requireEqual(t, 1, counter)
	}
}

func FuzzTokenize(f *testing.F) {
	f.Add("")
	f.Add(`hello world`)
	f.Add(`{3} items`)
	f.Add(`{he} lost {himself} in {his} thoughts`)
	f.Add(`\n`)
	f.Add(`\{not a placeholder}\{again, not a placeholder}`)
	f.Add(`\\text after`)
	f.Add(`\\\\text after`)
	f.Add("You're {4th} out of {2} contenders")
	f.Add("{unknown}")
	f.Add(`{3:45PM}{3:45:30PM}{April 2}{Apr 2}
		{Apr 2025}{Monday}{April 2, 3:45PM}
		{2025}{April 2, 3:45:30PM}`)

	f.Fuzz(func(t *testing.T, input string) {
		var tokenizer tik.Parser
		tk, err := tokenizer.Parse(input)
		// If an error occurs, ensure it's one of the expected error types.
		if err != nil {
			_ = err.Error()
			return
		}
		for range tk.Placeholders() {
			// Just iterate to ensure it doesn't panic.
		}
	})
}

func BenchmarkParseFnPlaceholdersOnly(b *testing.B) {
	parser := tik.NewParser(nil) // With default config.
	for b.Loop() {
		err := parser.ParseFn(`{3:45PM}{3:45:30PM}{April 2}{Apr 2}
		{Apr 2025}{Monday}{April 2, 3:45PM}
		{2025}{April 2, 3:45:30PM}`, func(_ tik.TIK) error { return nil })
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkParseFnFewPlaceholders(b *testing.B) {
	parser := tik.NewParser(nil) // With default config.
	var err error

	loremIpsum, err := os.ReadFile("testdata/lorem_ipsum.txt")
	requireNoErr(b, err)
	input := string(loremIpsum)

	for b.Loop() {
		err := parser.ParseFn(input, func(_ tik.TIK) error { return nil })
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkParseFnNoPlaceholders(b *testing.B) {
	parser := tik.NewParser(nil) // With default config.

	loremIpsum, err := os.ReadFile("testdata/lorem_ipsum_fewplaceholders.txt")
	requireNoErr(b, err)
	input := string(loremIpsum)

	for b.Loop() {
		err := parser.ParseFn(input, func(_ tik.TIK) error { return nil })
		if err != nil {
			panic(err)
		}
	}
}

func requireDeepEqual[T any](tb testing.TB, expect, actual T) {
	tb.Helper()
	if !reflect.DeepEqual(expect, actual) {
		tb.Fatalf("expected %#v; received: %#v", expect, actual)
	}
}

func requireEqual[T comparable](tb testing.TB, expect, actual T) {
	tb.Helper()
	if expect != actual {
		tb.Fatalf("expected %#v; received: %#v", expect, actual)
	}
}

func requireErrIs(tb testing.TB, expect, actual error) {
	tb.Helper()
	if !errors.Is(actual, expect) {
		tb.Fatalf("expected %#v; received: %#v", expect, actual)
	}
}

func requireNoErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("expected no error; received: %#v", err)
	}
}
