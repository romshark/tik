package tik_test

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"testing"

	tik "github.com/romshark/tik/tik-go"
)

type Token struct {
	Str  string
	Type tik.TokenType
}

func ToTestTokens(input string, toks tik.Tokens) []Token {
	if len(toks) == 0 {
		return nil
	}
	t := make([]Token, len(toks))
	for i, tok := range toks {
		t[i] = Token{
			Str:  tok.String(input),
			Type: tok.Type,
		}
	}
	return t
}

func TestParse(t *testing.T) {
	t.Parallel()

	parser := tik.NewParser(tik.DefaultConfig)
	f := func(t *testing.T, input string, expect ...Token) {
		t.Helper()
		got, err := parser.Parse(input)
		requireNoErr(t, err)
		actual := ToTestTokens(input, got.Tokens)
		requireDeepEqual(t, expect, actual)
	}

	// String literal only
	f(t, "hello world",
		Token{"hello world", tik.TokenTypeLiteral},
	)
	f(t, "  hello world  ",
		Token{"hello world", tik.TokenTypeLiteral},
	)

	// Context.
	f(t, "[c] hello world",
		Token{"[c]", tik.TokenTypeContext},
		Token{"hello world", tik.TokenTypeLiteral},
	)
	f(t, "\r\n\t [c]\t\r\n hello world\r\n\t ",
		Token{"[c]", tik.TokenTypeContext},
		Token{"hello world", tik.TokenTypeLiteral},
	)
	f(t, "[c][b]okay",
		Token{"[c]", tik.TokenTypeContext},
		Token{"[b]okay", tik.TokenTypeLiteral},
	)

	// Integer placeholders
	f(t, "integer: {integer}",
		Token{"integer: ", tik.TokenTypeLiteral},
		Token{"{integer}", tik.TokenTypeInteger},
	)
	f(t, "  {integer} suffix  ",
		Token{"{integer}", tik.TokenTypeInteger},
		Token{" suffix", tik.TokenTypeLiteral},
	)
	f(t, "[context]{integer} suffix",
		Token{"[context]", tik.TokenTypeContext},
		Token{"{integer}", tik.TokenTypeInteger},
		Token{" suffix", tik.TokenTypeLiteral},
	)
	f(t, "  [context]  {integer} suffix  ",
		Token{"[context]", tik.TokenTypeContext},
		Token{"{integer}", tik.TokenTypeInteger},
		Token{" suffix", tik.TokenTypeLiteral},
	)

	// Number placeholders
	f(t, "{number} suffix",
		Token{"{number}", tik.TokenTypeNumber},
		Token{" suffix", tik.TokenTypeLiteral},
	)
	f(t, "  {number} suffix  ",
		Token{"{number}", tik.TokenTypeNumber},
		Token{" suffix", tik.TokenTypeLiteral},
	)
	f(t, "[context]{number} suffix",
		Token{"[context]", tik.TokenTypeContext},
		Token{"{number}", tik.TokenTypeNumber},
		Token{" suffix", tik.TokenTypeLiteral},
	)
	f(t, "  [context]  {number} suffix  ",
		Token{"[context]", tik.TokenTypeContext},
		Token{"{number}", tik.TokenTypeNumber},
		Token{" suffix", tik.TokenTypeLiteral},
	)

	// Text placeholders
	f(t, `{text} {text}, {text}{text}`,
		Token{`{text}`, tik.TokenTypeText},
		Token{" ", tik.TokenTypeLiteral},
		Token{`{text}`, tik.TokenTypeText},
		Token{", ", tik.TokenTypeLiteral},
		Token{`{text}`, tik.TokenTypeText},
		Token{`{text}`, tik.TokenTypeText},
	)

	// Text placeholders with gender
	f(t, "{name}",
		Token{"{name}", tik.TokenTypeTextWithGender},
	)
	f(t, "{name}",
		Token{"{name}", tik.TokenTypeTextWithGender},
	)
	f(t, "Welcome, {name}",
		Token{"Welcome, ", tik.TokenTypeLiteral},
		Token{"{name}", tik.TokenTypeTextWithGender},
	)

	// Pluralization.
	f(t, "You're {ordinal} out of {# contenders}",
		Token{"You're ", tik.TokenTypeLiteral},
		Token{"{ordinal}", tik.TokenTypeOrdinalPlural},
		Token{" out of ", tik.TokenTypeLiteral},
		Token{"{# ", tik.TokenTypeCardinalPluralStart},
		Token{"contenders", tik.TokenTypeLiteral},
		Token{"}", tik.TokenTypeCardinalPluralEnd},
	)
	f(t, `{# files are being copied to folder '{text}'}`,
		Token{"{# ", tik.TokenTypeCardinalPluralStart},
		Token{"files are being copied to folder '", tik.TokenTypeLiteral},
		Token{`{text}`, tik.TokenTypeText},
		Token{"'", tik.TokenTypeLiteral},
		Token{"}", tik.TokenTypeCardinalPluralEnd},
	)
	f(t, `{# messages from {text}} at {time-medium} on {date-short}`,
		Token{"{# ", tik.TokenTypeCardinalPluralStart},
		Token{"messages from ", tik.TokenTypeLiteral},
		Token{`{text}`, tik.TokenTypeText},
		Token{"}", tik.TokenTypeCardinalPluralEnd},
		Token{" at ", tik.TokenTypeLiteral},
		Token{"{time-medium}", tik.TokenTypeTimeMedium},
		Token{" on ", tik.TokenTypeLiteral},
		Token{"{date-short}", tik.TokenTypeDateShort},
	)
	f(t, `{# new files in}{# folders}`,
		Token{"{# ", tik.TokenTypeCardinalPluralStart},
		Token{"new files in", tik.TokenTypeLiteral},
		Token{"}", tik.TokenTypeCardinalPluralEnd},
		Token{"{# ", tik.TokenTypeCardinalPluralStart},
		Token{`folders`, tik.TokenTypeLiteral},
		Token{"}", tik.TokenTypeCardinalPluralEnd},
	)

	// Date and time.
	f(t, `{date-full}{date-long}{date-medium}{date-short}
		{time-full}{time-long}{time-medium}{time-short}`,
		Token{"{date-full}", tik.TokenTypeDateFull},
		Token{"{date-long}", tik.TokenTypeDateLong},
		Token{"{date-medium}", tik.TokenTypeDateMedium},
		Token{"{date-short}", tik.TokenTypeDateShort},
		Token{"\n\t\t", tik.TokenTypeLiteral},
		Token{"{time-full}", tik.TokenTypeTimeFull},
		Token{"{time-long}", tik.TokenTypeTimeLong},
		Token{"{time-medium}", tik.TokenTypeTimeMedium},
		Token{"{time-short}", tik.TokenTypeTimeShort},
	)

	// Escaped braces.
	f(t, `\{not a placeholder\}`,
		Token{`{not a placeholder}`, tik.TokenTypeLiteral},
	)
	f(t, `\\\{not a placeholder\\\}`,
		Token{`\{not a placeholder\}`, tik.TokenTypeLiteral},
	)
	f(t, `\\text after`,
		Token{`\text after`, tik.TokenTypeLiteral},
	)
	f(t, `\ntext after\n\t\\\n`,
		Token{`\ntext after\n\t\\n`, tik.TokenTypeLiteral},
	)

	// Sequence of escaped reverse solidus.
	f(t, `before \\\\{time-medium} after`,
		Token{`before \\`, tik.TokenTypeLiteral},
		Token{`{time-medium}`, tik.TokenTypeTimeMedium},
		Token{` after`, tik.TokenTypeLiteral},
	)
}

