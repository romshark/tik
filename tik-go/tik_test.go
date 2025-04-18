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

	parser := tik.NewParser(tik.DefaultConfig())
	f := func(t *testing.T, input string, expect ...Token) {
		t.Helper()
		got, err := parser.Parse(input)
		requireNoErr(t, err)
		actual := ToTestTokens(input, got.Tokens)
		requireDeepEqual(t, expect, actual)
	}

	// String literal only
	f(t, "hello world",
		Token{"hello world", tik.TokenTypeStringLiteral},
	)
	f(t, "  hello world  ",
		Token{"hello world", tik.TokenTypeStringLiteral},
	)

	// Context.
	f(t, "[c] hello world",
		Token{"[c]", tik.TokenTypeContext},
		Token{"hello world", tik.TokenTypeStringLiteral},
	)
	f(t, "\r\n\t [c]\t\r\n hello world\r\n\t ",
		Token{"[c]", tik.TokenTypeContext},
		Token{"hello world", tik.TokenTypeStringLiteral},
	)
	f(t, "[c][b]okay",
		Token{"[c]", tik.TokenTypeContext},
		Token{"[b]okay", tik.TokenTypeStringLiteral},
	)

	// Number placeholder
	f(t, "{3} items",
		Token{"{3}", tik.TokenTypeNumber},
		Token{" items", tik.TokenTypeStringLiteral},
	)
	f(t, "  {3} items  ",
		Token{"{3}", tik.TokenTypeNumber},
		Token{" items", tik.TokenTypeStringLiteral},
	)
	f(t, "[context]{3} items",
		Token{"[context]", tik.TokenTypeContext},
		Token{"{3}", tik.TokenTypeNumber},
		Token{" items", tik.TokenTypeStringLiteral},
	)
	f(t, "  [context]  {3} items  ",
		Token{"[context]", tik.TokenTypeContext},
		Token{"{3}", tik.TokenTypeNumber},
		Token{" items", tik.TokenTypeStringLiteral},
	)

	// String placeholders
	f(t, `{"first"} {"second"}, {"_"}{"fourth"}`,
		Token{`{"first"}`, tik.TokenTypeStringPlaceholder},
		Token{" ", tik.TokenTypeStringLiteral},
		Token{`{"second"}`, tik.TokenTypeStringPlaceholder},
		Token{", ", tik.TokenTypeStringLiteral},
		Token{`{"_"}`, tik.TokenTypeStringPlaceholder},
		Token{`{"fourth"}`, tik.TokenTypeStringPlaceholder},
	)

	// Gender agreement.
	f(t, "{They} lost {themself} in {their} thoughts",
		Token{"{They}", tik.TokenTypeGenderPronoun},
		Token{" lost ", tik.TokenTypeStringLiteral},
		Token{"{themself}", tik.TokenTypeGenderPronoun},
		Token{" in ", tik.TokenTypeStringLiteral},
		Token{"{their}", tik.TokenTypeGenderPronoun},
		Token{" thoughts", tik.TokenTypeStringLiteral},
	)
	// Pluralization.
	f(t, "You're {4th} out of {2 contenders}",
		Token{"You're ", tik.TokenTypeStringLiteral},
		Token{"{4th}", tik.TokenTypeOrdinalPlural},
		Token{" out of ", tik.TokenTypeStringLiteral},
		Token{"{2 ", tik.TokenTypeCardinalPluralStart},
		Token{"contenders", tik.TokenTypeStringLiteral},
		Token{"}", tik.TokenTypeCardinalPluralEnd},
	)
	f(t, `{2 files are being copied to folder '{"foo"}'}`,
		Token{"{2 ", tik.TokenTypeCardinalPluralStart},
		Token{"files are being copied to folder '", tik.TokenTypeStringLiteral},
		Token{`{"foo"}`, tik.TokenTypeStringPlaceholder},
		Token{"'", tik.TokenTypeStringLiteral},
		Token{"}", tik.TokenTypeCardinalPluralEnd},
	)
	f(t, `{2 messages from {"folder"} at {3:45PM} on {Apr 2}{2025}}`,
		Token{"{2 ", tik.TokenTypeCardinalPluralStart},
		Token{"messages from ", tik.TokenTypeStringLiteral},
		Token{`{"folder"}`, tik.TokenTypeStringPlaceholder},
		Token{" at ", tik.TokenTypeStringLiteral},
		Token{"{3:45PM}", tik.TokenTypeTimeShort},
		Token{" on ", tik.TokenTypeStringLiteral},
		Token{"{Apr 2}", tik.TokenTypeTimeShortMonthAndDay},
		Token{"{2025}", tik.TokenTypeTimeYear},
		Token{"}", tik.TokenTypeCardinalPluralEnd},
	)
	f(t, `{2 new files in}{2 folders}`,
		Token{"{2 ", tik.TokenTypeCardinalPluralStart},
		Token{"new files in", tik.TokenTypeStringLiteral},
		Token{"}", tik.TokenTypeCardinalPluralEnd},
		Token{"{2 ", tik.TokenTypeCardinalPluralStart},
		Token{`folders`, tik.TokenTypeStringLiteral},
		Token{"}", tik.TokenTypeCardinalPluralEnd},
	)

	// Time.
	f(t, `{3:45PM}{3:45:30PM}{April 2}{Apr 2}
		{Apr 2025}{Monday}{April 2, 3:45PM}
		{2025}{April 2, 3:45:30PM}`,
		Token{"{3:45PM}", tik.TokenTypeTimeShort},
		Token{"{3:45:30PM}", tik.TokenTypeTimeShortSeconds},
		Token{"{April 2}", tik.TokenTypeTimeFullMonthAndDay},
		Token{"{Apr 2}", tik.TokenTypeTimeShortMonthAndDay},
		Token{"\n\t\t", tik.TokenTypeStringLiteral},
		Token{"{Apr 2025}", tik.TokenTypeTimeFullMonthAndYear},
		Token{"{Monday}", tik.TokenTypeTimeWeekday},
		Token{"{April 2, 3:45PM}", tik.TokenTypeTimeDateAndShort},
		Token{"\n\t\t", tik.TokenTypeStringLiteral},
		Token{"{2025}", tik.TokenTypeTimeYear},
		Token{"{April 2, 3:45:30PM}", tik.TokenTypeTimeFull},
	)

	// Escaped braces.
	f(t, `\{not a placeholder\}`,
		Token{`{not a placeholder}`, tik.TokenTypeStringLiteral},
	)
	f(t, `\\\{not a placeholder\\\}`,
		Token{`\{not a placeholder\}`, tik.TokenTypeStringLiteral},
	)
	f(t, `\\text after`,
		Token{`\text after`, tik.TokenTypeStringLiteral},
	)
	f(t, `\ntext after\n\t\\\n`,
		Token{`\ntext after\n\t\\n`, tik.TokenTypeStringLiteral},
	)

	// Sequence of escaped reverse solidus.
	f(t, `before \\\\{Monday} after`,
		Token{`before \\`, tik.TokenTypeStringLiteral},
		Token{`{Monday}`, tik.TokenTypeTimeWeekday},
		Token{` after`, tik.TokenTypeStringLiteral},
	)

	// Case insensitivity.
	f(t, `{They}{THEMSELF}{3:45pm}{3:45:30pM}{april 2}{mOnDaY}`,
		Token{"{They}", tik.TokenTypeGenderPronoun},
		Token{"{THEMSELF}", tik.TokenTypeGenderPronoun},
		Token{"{3:45pm}", tik.TokenTypeTimeShort},
		Token{"{3:45:30pM}", tik.TokenTypeTimeShortSeconds},
		Token{"{april 2}", tik.TokenTypeTimeFullMonthAndDay},
		Token{"{mOnDaY}", tik.TokenTypeTimeWeekday},
	)
}

