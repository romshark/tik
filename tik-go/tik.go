// Package tik provides the parser for the Textual Internationalization Key format.
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
	TokenTypeStringLiteral

	// String.
	TokenTypeStringPlaceholder // {"..."}

	// Numbers.
	TokenTypeNumber // {3}

	// Pluralization.
	TokenTypeCardinalPlural // {2}
	TokenTypeOrdinalPlural  // {4th}

	// Gender agreement.
	TokenTypeGenderAgreement // {he}, {his}, {him}, {himself}

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

// Token is a lexical TIK token.
type Token struct {
	Str string
	// Index defines the start index of this token in the original TIK.
	Index int
	Type  TokenType
}

var defaultMapping = map[string]TokenType{
	"3":   TokenTypeNumber,
	"2":   TokenTypeCardinalPlural,
	"4th": TokenTypeOrdinalPlural,

	"he":      TokenTypeGenderAgreement,
	"his":     TokenTypeGenderAgreement,
	"him":     TokenTypeGenderAgreement,
	"himself": TokenTypeGenderAgreement,

	"3:45PM":             TokenTypeTimeShort,
	"3:45:30PM":          TokenTypeTimeShortSeconds,
	"April 2":            TokenTypeTimeFullMonthAndDay,
	"Apr 2":              TokenTypeTimeShortMonthAndDay,
	"Apr 2025":           TokenTypeTimeFullMonthAndYear,
	"Monday":             TokenTypeTimeWeekday,
	"April 2, 3:45PM":    TokenTypeTimeDateAndShort,
	"2025":               TokenTypeTimeYear,
	"April 2, 3:45:30PM": TokenTypeTimeFull,

	"$1":       TokenTypeCurrencyRounded,
	"$1.20":    TokenTypeCurrencyFull,
	"USD 1":    TokenTypeCurrencyCodeRounded,
	"USD 1.20": TokenTypeCurrencyCodeFull,
}

type Config struct {
	Mapping map[string]TokenType
}

var (
	ErrUknownPlaceholder         = errors.New("unknown placeholder")
	ErrInvalidCustomPlaceholders = errors.New("expected exactly 17 mapped typed")
	ErrInvalidCustomPlaceholder  = errors.New("invalid custom placeholder")
	ErrStringPlaceholderEmpty    = errors.New(
		"string placeholder text body is empty",
	)
	ErrStringPlaceholderInvSpace = errors.New(
		"string placeholder starts or ends with a whitespace character",
	)
	ErrStringPlaceholderIllegalChars = errors.New(
		"string placeholder contains illegal characters",
	)
	ErrUnclosedStringPlaceholder = errors.New("unclosed string placeholder")
	ErrUnclosedPlaceholder       = errors.New("unclosed placeholder")
)

func CustomPlaceholders(
	mutate func(map[string]TokenType),
) (map[string]TokenType, error) {
	cp := make(map[string]TokenType, len(defaultMapping))
	for k, v := range defaultMapping {
		cp[k] = v
	}

	mutate(cp)

	byType := make(map[TokenType]struct{})
	for k, v := range cp {
		if err := ValidateCustomPlaceholder(k); err != nil {
			return nil, err
		}
		byType[v] = struct{}{}
	}
	if len(byType) != 17 {
		return nil, fmt.Errorf("%w, got: %d",
			ErrInvalidCustomPlaceholders, len(byType))
	}
	return cp, nil
}

func ValidateCustomPlaceholder(s string) error {
	if strings.ContainsAny(s, "\"") {
		return ErrInvalidCustomPlaceholder
	}
	return nil
}

type Tokenizer struct {
	builder strings.Builder
}

