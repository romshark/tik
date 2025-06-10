package tik

import (
	"bytes"
	"strconv"
	"strings"
)

// ICUTranslator is a reusable TIK to ICU message translator.
type ICUTranslator struct {
	b    bytes.Buffer
	conf Config
}

func NewICUTranslator(conf Config) *ICUTranslator {
	return &ICUTranslator{conf: conf}
}

func (i *ICUTranslator) writePositionalPlaceholder(index int, suffix string) {
	i.b.WriteString("var")
	i.b.WriteString(strconv.Itoa(index))
	i.b.WriteString(suffix)
}

func (i *ICUTranslator) write(s string) { _, _ = i.b.WriteString(s) }

var replacerEscapeQuote = strings.NewReplacer("'", "''")

// TIK2ICUBuf similar TIK2ICU but gives temporary access to the internal buffer
// to avoid string allocation if only a temporary byte slice is needed.
// This function can be used instead TIK2ICU to achieve efficiency when possible
// but must be used with caution!
//
// WARNING: Never use or alias buf outside fn!
func (i *ICUTranslator) TIK2ICUBuf(
	tik TIK, fn func(buf *bytes.Buffer),
) {
	i.b.Reset()

	positionalIndex := 0

	for _, token := range tik.Tokens {
		switch token.Type {
		case TokenTypeLiteral:
			s := token.String(tik.Raw)
			s = replacerEscapeQuote.Replace(s)
			i.write(s)
		case TokenTypeText, TokenTypeTextWithGender:
			pos := positionalIndex
			positionalIndex++
			i.write("{")
			i.writePositionalPlaceholder(pos, "")
			i.write("}")

		case TokenTypeInteger:
			pos := positionalIndex
			positionalIndex++
			i.write("{")
			i.writePositionalPlaceholder(pos, "")
			i.write(", number, integer}")

		case TokenTypeNumber:
			pos := positionalIndex
			positionalIndex++
			i.write("{")
			i.writePositionalPlaceholder(pos, "")
			i.write(", number}")

		case TokenTypeCurrency:
			pos := positionalIndex
			positionalIndex++
			i.write("{")
			i.writePositionalPlaceholder(pos, "")
			i.write(", number, ::currency/auto}")

		case TokenTypeTimeFull,
			TokenTypeTimeLong,
			TokenTypeTimeMedium,
			TokenTypeTimeShort,
			TokenTypeDateFull,
			TokenTypeDateLong,
			TokenTypeDateMedium,
			TokenTypeDateShort:
			pos := positionalIndex
			positionalIndex++
			var varType, style string
			switch token.Type {
			case TokenTypeTimeFull:
				varType, style = "time", "full"
			case TokenTypeTimeLong:
				varType, style = "time", "long"
			case TokenTypeTimeMedium:
				varType, style = "time", "medium"
			case TokenTypeTimeShort:
				varType, style = "time", "short"
			case TokenTypeDateFull:
				varType, style = "date", "full"
			case TokenTypeDateLong:
				varType, style = "date", "long"
			case TokenTypeDateMedium:
				varType, style = "date", "medium"
			case TokenTypeDateShort:
				varType, style = "date", "short"
			default:
				panic("unexpected token type")
			}

			i.write("{") // Start placeholder.
			i.writePositionalPlaceholder(pos, "")
			i.write(", ")
			i.write(varType)
			i.write(", ")
			i.write(style)
			i.write("}")

		case TokenTypeOrdinalPlural:
			pos := positionalIndex
			positionalIndex++

			i.write("{") // Start plural block.
			i.writePositionalPlaceholder(pos, "")
			i.write(", selectordinal, ")
			i.write("other {#")
			i.write(i.conf.OrdinalPluralOtherSuffix)
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
		}
	}

	fn(&i.b)
}

// TIK2ICU translates a TIK into an incomplete ICU message
// that needs to be translated later.
// (See https://unicode-org.github.io/icu/userguide/format_parse/messages/)
// modifiers define positional modifiers such as gender and pluralization
// that weren't defined in the tik.
func (i *ICUTranslator) TIK2ICU(tik TIK) (str string) {
	i.TIK2ICUBuf(tik, func(buf *bytes.Buffer) { str = buf.String() })
	return str
}
