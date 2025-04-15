package tik

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Config defines the TIK environment configuration.
type Config struct {
	MagicConstants MagicConstants
}

// Validate returns an error if the config is invalid, otherwise returns nil.
func (c *Config) Validate() error {
	return validateCustomMagicConstants(c.MagicConstants)
}

// MagicConstants defines the magic constants used in the configured environment.
type MagicConstants struct {
	Number              string                     // {3}
	CardinalPluralStart string                     // {2 ...}
	OrdinalPlural       MagicConstantOrdinalPlural // {4th}

	GenderPronouns       []string // {they}, {them}, {their}, {theirs}, {themself}
	TimeShort            string   // {3:45PM}
	TimeShortSeconds     string   // {3:45:30PM}
	TimeFullMonthAndDay  string   // {April 2}
	TimeShortMonthAndDay string   // {Apr 2}
	TimeFullMonthAndYear string   // {Apr 2025}
	TimeWeekday          string   // {Monday}
	TimeDateAndShort     string   // {April 2, 3:45PM}
	TimeYear             string   // {2025}
	TimeFull             string   // {April 2, 3:45:30PM}
	CurrencyRounded      string   // {$1}
	CurrencyFull         string   // {$1.20}
	CurrencyCodeRounded  string   // {USD 1}
	CurrencyCodeFull     string   // {USD 1.20}
}

var defaultConfig = &Config{
	MagicConstants: MagicConstants{
		Number:              "3",
		CardinalPluralStart: "2",
		OrdinalPlural: MagicConstantOrdinalPlural{
			Constant:         "4th",
			DefaultICUSuffix: "th",
		},
		GenderPronouns:       []string{"they", "them", "their", "theirs", "themself"},
		TimeShort:            "3:45PM",
		TimeShortSeconds:     "3:45:30PM",
		TimeFullMonthAndDay:  "April 2",
		TimeShortMonthAndDay: "Apr 2",
		TimeFullMonthAndYear: "Apr 2025",
		TimeWeekday:          "Monday",
		TimeDateAndShort:     "April 2, 3:45PM",
		TimeYear:             "2025",
		TimeFull:             "April 2, 3:45:30PM",
		CurrencyRounded:      "$1",
		CurrencyFull:         "$1.20",
		CurrencyCodeRounded:  "USD 1",
		CurrencyCodeFull:     "USD 1.20",
	},
}

type MagicConstantOrdinalPlural struct {
	// Constant is the magic TIK constant
	Constant string

	// DefaultICUSuffix is used during ICU message generation.
	DefaultICUSuffix string
}

// DefaultConfig returns a deep copy of the original default config struct.
func DefaultConfig() *Config {
	cp := *defaultConfig
	cp.MagicConstants.GenderPronouns = make(
		[]string, len(defaultConfig.MagicConstants.GenderPronouns),
	)
	copy(cp.MagicConstants.GenderPronouns, defaultConfig.MagicConstants.GenderPronouns)
	return &cp
}

func validateCustomMagicConstants(m MagicConstants) error {
	byStr := make(map[string]struct{}, 20)
	check := func(v string) error {
		if err := validateMagicPlaceholder(v); err != nil {
			return err
		}
		if _, ok := byStr[v]; ok {
			return fmt.Errorf("%w: %q", ErrConfMagicConstantNonUnique, v)
		}
		byStr[v] = struct{}{}
		return nil
	}
	for _, v := range [...]string{
		m.Number,
		m.CardinalPluralStart,
		m.OrdinalPlural.Constant,
		m.TimeShort,
		m.TimeShortSeconds,
		m.TimeFullMonthAndDay,
		m.TimeShortMonthAndDay,
		m.TimeFullMonthAndYear,
		m.TimeWeekday,
		m.TimeDateAndShort,
		m.TimeYear,
		m.TimeFull,
		m.CurrencyRounded,
		m.CurrencyFull,
		m.CurrencyCodeRounded,
		m.CurrencyCodeFull,
	} {
		if err := check(v); err != nil {
			return err
		}
	}

	if m.OrdinalPlural.DefaultICUSuffix == "" {
		return fmt.Errorf("%w: ordinal plural ICU suffix", ErrConfMissingDefault)
	}

	if len(m.GenderPronouns) == 0 {
		return fmt.Errorf("%w: no gender pronouns", ErrConfMagicConstantInvalid)
	}
	for _, v := range m.GenderPronouns {
		if err := check(v); err != nil {
			return err
		}
	}
	return nil
}

func validateMagicPlaceholder(s string) error {
	if s == "" {
		return ErrConfMagicConstantInvalid
	}
	if strings.ContainsAny(s, "\"{}") {
		return ErrConfMagicConstantInvalid
	}

	runeL, _ := utf8.DecodeRuneInString(s)     // First.
	runeR, _ := utf8.DecodeLastRuneInString(s) // Last.

	if unicode.IsSpace(runeL) {
		return ErrConfMagicConstantInvalid
	}
	if unicode.IsSpace(runeR) {
		return ErrConfMagicConstantInvalid
	}
	return nil
}
