package tik

import (
	"errors"
	"fmt"
	"iter"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Tokens is a slice of the lexical tokens of a textual internationalization key.
type Tokens []Token

// TokenType defines the type of a TIK lexical token.
type TokenType uint8

const (
	_ TokenType = iota
	TokenTypeContext
	TokenTypeLiteral

	// String.
	TokenTypeText           // {text}
	TokenTypeTextWithGender // {name}

	// Numbers.
	TokenTypeInteger // {integer}
	TokenTypeNumber  // {number}

	// Pluralization.
	TokenTypeCardinalPluralStart // `{# `
	TokenTypeCardinalPluralEnd   // `}`
	TokenTypeOrdinalPlural       // {ordinal}

	// DATE

	// TokenTypeDateFull equals "EEEE, MMMM d, y"
	TokenTypeDateFull // {date-full}

	// TokenTypeDateLong equals "MMMM d, y"
	TokenTypeDateLong // {date-long}

	// TokenTypeDateMedium equals "MMM d, y"
	TokenTypeDateMedium // {date-medium}

	// TokenTypeDateShort equals "M/d/yy"
	TokenTypeDateShort // {date-short}

	// TIME

	// TokenTypeTimeFull equals "hour(h/H), minute(mm), second(ss), and zone(zzzz)."
	TokenTypeTimeFull // {time-full}

	// TokenTypeTimeLong equals "hour, minute, second, and zone(z)"
	TokenTypeTimeLong // {time-long}

	// TokenTypeTimeMedium equals "hour, minute, second."
	TokenTypeTimeMedium // {time-medium}

	// TokenTypeTimeShort equals "hour, minute."
	TokenTypeTimeShort // {time-short}

	// Currency.
	TokenTypeCurrency // {currency}
)

func (t TokenType) String() string {
	switch t {
	case TokenTypeContext:
		return `context`
	case TokenTypeLiteral:
		return `literal`
	case TokenTypeText:
		return `text`
	case TokenTypeTextWithGender:
		return `text with gender`
	case TokenTypeInteger:
		return `integer`
	case TokenTypeNumber:
		return `number`
	case TokenTypeCardinalPluralStart:
		return `pluralization`
	case TokenTypeCardinalPluralEnd:
		return `pluralization block end`
	case TokenTypeOrdinalPlural:
		return `ordinal plural`
	case TokenTypeTimeShort:
		return `time short`
	case TokenTypeTimeMedium:
		return `time medium`
	case TokenTypeTimeLong:
		return `time long`
	case TokenTypeTimeFull:
		return `time full`
	case TokenTypeDateShort:
		return `date short`
	case TokenTypeDateMedium:
		return `date medium`
	case TokenTypeDateLong:
		return `date long`
	case TokenTypeDateFull:
		return `date full`
	case TokenTypeCurrency:
		return `currency`
	}
	return "unknown"
}

// Token is a lexical TIK token.
type Token struct {
	// IndexStart defines the start index of this token in the original TIK.
	IndexStart int
	// IndexEnd defines the end index of this token in the original TIK.
	IndexEnd int
	Type     TokenType
}

var replacerTokenStringify = strings.NewReplacer("\\\\", "\\", "\\{", "{", "\\}", "}")

func (t Token) String(source string) string {
	s := source[t.IndexStart:t.IndexEnd]
	if strings.IndexByte(s, '\\') == -1 {
		// Fast path, no reverse solidus
		return s
	}
	return replacerTokenStringify.Replace(s)
}

var (
	ErrTextEmpty                     = errors.New("empty text body")
	ErrUnexpClosure                  = errors.New("unexpected directive closure")
	ErrUknownPlaceholder             = errors.New("unknown placeholder")
	ErrCardinalPluralEmpty           = errors.New("empty cardinal pluralization")
	ErrDirectiveStartsCardinalPlural = errors.New(
		"directive starts a cardinal pluralization",
	)
	ErrUnclosedPlaceholder = errors.New("unclosed placeholder")
	ErrNestedPluralization = errors.New("nested pluralization")
	ErrContextUnclosed     = errors.New("unclosed context")
	ErrContextEmpty        = errors.New("empty context")
	ErrContextInvalid      = errors.New("invalid context")
)

// Config defines the TIK environment configuration.
type Config struct {
	OrdinalPluralOtherSuffix string
}

var DefaultConfig = Config{
	OrdinalPluralOtherSuffix: "th",
}

type Tokenizer struct{}

// Tokenize appends all tokens from input to buffer and returns the buffer.
// If c == nil the default configuration applies.
func (t *Tokenizer) Tokenize(buffer Tokens, s string, c Config) (Tokens, ErrParser) {
	inPluralDirective := false
	offset := 0

	// Skip prefix spaces.
	for offset < len(s) {
		l, size := utf8.DecodeRuneInString(s[offset:])
		if !unicode.IsSpace(l) {
			break
		}
		offset += size
	}

	if offset >= len(s) {
		return nil, err(0, ErrTextEmpty)
	}
	if s[offset] == '[' {
		start := offset
		offset++
		// TIK has context.
		contextEnd := strings.IndexByte(s[offset:], ']')
		if contextEnd == -1 {
			return buffer, ErrParser{Err: ErrContextUnclosed}
		}

		context := s[offset : offset+contextEnd]
		if strings.TrimSpace(context) == "" {
			return buffer, ErrParser{Err: ErrContextEmpty}
		}
		if strings.ContainsAny(context, "{}[]\\") {
			// Contains either of: { } [ ] \
			return buffer, ErrParser{Err: ErrContextInvalid}
		}
		offset += contextEnd + 1
		buffer = append(buffer, Token{
			IndexStart: start,
			IndexEnd:   offset,
			Type:       TokenTypeContext,
		})

		// Skip spaces before the start of the text.
		for offset < len(s) {
			l, size := utf8.DecodeRuneInString(s[offset:])
			if !unicode.IsSpace(l) {
				break
			}
			offset += size
		}

		if offset >= len(s) {
			return buffer, ErrParser{Index: offset, Err: ErrTextEmpty}
		}
	}

	{
		i := strings.IndexByte(s[offset:], '{')
		j := strings.IndexByte(s[offset:], '}')
		if i == -1 && j == -1 {
			// Fast path for simple inputs without {}.
			indexEnd := len(s)
			// Ignore suffix spaces.
			for indexEnd >= 0 {
				l, size := utf8.DecodeLastRuneInString(s[offset:indexEnd])
				if !unicode.IsSpace(l) {
					break
				}
				indexEnd -= size
			}
			return append(buffer, Token{
				IndexStart: offset,
				IndexEnd:   indexEnd,
				Type:       TokenTypeLiteral,
			}), ErrParser{}
		}
	}

	for {
		var iDir int
		for literalOffset := offset; ; {
			// Read string literal before the next directive.
			iDir = strings.IndexAny(s[offset:], "{}")
			if iDir == -1 {
				// There is no next directive.
				if literalOffset != len(s) {
					// End of string literal.
					indexEnd := len(s)
					// Ignore suffix spaces.
					for indexEnd >= 0 {
						l, size := utf8.DecodeLastRuneInString(s[:indexEnd])
						if !unicode.IsSpace(l) {
							break
						}
						indexEnd -= size
					}
					buffer = append(buffer, Token{
						IndexStart: literalOffset,
						IndexEnd:   indexEnd,
						Type:       TokenTypeLiteral,
					})
				}
				// End of TIK.
				return buffer, ErrParser{}
			}

			iDir += offset
			if s[iDir] == '}' {
				// A dangling } must be escaped if it was meant to just be a literal '}'.
				if !inPluralDirective {
					if isEscaped(s, iDir-1) {
						// Escaped, continue reading literal.
						offset = iDir + 1
						continue
					}
					return nil, err(iDir, ErrUnexpClosure)
				}
				if literalOffset != iDir {
					// End of string literal.
					buffer = append(buffer, Token{
						IndexStart: literalOffset,
						IndexEnd:   iDir,
						Type:       TokenTypeLiteral,
					})
				}
				if t := buffer[len(buffer)-1]; t.Type == TokenTypeCardinalPluralStart {
					// Cardinal plural blocks must contain at least 1 token.
					return nil, err(t.IndexStart, ErrCardinalPluralEmpty)
				}
				buffer = append(buffer, Token{
					IndexStart: iDir,
					IndexEnd:   iDir + 1,
					Type:       TokenTypeCardinalPluralEnd,
				})
				inPluralDirective = false

				// Restart literal parsing cycle.
				offset = iDir + 1
				literalOffset = offset
				continue
			}

			// Directive opener { discovered.
			// Count the preceding reverse-solidus.
			if isEscaped(s, iDir-1) {
				// Escaped directive opener, continue reading string literal.
				offset = iDir + 1
				continue
			}

			if literalOffset != iDir {
				buffer = append(buffer, Token{
					IndexStart: literalOffset,
					IndexEnd:   iDir,
					Type:       TokenTypeLiteral,
				})
			}
			break
		}

		iDirClose := strings.IndexByte(s[iDir+1:], '}')
		if iDirClose == -1 {
			return nil, err(iDir, ErrUnclosedPlaceholder)
		}
		iDirClose += iDir

		directive := s[iDir+1 : iDirClose+1]
		tp, ln := match(directive)
		switch tp {
		case TokenTypeCardinalPluralStart:
			if inPluralDirective {
				return nil, err(iDir, ErrNestedPluralization)
			}
			inPluralDirective = true
			// +2 for the '{' and the space after.
			buffer = append(buffer, Token{
				IndexStart: iDir,
				IndexEnd:   iDir + ln + 2,
				Type:       TokenTypeCardinalPluralStart,
			})
			offset = iDir + ln + 2 // Skip only the plural block start.
			continue
		case 0:
			return nil, err(iDir, ErrUknownPlaceholder)
		}

		if b := buffer; len(b) > 0 && b[len(b)-1].Type == TokenTypeCardinalPluralStart {
			// Cardinal pluralization block must not begin with another directive.
			return nil, err(iDir, ErrDirectiveStartsCardinalPlural)
		}
		buffer = append(buffer, Token{
			IndexStart: iDir,
			IndexEnd:   iDirClose + 2,
			Type:       tp,
		})
		offset = iDirClose + 2
	}
}

func match(s string) (tokenType TokenType, length int) {
	switch s {
	case "text":
		return TokenTypeText, len("text")
	case "name":
		return TokenTypeTextWithGender, len("name")
	case "integer":
		return TokenTypeInteger, len("integer")
	case "number":
		return TokenTypeNumber, len("number")
	case "ordinal":
		return TokenTypeOrdinalPlural, len("ordinal")
	case "time-full":
		return TokenTypeTimeFull, len("time-full")
	case "time-long":
		return TokenTypeTimeLong, len("time-long")
	case "time-medium":
		return TokenTypeTimeMedium, len("time-medium")
	case "time-short":
		return TokenTypeTimeShort, len("time-short")
	case "date-full":
		return TokenTypeDateFull, len("date-full")
	case "date-long":
		return TokenTypeDateLong, len("date-long")
	case "date-medium":
		return TokenTypeDateMedium, len("date-medium")
	case "date-short":
		return TokenTypeDateShort, len("date-short")
	case "currency":
		return TokenTypeCurrency, len("currency")
	}
	if strings.HasPrefix(s, "#") {
		if l, _ := utf8.DecodeRuneInString(s[len("#"):]); !unicode.IsSpace(l) {
			// A whitespace must follow the cardinal plural block start.
			return 0, 0
		}
		return TokenTypeCardinalPluralStart, len("#")
	}
	return 0, 0
}

// isEscaped expects i to point to index -1 relative to the subject byte.
func isEscaped(s string, i int) bool {
	pRevSol := 0
	for ; i >= 0; i, pRevSol = i-1, pRevSol+1 {
		if s[i] != '\\' {
			break
		}
	}
	return pRevSol%2 != 0
}

// TIK is a parsed and validated textual internationalization token.
type TIK struct {
	Raw    string
	Tokens Tokens
}

// Placeholders returns an iterators that iterates over placeholder tokens.
func (t TIK) Placeholders() iter.Seq2[int, Token] {
	return func(yield func(int, Token) bool) {
		i := 0
		for _, t := range t.Tokens {
			switch t.Type {
			case TokenTypeContext, TokenTypeLiteral, TokenTypeCardinalPluralEnd:
				continue
			}
			if !yield(i, t) {
				break
			}
			i++
		}
	}
}

// Parser is a TIK parser instance.
type Parser struct {
	t      Tokenizer
	tokBuf Tokens
	conf   Config
}

// NewParser creates a new TIK parser instance.
func NewParser(conf Config) *Parser {
	return &Parser{
		tokBuf: make(Tokens, 0, 16),
		conf:   conf,
	}
}

type ErrParser struct {
	Index int
	Err   error
}

func (e ErrParser) Error() string {
	return fmt.Sprintf("at index %d: %v", e.Index, e.Err)
}

func (e ErrParser) Unwrap() error { return e.Err }

// ParseFn is similar to Parse but avoid copying the token buffer
// and instead uses the original buffer of the parser in the tik provided to fn.
//
// WARNING: Do not alias and use the token slice once fn returns!
func (p *Parser) ParseFn(input string, fn func(tik TIK)) ErrParser {
	p.tokBuf = p.tokBuf[:0] // Reset buffer.
	var err ErrParser
	p.tokBuf, err = p.t.Tokenize(p.tokBuf, input, p.conf)
	if err.Err != nil {
		return err
	}
	fn(TIK{Raw: input, Tokens: p.tokBuf})
	return ErrParser{}
}

// Parse parses input and returns a validated TIK, otherwise returns an error.
// The tokens slice in the returned TIK is a copy of the buffer and doesn't alias
// the internal parser buffer.
// If you want to avoid the token buffer copy use ParseFn with caution instead.
func (p *Parser) Parse(input string) (tik TIK, err error) {
	errParser := p.ParseFn(input, func(ref TIK) {
		cp := make(Tokens, len(ref.Tokens))
		copy(cp, ref.Tokens)
		tik.Raw, tik.Tokens = input, cp
	})
	if errParser.Err != nil {
		return TIK{}, errParser
	}
	return tik, nil
}

func err(index int, err error) ErrParser {
	return ErrParser{Index: index, Err: err}
}
