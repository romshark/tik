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
	TokenTypeStringLiteral

	// String.
	TokenTypeStringPlaceholder // {"..."}

	// Numbers.
	TokenTypeNumber // {3}

	// Pluralization.
	TokenTypeCardinalPluralStart // `{2 `
	TokenTypeCardinalPluralEnd   // `}`
	TokenTypeOrdinalPlural       // {4th}

	// Gender agreement.
	TokenTypeGenderPronoun // {they}, {them}, {their}, {theirs}, {themself}

	// Time.
	TokenTypeTimeShort            // {3:45PM}
	TokenTypeTimeShortSeconds     // {3:45:30PM}
	TokenTypeTimeFullMonthAndDay  // {April 2}
	TokenTypeTimeShortMonthAndDay // {Apr 2}
	TokenTypeTimeFullMonthAndYear // {Apr 2025}
	TokenTypeTimeWeekday          // {Monday}
	TokenTypeTimeDateAndShort     // {April 2, 3:45PM}
	TokenTypeTimeYear             // {2025}
	TokenTypeTimeFull             // {April 2, 3:45:30PM}

	// Currency.
	TokenTypeCurrencyRounded     // {$1}
	TokenTypeCurrencyFull        // {$1.20}
	TokenTypeCurrencyCodeRounded // {USD 1}
	TokenTypeCurrencyCodeFull    // {USD 1.20}
)