func TestParseErr(t *testing.T) {
	t.Parallel()

	parser := tik.NewParser(tik.DefaultConfig)

	f := func(t *testing.T, expectErr error, input string) {
		t.Helper()
		tk, err := parser.Parse(input)
		requireErrIs(t, expectErr, err)
		requireDeepEqual(t, tik.TIK{}, tk)
	}

	f(t, tik.ErrTextEmpty, ``)
	f(t, tik.ErrTextEmpty, `   `)
	f(t, tik.ErrTextEmpty, `[context]`)
	f(t, tik.ErrTextEmpty, `[context]   `)
	f(t, tik.ErrTextEmpty, "\t\r\n ")
	f(t, tik.ErrContextEmpty, `[] Text`)
	f(t, tik.ErrContextEmpty, "[]")
	f(t, tik.ErrContextEmpty, `[  ] Text`)
	f(t, tik.ErrContextEmpty, "[\r\n\t ] Text")
	f(t, tik.ErrContextInvalid, `[not escaped\] Text`)
	f(t, tik.ErrContextInvalid, `[{invalid}] Text`)
	f(t, tik.ErrContextInvalid, `[{] Text`)
	f(t, tik.ErrContextInvalid, `[}] Text`)
	f(t, tik.ErrContextInvalid, `[[]] Text`)
	f(t, tik.ErrContextInvalid, `[[nope]] Text`)
	f(t, tik.ErrContextInvalid, `[a[b]c] Text`)
	f(t, tik.ErrContextInvalid, `[a\[b\]c] Text`)
	f(t, tik.ErrContextUnclosed, "[")
	f(t, tik.ErrContextUnclosed, "[abc")
	f(t, tik.ErrContextUnclosed, "[\t\r\n ")
	f(t, tik.ErrUknownPlaceholder, `no space after cardinal plural: {#abc}`)
	f(t, tik.ErrUknownPlaceholder, `unknown placeholder: {#026}`)
	f(t, tik.ErrUknownPlaceholder, `unknown placeholder: {April 21}`)
	f(t, tik.ErrUknownPlaceholder, `unknown placeholder: {8/16/99}`)
	f(t, tik.ErrUnclosedPlaceholder, `unexpected EOF: {`)
	f(t, tik.ErrUnclosedPlaceholder, `unexpected EOF: {x`)
	f(t, tik.ErrUnclosedPlaceholder, `unexpected EOF: {{`)
	f(t, tik.ErrNestedPluralization, `nested pluralization: {# messages in {# folders}}`)
	f(t, tik.ErrCardinalPluralEmpty, `empty pluralization: {# }`)
	f(t, tik.ErrDirectiveStartsCardinalPlural,
		`illegal pluralization: {# {date-full}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {# {date-full}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {# {date-long}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {# {date-medium}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {# {date-short}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {# {time-full}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {# {time-long}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {# {time-medium}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {# {time-short}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {# {currency}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {# {integer}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {# {number}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {# {ordinal}}`)
}