// Tokenize appends all tokens from input to buffer and returns the buffer.
// If c == nil the default configuration applies.
func (t *Tokenizer) Tokenize(buffer Tokens, input string, c *Config) (Tokens, error) {
	if input == "" {
		return nil, nil
	}
	if strings.IndexByte(input, '{') == -1 {
		// Fast path for when the input doesn't contain any placeholders.
		return append(buffer, Token{Type: TokenTypeStringLiteral, Str: input}), nil
	}

	mapping := defaultMapping
	if c != nil && c.Mapping != nil {
		mapping = c.Mapping
	}
	t.builder.Reset()
	// literalStart tracks the starting byte index of the current literal segment.
	literalStart := 0

	flushLiteral := func(cur int) {
		if t.builder.Len() > 0 {
			buffer = append(buffer, Token{
				Str:   t.builder.String(),
				Index: literalStart,
				Type:  TokenTypeStringLiteral,
			})
			t.builder.Reset()
		}
		literalStart = cur
	}

	i := 0
	for i < len(input) {
		// Jump to next '\' or '{' in one call.
		nextSpecial := strings.IndexAny(input[i:], `{\`)
		if nextSpecial == -1 {
			t.builder.WriteString(input[i:])
			i = len(input)
			break
		}
		specialIndex := i + nextSpecial
		// Batch write all non-special literal characters.
		if specialIndex > i {
			t.builder.WriteString(input[i:specialIndex])
		}

		switch input[specialIndex] {
		case '\\':
			// Count consecutive backslashes.
			j := specialIndex
			for j < len(input) && input[j] == '\\' {
				j++
			}
			count := j - specialIndex
			i = j
			// Check if a '{' follows.
			if i < len(input) && input[i] == '{' {
				if count%2 == 1 {
					// Odd count: escape the '{'.
					t.builder.WriteString(strings.Repeat("\\", count/2))
					t.builder.WriteByte('{')
					i++ // Consume the '{'.
				} else {
					// Even count: backslashes are literal
					// and the '{' starts a placeholder.
					t.builder.WriteString(strings.Repeat("\\", count/2))
					// Let the '{' be handled in the next iteration.
				}
			}
		case '{':
			flushLiteral(specialIndex)
			startPlaceholder := specialIndex

			if startPlaceholder+1 < len(input) && input[startPlaceholder+1] == '"' {
				if startPlaceholder+2 >= len(input) {
					return nil, fmt.Errorf("%w at index %d",
						ErrUnclosedStringPlaceholder, startPlaceholder)
				}

				// This is a string placeholder: {"..."}
				eo := strings.IndexByte(input[startPlaceholder+2:], '}')
				if eo == -1 {
					return nil, fmt.Errorf("%w at index %d",
						ErrUnclosedStringPlaceholder, startPlaceholder)
				}
				eo += startPlaceholder + 2
				if eo >= len(input) || input[eo-1] != '"' {
					return nil, fmt.Errorf(
						"%w, expected } at index %d",
						ErrUnclosedStringPlaceholder, startPlaceholder)
				}

				str := input[startPlaceholder+2 : eo-1]
				if err := validateTextPlaceholderStr(eo, str); err != nil {
					return nil, err
				}

				buffer = append(buffer, Token{
					Str:   str,
					Index: startPlaceholder,
					Type:  TokenTypeStringPlaceholder,
				})
				i = eo + 1
				literalStart = i
				continue
			}

			i = specialIndex + 1 // Skip the '{'.
			// Use IndexByte to jump to the closing '}'.
			eo := strings.IndexByte(input[i:], '}')
			if eo == -1 {
				return nil, fmt.Errorf("%w at index %d",
					ErrUnclosedPlaceholder, startPlaceholder)
			}
			placeholderEnd := i + eo
			placeholder := input[i:placeholderEnd]
			wrapped := input[startPlaceholder : placeholderEnd+1]
			typ := match(placeholder, mapping)
			if typ == 0 {
				return nil, fmt.Errorf("%w at index %d",
					ErrUknownPlaceholder, startPlaceholder)
			}
			buffer = append(buffer, Token{
				Str:   wrapped,
				Index: startPlaceholder,
				Type:  typ,
			})
			i = placeholderEnd + 1 // Skip the '}'
			literalStart = i
		}
	}
	flushLiteral(i)
	return buffer, nil
}

func match(s string, mapping map[string]TokenType) TokenType {
	for k, v := range mapping {
		if strings.EqualFold(s, k) {
			return v
		}
	}
	return 0
}

// TIK is a parsed and validated textual internationalization token.
type TIK struct {
	Tokens Tokens
}

// Placeholders returns an iterators that iterates over placeholder tokens.
func (t TIK) Placeholders() iter.Seq2[int, Token] {
	return func(yield func(int, Token) bool) {
		i := 0
		for _, t := range t.Tokens {
			if t.Type == TokenTypeStringLiteral {
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
	return &Parser{
		t:      Tokenizer{},
		tokBuf: make(Tokens, 0, 16),
		conf:   conf,
	}
}

// ParseFn is similar to Parse but avoid copying the token buffer
// and instead uses the original buffer of the parser in the tik provided to fn.
//
// WARNING: Do not alias and use the token slice once fn returns!
func (p *Parser) ParseFn(input string, fn func(tik TIK) error) (err error) {
	p.tokBuf = p.tokBuf[:0] // Reset buffer.
	p.tokBuf, err = p.t.Tokenize(p.tokBuf, input, p.conf)
	if err != nil {
		return err
	}
	return fn(TIK{Tokens: p.tokBuf})
}

// Parse parses input and returns a validated TIK, otherwise returns an error.
// The tokens slice in the returned TIK is a copy of the buffer and doesn't alias
// the internal parser buffer.
// If you want to avoid the token buffer copy use ParseFn with caution instead.
func (p *Parser) Parse(input string) (tik TIK, err error) {
	err = p.ParseFn(input, func(ref TIK) error {
		cp := make(Tokens, len(ref.Tokens))
		copy(cp, ref.Tokens)
		tik.Tokens = cp
		return nil
	})
	return tik, err
}

func validateTextPlaceholderStr(index int, s string) error {
	if s == "" {
		return fmt.Errorf("%w at index %d", ErrStringPlaceholderEmpty, index)
	}

	runeL, _ := utf8.DecodeRuneInString(s)     // First.
	runeR, _ := utf8.DecodeLastRuneInString(s) // Last.

	if unicode.IsSpace(runeL) || unicode.IsSpace(runeR) {
		return ErrStringPlaceholderInvSpace
	}

	if iRevSol := strings.IndexAny(s, "\\{}\""); iRevSol != -1 {
		return fmt.Errorf(
			"%w at index %d", ErrStringPlaceholderIllegalChars, index+iRevSol)
	}

	return nil
}