func (t TokenType) String() string {
	switch t {
	case TokenTypeContext:
		return `context`
	case TokenTypeStringLiteral:
		return `literal`
	case TokenTypeStringPlaceholder:
		return `string placeholder`
	case TokenTypeNumber:
		return `number`
	case TokenTypeCardinalPluralStart:
		return `pluralization`
	case TokenTypeCardinalPluralEnd:
		return `pluralization block end`
	case TokenTypeOrdinalPlural:
		return `ordinal plural`
	case TokenTypeGenderPronoun:
		return `gender pronoun`
	case TokenTypeTimeShort:
		return `time short`
	case TokenTypeTimeShortSeconds:
		return `time short seconds`
	case TokenTypeTimeFullMonthAndDay:
		return `time full month and day`
	case TokenTypeTimeShortMonthAndDay:
		return `time short month and day`
	case TokenTypeTimeFullMonthAndYear:
		return `time full month and year`
	case TokenTypeTimeWeekday:
		return `time weekday`
	case TokenTypeTimeDateAndShort:
		return `time date and short`
	case TokenTypeTimeYear:
		return `time year`
	case TokenTypeTimeFull:
		return `time full`
	case TokenTypeCurrencyRounded:
		return `currency rounded`
	case TokenTypeCurrencyFull:
		return `currency full`
	case TokenTypeCurrencyCodeRounded:
		return `currency code rounded`
	case TokenTypeCurrencyCodeFull:
		return `currency code full`
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
	ErrConfMagicConstantNonUnique    = errors.New("non-unique magic constant")
	ErrConfMagicConstantInvalid      = errors.New("invalid magic constant")
	ErrConfMissingDefault            = errors.New("missing default value")
	ErrStringPlaceholderEmpty        = errors.New("empty string placeholder text body")
	ErrCardinalPluralEmpty           = errors.New("empty cardinal pluralization")
	ErrDirectiveStartsCardinalPlural = errors.New(
		"directive starts a cardinal pluralization",
	)
	ErrStringPlaceholderInvSpace = errors.New(
		"string placeholder starts or ends with a whitespace character",
	)
	ErrStringPlaceholderIllegalChars = errors.New(
		"string placeholder contains illegal characters",
	)
	ErrUnclosedPlaceholder = errors.New("unclosed placeholder")
	ErrNestedPluralization = errors.New("nested pluralization")
	ErrContextUnclosed     = errors.New("unclosed context")
	ErrContextEmpty        = errors.New("empty context")
	ErrContextInvalid      = errors.New("invalid context")
)

type Tokenizer struct{}

// Tokenize appends all tokens from input to buffer and returns the buffer.
// If c == nil the default configuration applies.
func (t *Tokenizer) Tokenize(buffer Tokens, s string, c *Config) (Tokens, ErrParser) {
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
				Type:       TokenTypeStringLiteral,
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
						Type:       TokenTypeStringLiteral,
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
						Type:       TokenTypeStringLiteral,
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
			// Count the preceeding reverse-solidus.
			if isEscaped(s, iDir-1) {
				// Escaped directive opener, continue reading string literal.
				offset = iDir + 1
				continue
			}

			if literalOffset != iDir {
				buffer = append(buffer, Token{
					IndexStart: literalOffset,
					IndexEnd:   iDir,
					Type:       TokenTypeStringLiteral,
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
		tp, value := match(directive, c)
		switch tp {
		case TokenTypeStringPlaceholder:
			err := validateStringPlaceholder(value)
			if err.Err != nil {
				// +2 for the two '"'.
				err.Index += iDir + 2
				return nil, err
			}
		case TokenTypeCardinalPluralStart:
			if inPluralDirective {
				return nil, err(iDir, ErrNestedPluralization)
			}
			inPluralDirective = true
			// +2 for the '{' and the space after.
			buffer = append(buffer, Token{
				IndexStart: iDir,
				IndexEnd:   iDir + len(value) + 2,
				Type:       TokenTypeCardinalPluralStart,
			})
			offset = iDir + len(value) + 2 // Skip only the plural block start.
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

func match(s string, c *Config) (tokenType TokenType, value string) {
	if s != "" && s[0] == '"' {
		return TokenTypeStringPlaceholder, s
	}
	if strings.EqualFold(s, c.MagicConstants.Number) {
		return TokenTypeNumber, c.MagicConstants.Number
	}
	if strings.EqualFold(s, c.MagicConstants.OrdinalPlural.Constant) {
		return TokenTypeOrdinalPlural, c.MagicConstants.OrdinalPlural.Constant
	}
	if strings.EqualFold(s, c.MagicConstants.TimeShort) {
		return TokenTypeTimeShort, c.MagicConstants.TimeShort
	}
	if strings.EqualFold(s, c.MagicConstants.TimeShortSeconds) {
		return TokenTypeTimeShortSeconds, c.MagicConstants.TimeShortSeconds
	}
	if strings.EqualFold(s, c.MagicConstants.TimeFullMonthAndDay) {
		return TokenTypeTimeFullMonthAndDay, c.MagicConstants.TimeFullMonthAndDay
	}
	if strings.EqualFold(s, c.MagicConstants.TimeShortMonthAndDay) {
		return TokenTypeTimeShortMonthAndDay, c.MagicConstants.TimeShortMonthAndDay
	}
	if strings.EqualFold(s, c.MagicConstants.TimeFullMonthAndYear) {
		return TokenTypeTimeFullMonthAndYear, c.MagicConstants.TimeFullMonthAndYear
	}
	if strings.EqualFold(s, c.MagicConstants.TimeWeekday) {
		return TokenTypeTimeWeekday, c.MagicConstants.TimeWeekday
	}
	if strings.EqualFold(s, c.MagicConstants.TimeDateAndShort) {
		return TokenTypeTimeDateAndShort, c.MagicConstants.TimeDateAndShort
	}
	if strings.EqualFold(s, c.MagicConstants.TimeYear) {
		return TokenTypeTimeYear, c.MagicConstants.TimeYear
	}
	if strings.EqualFold(s, c.MagicConstants.TimeFull) {
		return TokenTypeTimeFull, c.MagicConstants.TimeFull
	}
	if strings.EqualFold(s, c.MagicConstants.CurrencyRounded) {
		return TokenTypeCurrencyRounded, c.MagicConstants.CurrencyRounded
	}
	if strings.EqualFold(s, c.MagicConstants.CurrencyFull) {
		return TokenTypeCurrencyFull, c.MagicConstants.CurrencyFull
	}
	if strings.EqualFold(s, c.MagicConstants.CurrencyCodeRounded) {
		return TokenTypeCurrencyCodeRounded, c.MagicConstants.CurrencyCodeRounded
	}
	if strings.EqualFold(s, c.MagicConstants.CurrencyCodeFull) {
		return TokenTypeCurrencyCodeFull, c.MagicConstants.CurrencyCodeFull
	}
	for _, v := range c.MagicConstants.GenderPronouns {
		if strings.EqualFold(s, v) {
			return TokenTypeGenderPronoun, v
		}
	}
	if p := getPrefixEqualFold(s, c.MagicConstants.CardinalPluralStart); p != "" {
		if l, _ := utf8.DecodeRuneInString(s[len(p):]); !unicode.IsSpace(l) {
			// A whitespace must follow the cardinal plural block start.
			return 0, ""
		}
		return TokenTypeCardinalPluralStart, p
	}
	return 0, ""
}

func getPrefixEqualFold(s, prefix string) string {
	var i, j int
	for i < len(s) && j < len(prefix) {
		r1, size1 := utf8.DecodeRuneInString(s[i:])
		r2, size2 := utf8.DecodeRuneInString(prefix[j:])

		if unicode.ToLower(r1) != unicode.ToLower(r2) {
			return ""
		}

		i, j = i+size1, j+size2
	}
	if j == len(prefix) {
		return s[:i]
	}
	return ""
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
			case TokenTypeContext, TokenTypeStringLiteral, TokenTypeCardinalPluralEnd:
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
	conf   *Config
}

// NewParser creates a new TIK parser instance.
func NewParser(conf *Config) *Parser {
	if conf == nil {
		conf = DefaultConfig()
	}
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

func validateStringPlaceholder(s string) ErrParser {
	if s[len(s)-1] != '"' || len(s) < 2 {
		return err(len(s)-1, ErrStringPlaceholderIllegalChars)
	}

	s = s[1 : len(s)-1]
	if s == "" {
		return err(0, ErrStringPlaceholderEmpty)
	}

	runeL, _ := utf8.DecodeRuneInString(s)         // First.
	runeR, rSize := utf8.DecodeLastRuneInString(s) // Last.

	if unicode.IsSpace(runeL) {
		return err(0, ErrStringPlaceholderInvSpace)
	}
	if unicode.IsSpace(runeR) {
		return err(len(s)-rSize, ErrStringPlaceholderInvSpace)
	}

	if i := strings.IndexAny(s, "\\{}\""); i != -1 {
		return err(i, ErrStringPlaceholderIllegalChars)
	}

	return ErrParser{}
}

func err(index int, err error) ErrParser {
	return ErrParser{Index: index, Err: err}
}