func TestTokenizeErrMsg(t *testing.T) {
	t.Parallel()

	parser := tik.NewParser(tik.DefaultConfig)

	f := func(t *testing.T, input string, expectErrMsg string) {
		t.Helper()
		tk, err := parser.Parse(input)
		requireEqual(t, expectErrMsg, err.Error())
		requireDeepEqual(t, tik.TIK{}, tk)
	}

	// String literal only.
	f(t, "hello world {", "at index 12: unclosed placeholder")
	f(t, "{unknown}", "at index 0: unknown placeholder")
	f(t, "{# messages in {# folders}}", "at index 15: nested pluralization")
}

func TestTIKPlaceholdersIter(t *testing.T) {
	t.Parallel()

	p := tik.NewParser(tik.DefaultConfig)

	tk, err := p.Parse(`[context]
		{date-full}
		{date-long}
		{date-medium}
		{date-short}
		{time-short}
		{time-medium}
		{time-long}
		{time-full}
		{currency}
		{# messages}{ordinal}`)
	requireNoErr(t, err)
	expect := []tik.Token{
		// 0 is a context.
		tk.Tokens[1],
		// 2 is a string literal.
		tk.Tokens[3],
		// 4 is a string literal.
		tk.Tokens[5],
		// 6 is a string literal.
		tk.Tokens[7],
		// 8 is a string literal.
		tk.Tokens[9],
		// 10 is a string literal.
		tk.Tokens[11],
		// 12 is a string literal.
		tk.Tokens[13],
		// 14 is a string literal.
		tk.Tokens[15],
		// 16 is a string literal.
		tk.Tokens[17],
		// 18 is a string literal.
		tk.Tokens[19],
		// 20 is a string literal.
		// 21 is a cardinal plural end.
		tk.Tokens[22],
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

func TestTokenType_String(t *testing.T) {
	f := func(t *testing.T, expect string, value tik.TokenType) {
		t.Helper()
		requireDeepEqual(t, expect, value.String())
	}

	f(t, `unknown`, 0)
	f(t, `unknown`, 255)
	f(t, `context`, tik.TokenTypeContext)
	f(t, `literal`, tik.TokenTypeLiteral)
	f(t, `text`, tik.TokenTypeText)
	f(t, `text with gender`, tik.TokenTypeTextWithGender)
	f(t, `integer`, tik.TokenTypeInteger)
	f(t, `number`, tik.TokenTypeNumber)
	f(t, `pluralization`, tik.TokenTypeCardinalPluralStart)
	f(t, `pluralization block end`, tik.TokenTypeCardinalPluralEnd)
	f(t, `ordinal plural`, tik.TokenTypeOrdinalPlural)
	f(t, `date full`, tik.TokenTypeDateFull)
	f(t, `date long`, tik.TokenTypeDateLong)
	f(t, `date medium`, tik.TokenTypeDateMedium)
	f(t, `date short`, tik.TokenTypeDateShort)
	f(t, `time full`, tik.TokenTypeTimeFull)
	f(t, `time long`, tik.TokenTypeTimeLong)
	f(t, `time medium`, tik.TokenTypeTimeMedium)
	f(t, `time short`, tik.TokenTypeTimeShort)
	f(t, `currency`, tik.TokenTypeCurrency)
}

func TestICUTranslator(t *testing.T) {
	t.Parallel()

	translator := tik.NewICUTranslator(tik.DefaultConfig)
	p := tik.NewParser(tik.DefaultConfig)

	f := func(t *testing.T, expect, tikInput string) {
		t.Helper()
		tk, err := p.Parse(tikInput)
		requireNoErr(t, err)
		actual := translator.TIK2ICU(tk)
		requireEqual(t, expect, actual)
	}

	f(t, "hello world", "hello world")
	f(t, "hello world", "[context] hello world")
	f(t, "hello {var0}", `hello {text}`)
	f(t, "hello {var0}", `[more context] hello {text}`)
	f(t,
		"today''s lucky number is {var0, number, integer}",
		`today's lucky number is {integer}`)
	f(t,
		"it''s {var0, number} degrees",
		`it's {number} degrees`)
	f(t,
		"your account balance: {var0, number, ::currency/auto}",
		`your account balance: {currency}`)
	f(t,
		`today is {var0, date, full}`,
		"today is {date-full}")
	f(t,
		`today is {var0, date, long}`,
		"today is {date-long}")
	f(t,
		`today is {var0, date, medium}`,
		"today is {date-medium}")
	f(t,
		`today is {var0, date, short}`,
		"today is {date-short}")
	f(t,
		`current time is {var0, time, full}`,
		"current time is {time-full}")
	f(t,
		`current time is {var0, time, long}`,
		"current time is {time-long}")
	f(t,
		`current time is {var0, time, medium}`,
		"current time is {time-medium}")
	f(t,
		`current time is {var0, time, short}`,
		"current time is {time-short}")
	f(t,
		"You''re {var0, selectordinal, other {#th}}",
		`You're {ordinal}`)
	f(t,
		"hello {var0} and {var1}",
		`hello {text} and {text}`)
	f(t,
		"{var0} received the message",
		`{name} received the message`)
	f(t,
		"it''s {var0, date, long}, {var1, time, short}",
		`it's {date-long}, {time-short}`)
	f(t,
		"You have {var0, plural, other {# messages}}",
		`You have {# messages}`)
	f(t,
		"You have {var0, plural, other {# messages}} "+
			"in {var1, plural, other {# folders}}.",
		`You have {# messages} in {# folders}.`)
}

func FuzzTokenize(f *testing.F) {
	f.Add("")
	f.Add(`hello world`)
	f.Add(`integer: {integer}`)
	f.Add(`\n`)
	f.Add(`\{not a placeholder}\{again, not a placeholder}`)
	f.Add(`\\text after`)
	f.Add(`\\\\text after`)
	f.Add("You're {4th} out of {# contenders}")
	f.Add("{unknown}")
	f.Add(`
		{text}
		{name}
		{integer}
		{ordinal}
		{# something}
		{number}
		{currency}
		{date-full}
		{date-long}
		{date-medium}
		{date-short}
		{time-full}
		{time-long}
		{time-medium}
		{time-short}
	`)

	f.Fuzz(func(t *testing.T, input string) {
		parser := tik.NewParser(tik.DefaultConfig)
		tk, err := parser.Parse(input)
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
	parser := tik.NewParser(tik.DefaultConfig)
	for b.Loop() {
		err := parser.ParseFn(`{date-full}{date-long}{date-medium}{date-short}`+
			`{time-short}{time-medium}{time-long}{time-full}`+
			`{integer}{number}{text}{name}{currency}{ordinal}`,
			func(_ tik.TIK) {})
		if err.Err != nil {
			panic(err)
		}
	}
}

func BenchmarkParseFnFewPlaceholders(b *testing.B) {
	parser := tik.NewParser(tik.DefaultConfig)

	loremIpsum, err := os.ReadFile("testdata/lorem_ipsum_fewplaceholders.txt")
	requireNoErr(b, err)
	input := string(loremIpsum)

	for b.Loop() {
		err := parser.ParseFn(input, func(_ tik.TIK) {})
		if err.Err != nil {
			panic(err)
		}
	}
}

func BenchmarkParseFnNoPlaceholders(b *testing.B) {
	parser := tik.NewParser(tik.DefaultConfig)

	loremIpsum, err := os.ReadFile("testdata/lorem_ipsum.txt")
	requireNoErr(b, err)
	input := string(loremIpsum)

	for b.Loop() {
		err := parser.ParseFn(input, func(_ tik.TIK) {})
		if err.Err != nil {
			panic(err)
		}
	}
}

func BenchmarkParseFnNoPlaceholdersShort(b *testing.B) {
	parser := tik.NewParser(tik.DefaultConfig)

	input := string("Short key")

	for b.Loop() {
		err := parser.ParseFn(input, func(_ tik.TIK) {})
		if err.Err != nil {
			panic(err)
		}
	}
}

func BenchmarkTIK2ICUBuf(b *testing.B) {
	parser := tik.NewParser(tik.DefaultConfig)
	translator := tik.NewICUTranslator(tik.DefaultConfig)

	input := string("On {July 16, 1999} you had " +
		"{# messages at {10:30:45 pm PDT}} in {# main folders}")
	tk, err := parser.Parse(input)
	requireNoErr(b, err)

	for b.Loop() {
		translator.TIK2ICUBuf(tk, func(buf *bytes.Buffer) {
			_ = buf // Simulate doing something with the buffer.
		})
	}
}

func requireDeepEqual[T any](tb testing.TB, expect, actual T) {
	tb.Helper()
	if !reflect.DeepEqual(expect, actual) {
		tb.Fatalf("\nexpected: %#v;\nreceived: %#v", expect, actual)
	}
}

func requireEqual[T comparable](tb testing.TB, expect, actual T) {
	tb.Helper()
	if expect != actual {
		tb.Fatalf("\nexpected: %#v;\nreceived: %#v", expect, actual)
	}
}

func requireErrIs(tb testing.TB, expect, actual error) {
	tb.Helper()
	if !errors.Is(actual, expect) {
		var msg string
		if actual != nil {
			msg = actual.Error()
		}
		tb.Fatalf("\nexpected: %#v;\nreceived: %#v (%s)", expect, actual, msg)
	}
}

func requireNoErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("\nexpected: no error;\nreceived: %#v", err)
	}
}