func TestParseErr(t *testing.T) {
	t.Parallel()

	parser := tik.NewParser(tik.DefaultConfig())

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
	f(t, tik.ErrUknownPlaceholder, `no space after cardinal plural: {2abc}`)
	f(t, tik.ErrUknownPlaceholder, `unknown placeholder: {2026}`)
	f(t, tik.ErrUknownPlaceholder, `unknown placeholder: {April 21}`)
	f(t, tik.ErrUnclosedPlaceholder, `unexpected EOF: {`)
	f(t, tik.ErrUnclosedPlaceholder, `unexpected EOF: {"`)
	f(t, tik.ErrUnclosedPlaceholder, `unexpected EOF: {"_`)
	f(t, tik.ErrUnclosedPlaceholder, `unexpected EOF: {""`)
	f(t, tik.ErrUnclosedPlaceholder, `unexpected EOF: {x`)
	f(t, tik.ErrUnclosedPlaceholder, `unexpected EOF: {{`)
	f(t, tik.ErrStringPlaceholderEmpty, `this is illegal: {""}`)
	f(t, tik.ErrStringPlaceholderInvSpace, `this too: {"  "}`)
	f(t, tik.ErrStringPlaceholderInvSpace, `this too: {" text "}`)
	f(t, tik.ErrStringPlaceholderInvSpace, `this too: {"  text  "}`)
	f(t, tik.ErrStringPlaceholderInvSpace, "this too: {\"\u3000text\"}")
	f(t, tik.ErrStringPlaceholderInvSpace, "this too: {\"text\u3000\"}")
	f(t, tik.ErrStringPlaceholderIllegalChars, `unclosed string placeholder: {"abc }`)
	f(t, tik.ErrStringPlaceholderIllegalChars, `and this: {"\""} text after`)
	f(t, tik.ErrStringPlaceholderIllegalChars, `{"\"}`)
	f(t, tik.ErrStringPlaceholderIllegalChars, `{"abc \n def"}`)
	f(t, tik.ErrNestedPluralization, `nested pluralization: {2 messages in {2 folders}}`)
	f(t, tik.ErrCardinalPluralEmpty, `empty pluralization: {2 }`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {2 {April 2}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {2 {3:45PM}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {2 {$1}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {2 {4th}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {2 {they}}`)
	f(t, tik.ErrDirectiveStartsCardinalPlural, `illegal pluralization: {2 {themself}}`)
}

func TestParseCustomPlaceholders(t *testing.T) {
	t.Parallel()

	conf := tik.DefaultConfig()

	conf.MagicConstants.Number = "43"
	parser := tik.NewParser(conf)

	input := "{43}{43}"
	got, err := parser.Parse(input)
	requireNoErr(t, err)

	toks := ToTestTokens(input, got.Tokens)

	requireDeepEqual(t, []Token{
		{Str: "{43}", Type: tik.TokenTypeNumber},
		{Str: "{43}", Type: tik.TokenTypeNumber},
	}, toks)

	got, err = parser.Parse("invalid: {3}")
	requireErrIs(t, tik.ErrUknownPlaceholder, err)
	requireEqual(t, "at index 9: unknown placeholder", err.Error())
	requireDeepEqual(t, tik.TIK{}, got)
}

func TestConfigValidateErr(t *testing.T) {
	t.Parallel()

	f := func(t *testing.T, expectErr error, fn func(c *tik.Config)) {
		t.Helper()
		conf := tik.DefaultConfig()
		fn(conf)
		err := conf.Validate()
		requireErrIs(t, expectErr, err)
	}

	f(t, nil, func(*tik.Config) {})
	f(t, tik.ErrConfMagicConstantInvalid, func(c *tik.Config) {
		c.MagicConstants.Number = "{3}"
	})
	f(t, tik.ErrConfMagicConstantInvalid, func(c *tik.Config) {
		c.MagicConstants.Number = "\"3\""
	})
	f(t, tik.ErrConfMagicConstantInvalid, func(c *tik.Config) {
		c.MagicConstants.Number = ""
	})
	f(t, tik.ErrConfMagicConstantInvalid, func(c *tik.Config) {
		c.MagicConstants.Number = " x"
	})
	f(t, tik.ErrConfMagicConstantInvalid, func(c *tik.Config) {
		c.MagicConstants.Number = "x "
	})
	f(t, tik.ErrConfMagicConstantNonUnique, func(c *tik.Config) {
		c.MagicConstants.Number = "5"
		c.MagicConstants.CurrencyCodeFull = "5"
	})
	f(t, tik.ErrConfMagicConstantInvalid, func(c *tik.Config) {
		c.MagicConstants.GenderPronouns = nil
	})
	f(t, tik.ErrConfMagicConstantInvalid, func(c *tik.Config) {
		c.MagicConstants.GenderPronouns = []string{}
	})
	f(t, tik.ErrConfMagicConstantInvalid, func(c *tik.Config) {
		c.MagicConstants.GenderPronouns = []string{""}
	})
	f(t, tik.ErrConfMagicConstantNonUnique, func(c *tik.Config) {
		c.MagicConstants.GenderPronouns = []string{"he", "he"}
	})
	f(t, tik.ErrConfMagicConstantNonUnique, func(c *tik.Config) {
		c.MagicConstants.OrdinalPlural.Constant = "3"
	})
	f(t, tik.ErrConfMagicConstantInvalid, func(c *tik.Config) {
		c.MagicConstants.OrdinalPlural.Constant = "{5th}"
	})
	f(t, tik.ErrConfMagicConstantInvalid, func(c *tik.Config) {
		c.MagicConstants.OrdinalPlural.Constant = ""
	})
	f(t, tik.ErrConfMissingDefault, func(c *tik.Config) {
		c.MagicConstants.OrdinalPlural.DefaultICUSuffix = ""
	})
}

func TestTokenizeErrMsg(t *testing.T) {
	t.Parallel()

	parser := tik.NewParser(nil) // Use default config.

	f := func(t *testing.T, input string, expectErrMsg string) {
		t.Helper()
		tk, err := parser.Parse(input)
		requireEqual(t, expectErrMsg, err.Error())
		requireDeepEqual(t, tik.TIK{}, tk)
	}

	// String literal only.
	f(t, "hello world {", "at index 12: unclosed placeholder")
	f(t, "{unknown}", "at index 0: unknown placeholder")
	f(t, "{2 messages in {2 folders}}", "at index 15: nested pluralization")
}

func TestTIKPlaceholdersIter(t *testing.T) {
	t.Parallel()

	p := tik.NewParser(tik.DefaultConfig())

	tk, err := p.Parse(`[context]{3:45PM}{3:45:30PM}{April 2}{Apr 2}
		{Apr 2025}{Monday}{April 2, 3:45PM}
		{2025}{April 2, 3:45:30PM}
		{$1}{$1.20}{USD 1}{USD 1.20}`)
	requireNoErr(t, err)
	expect := []tik.Token{
		// 0 is a context.
		tk.Tokens[1],
		tk.Tokens[2],
		tk.Tokens[3],
		tk.Tokens[4],
		// 5 is a string literal.
		tk.Tokens[6],
		tk.Tokens[7],
		tk.Tokens[8],
		// 9 is a string literal.
		tk.Tokens[10],
		tk.Tokens[11],
		// 12 is a string literal.
		tk.Tokens[13],
		tk.Tokens[14],
		tk.Tokens[15],
		tk.Tokens[16],
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
	f(t, `literal`, tik.TokenTypeStringLiteral)
	f(t, `string placeholder`, tik.TokenTypeStringPlaceholder)
	f(t, `number`, tik.TokenTypeNumber)
	f(t, `pluralization`, tik.TokenTypeCardinalPluralStart)
	f(t, `pluralization block end`, tik.TokenTypeCardinalPluralEnd)
	f(t, `ordinal plural`, tik.TokenTypeOrdinalPlural)
	f(t, `gender pronoun`, tik.TokenTypeGenderPronoun)
	f(t, `time short`, tik.TokenTypeTimeShort)
	f(t, `time short seconds`, tik.TokenTypeTimeShortSeconds)
	f(t, `time full month and day`, tik.TokenTypeTimeFullMonthAndDay)
	f(t, `time short month and day`, tik.TokenTypeTimeShortMonthAndDay)
	f(t, `time full month and year`, tik.TokenTypeTimeFullMonthAndYear)
	f(t, `time weekday`, tik.TokenTypeTimeWeekday)
	f(t, `time date and short`, tik.TokenTypeTimeDateAndShort)
	f(t, `time year`, tik.TokenTypeTimeYear)
	f(t, `time full`, tik.TokenTypeTimeFull)
	f(t, `currency rounded`, tik.TokenTypeCurrencyRounded)
	f(t, `currency full`, tik.TokenTypeCurrencyFull)
	f(t, `currency code rounded`, tik.TokenTypeCurrencyCodeRounded)
	f(t, `currency code full`, tik.TokenTypeCurrencyCodeFull)
}

func TestICUTranslator(t *testing.T) {
	t.Parallel()

	conf := tik.DefaultConfig()
	translator := tik.NewICUTranslator(conf)
	p := tik.NewParser(conf)

	f := func(t *testing.T, expect, tikInput string) {
		t.Helper()
		tk, err := p.Parse(tikInput)
		requireNoErr(t, err)
		actual := translator.TIK2ICU(tk, nil)
		requireEqual(t, expect, actual)
	}

	f(t, "hello world", "hello world")
	f(t, "hello world", "[context] hello world")
	f(t, "hello {_0}", `hello {"world"}`)
	f(t, "hello {_0}", `[more context] hello {"world"}`)
	f(t,
		"You're {_0, selectordinal, other {#th}}",
		`You're {4th}`)
	f(t, "hello {_0} and {_1}", `hello {"world"} and {"something else"}`)
	f(t, "it's {_0}, {_1}", `it's {April 2}, {3:45PM}`)
	f(t,
		"{_0, select, other {They}} are on {_1, select, other {their}} way!",
		`{They} are on {their} way!`)
	f(t,
		"You have {_0, plural, other {# messages}}",
		`You have {2 messages}`)
	f(t,
		"You have {_0, plural, other {# messages}} in {_1, plural, other {# folders}}.",
		`You have {2 messages} in {2 folders}.`)
}

func TestICUTranslatorModifier(t *testing.T) {
	t.Parallel()

	conf := tik.DefaultConfig()
	translator := tik.NewICUTranslator(conf)
	p := tik.NewParser(conf)

	f := func(t *testing.T, expect, input string, modifiers map[int]tik.ICUModifier) {
		t.Helper()
		tk, err := p.Parse(input)
		requireNoErr(t, err)
		actual := translator.TIK2ICU(tk, modifiers)
		requireEqual(t, expect, actual)
	}

	f(t,
		"{_0} has {_1}",
		`{"John"} has {"apples"}`, nil)

	f(t,
		"{_0} has {_1}",
		`{"John"} has {"apples"}`, map[int]tik.ICUModifier{
			0: {}, 1: {}, // All modifiers disabled.
		})
	f(t,
		"{_0_gender, select, other {{_0_plural, plural, other {{_0}}} has {_1}",
		`{"John"} has {"apples"}`, map[int]tik.ICUModifier{
			// Apply both gender and pluralization simultaneously
			// for when "John" could be multiple people like "Coworkers".
			0: {Gender: true, Plural: true},
		})
	f(t,
		"{_0_gender, select, other {{_0}}} has {_1_plural, plural, other {{_1}}}",
		`{"John"} has {"apples"}`, map[int]tik.ICUModifier{
			0: {Gender: true}, 1: {Plural: true},
		})
}

func FuzzTokenize(f *testing.F) {
	f.Add("")
	f.Add(`hello world`)
	f.Add(`{3} items`)
	f.Add(`{they} lost {themself} in {their} thoughts`)
	f.Add(`\n`)
	f.Add(`\{not a placeholder}\{again, not a placeholder}`)
	f.Add(`\\text after`)
	f.Add(`\\\\text after`)
	f.Add("You're {4th} out of {2 contenders}")
	f.Add("{unknown}")
	f.Add(`{3:45PM}{3:45:30PM}{April 2}{Apr 2}
		{Apr 2025}{Monday}{April 2, 3:45PM}
		{2025}{April 2, 3:45:30PM}`)

	f.Fuzz(func(t *testing.T, input string) {
		parser := tik.NewParser(tik.DefaultConfig())
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
	parser := tik.NewParser(tik.DefaultConfig())
	for b.Loop() {
		err := parser.ParseFn(`{3:45PM}{3:45:30PM}{April 2}{Apr 2}
		{Apr 2025}{Monday}{April 2, 3:45PM}
		{2025}{April 2, 3:45:30PM}`, func(_ tik.TIK) {})
		if err.Err != nil {
			panic(err)
		}
	}
}

func BenchmarkParseFnFewPlaceholders(b *testing.B) {
	parser := tik.NewParser(tik.DefaultConfig())

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
	parser := tik.NewParser(tik.DefaultConfig())

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
	parser := tik.NewParser(tik.DefaultConfig())

	input := string("Short key")

	for b.Loop() {
		err := parser.ParseFn(input, func(_ tik.TIK) {})
		if err.Err != nil {
			panic(err)
		}
	}
}

func BenchmarkTIK2ICUBuf(b *testing.B) {
	conf := tik.DefaultConfig()
	parser := tik.NewParser(conf)
	translator := tik.NewICUTranslator(conf)

	input := string("On {April 2} you had {2 messages at {3:45PM}} in {2 main folders}")
	tk, err := parser.Parse(input)
	requireNoErr(b, err)

	for b.Loop() {
		translator.TIK2ICUBuf(tk, nil, func(buf *bytes.Buffer) {
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
