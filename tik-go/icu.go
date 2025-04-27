package tik

import (
	"bytes"
	"strconv"
)

// ICUTranslator is a reusable TIK to ICU message translator.
type ICUTranslator struct {
	b    bytes.Buffer
	conf *Config
}

func NewICUTranslator(conf *Config) *ICUTranslator {
	if conf == nil {
		conf = defaultConfig
	}
	return &ICUTranslator{conf: conf}
}

type ICUModifier struct{ Gender, Plural bool }

func (i *ICUTranslator) writeModifiers(
	pos int, modifiers map[int]ICUModifier,
	gender, plural bool, writeContent func(),
) {
	if modifiers == nil {
		writeContent()
		return
	}
	m, ok := modifiers[pos]
	if !ok {
		writeContent()
		return
	}

	applyGender := gender && m.Gender
	applyPlural := plural && m.Plural

	if !applyGender && !applyPlural {
		writeContent()
		return
	}

	if applyGender {
		i.write("{")
		i.writePositionalPlaceholder(pos, "_gender")
		i.write(", select, ")
		i.write("other {")
		if applyPlural {
			// Apply both gender and pluralization.
			i.write("{")
			i.writePositionalPlaceholder(pos, "_plural")
			i.write(", plural, ")
			i.write("other {")
			writeContent()
			i.write("}}")
		} else {
			writeContent()
			i.write("}}")
		}
	} else if applyPlural {
		// Apply only pluralization.
		i.write("{")
		i.writePositionalPlaceholder(pos, "_plural")
		i.write(", plural, ")
		i.write("other {")
		writeContent()
		i.write("}}")
	}
}

func (i *ICUTranslator) writePositionalPlaceholder(index int, suffix string) {
	i.b.WriteString("var")
	i.b.WriteString(strconv.Itoa(index))
	i.b.WriteString(suffix)
}

func (i *ICUTranslator) write(s string) { _, _ = i.b.WriteString(s) }

// TIK2ICUBuf similar TIK2ICU but gives temporary access to the internal buffer
// to avoid string allocation if only a temporary byte slice is needed.
// This function can be used instead TIK2ICU to achieve efficiency when possible
// but must be used with caution!
//
// WARNING: Never use or alias buf outside fn!
func (i *ICUTranslator) TIK2ICUBuf(
	tik TIK, modifiers map[int]ICUModifier, fn func(buf *bytes.Buffer),
) {
	i.b.Reset()

	positionalIndex := 0

	for _, token := range tik.Tokens {
		switch token.Type {
		case TokenTypeStringLiteral:
			i.write(token.String(tik.Raw))
		case TokenTypeStringPlaceholder,
			TokenTypeNumber,
			TokenTypeTimeShort,
			TokenTypeTimeShortSeconds,
			TokenTypeTimeFullMonthAndDay,
			TokenTypeTimeShortMonthAndDay,
			TokenTypeTimeFullMonthAndYear,
			TokenTypeTimeWeekday,
			TokenTypeTimeDateAndShort,
			TokenTypeTimeYear,
			TokenTypeTimeFull,
			TokenTypeCurrencyRounded,
			TokenTypeCurrencyFull,
			TokenTypeCurrencyCodeRounded,
			TokenTypeCurrencyCodeFull:

			pos := positionalIndex
			positionalIndex++

			i.writeModifiers(pos, modifiers, true, true, func() {
				i.write("{") // Start placeholder.
				i.writePositionalPlaceholder(pos, "")
				i.write("}")
			})

		case TokenTypeOrdinalPlural:
			pos := positionalIndex
			positionalIndex++

			i.write("{") // Start plural block.
			i.writePositionalPlaceholder(pos, "")
			i.write(", selectordinal, ")
			i.write("other {#")
			i.write(i.conf.MagicConstants.OrdinalPlural.DefaultICUSuffix)
			i.write("}}")

		case TokenTypeCardinalPluralStart:
			pos := positionalIndex
			positionalIndex++

			i.write("{") // Start plural block.
			i.writePositionalPlaceholder(pos, "")
			i.write(", plural, ")
			i.write("other {")
			i.write("# ") // Number placeholder.

		case TokenTypeCardinalPluralEnd:
			i.write("}}") // Finish both other and plural blocks.

		case TokenTypeGenderPronoun:
			pos := positionalIndex
			positionalIndex++

			i.writeModifiers(pos, modifiers, false, true, func() {
				i.write("{")
				i.writePositionalPlaceholder(pos, "")
				i.write(", select, ")
				i.write("other {")
				pronoun := token.String(tik.Raw)
				i.write(pronoun[1 : len(pronoun)-1])
				i.write("}}")
			})
		}
	}

	fn(&i.b)
}

// TIK2ICU translates a TIK into an incomplete ICU message
// that needs to be translated later.
// (See https://unicode-org.github.io/icu/userguide/format_parse/messages/)
// modifiers define positional modifiers such as gender and pluralization
// that weren't defined in the tik.
func (i *ICUTranslator) TIK2ICU(tik TIK, modifiers map[int]ICUModifier) (str string) {
	i.TIK2ICUBuf(tik, modifiers, func(buf *bytes.Buffer) { str = buf.String() })
	return str
}
