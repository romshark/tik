package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	defaultOutPath := filepath.Join("dist", "index.html")

	fOutFile := flag.String("o", defaultOutPath, "output file path")
	flag.Parse()

	dir := filepath.Dir(*fOutFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}
	of, err := os.OpenFile(*fOutFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}

	p := templatePage("TIK Cheatsheet", []Section{
		{
			ID:          "string",
			Header:      `<code>{"..."}</code> String`,
			Description: `Replaces <code>...</code> with an arbitrary string value.`,
			ExampleTIK:  `Avoid comparing {"apples"} to {"oranges"}.`,
			ICU:         `Avoid comparing {var0} to {var1}`,
			Examples: []Example{
				{
					Data: `{"var0": "cats 🐈", "var1": "dogs 🐕"}`,
					Text: `Avoid comparing cats 🐈 to dogs 🐕`,
				},
				{
					Data: `{"var0": "bananas", "var1": "anything really"}`,
					Text: `Avoid comparing bananas to anything really`,
				},
			},
		},
		{
			ID:          "integer",
			Header:      `<code>{7}</code> Integer`,
			Description: `Represents an integer number value.`,
			ExampleTIK:  `Your number is {7}`,
			ICU:         `Your number is {var0, number, integer}`,
			Examples: []Example{
				{
					Data: `{"var0": 42}`,
					Text: `Your number is 42`,
				},
				{
					Data: `{"var0": 0}`,
					Text: `Your number is 0`,
				},
			},
		},
		{
			ID:          "number",
			Header:      `<code>{3.14}</code> Number`,
			Description: `Represents a decimal number value.`,
			ExampleTIK:  `It's off by {3.14} degrees.`,
			ICU:         `It's {var0, number} degrees.`,
			Examples: []Example{
				{
					Data: `{"var0": 42}`,
					Text: `It's off by 42 degrees.`,
				},
				{
					Data: `{"var0": 0.628}`,
					Text: `It's off by 0.628 degrees.`,
				},
			},
		},
		{
			ID:     "cardinal-pluralization",
			Header: `<code>{"2 ..."}</code> Cardinal Pluralization`,
			Description: `Pluralizes <code>...</code> to the correct <a
					href="https://www.unicode.org/cldr/charts/47/supplemental/language_plural_rules.html">CLDR
					plural form</a>.`,
			ExampleTIK: `You have {2 unread messages} in {2 groups}.`,
			ICU:        `You have {var0, plural, one{# unread message} other{# unread messages}} in {var1, plural, one{# group} other{# groups}}}`,
			Examples: []Example{
				{
					Data: `{"var0": 4, "var1": 3}`,
					Text: `You have 4 unread messages in 3 groups`,
				},
				{
					Data: `{"var0": 1, "var1": 1}`,
					Text: `You have 1 unread message in 1 group`,
				},
			},
		},
		{
			ID:     "ordinal-pluralization",
			Header: `<code>{4th}</code> Ordinal Pluralization`,
			Description: `Represents a <a
					href="https://www.unicode.org/cldr/charts/47/supplemental/language_plural_rules.html">CLDR
					ordinal plural</a> value.`,
			ExampleTIK: `You are {4th} in the queue.`,
			ICU:        `You are {var0, selectordinal, one{#st} two{#nd} few{#rd} other{#th}} in the queue.`,
			Examples: []Example{
				{
					Data: `{"var0": 1}`,
					Text: `You are 1st in the queue.`,
				},
				{
					Data: `{"var0": 9}`,
					Text: `You are 9th in the queue.`,
				},
			},
		},
		{
			ID:          "gender",
			Header:      `<code>{They}</code> Gender`,
			Description: `Represents a gender subject.`,
			Aliases:     []string{"them", "their", "theirs", "themself"},
			ExampleTIK:  `{They} received the message!`,
			ICU:         `{var0, select, male{He} female{She} other{They}} received the message!`,
			Examples: []Example{
				{
					Data: `{"var0": "male"}`,
					Text: `He received the message`,
				},
				{
					Data: `{"var0": "female"}`,
					Text: `She received the message`,
				},
				{
					Data: `{"var0": "neutral"}`,
					Text: `They received the message`,
				},
			},
		},
		{
			ID:          "date-short",
			Header:      `<code>{7/16/99}</code> Date Short`,
			Description: `Represents a <a href="https://cldr.unicode.org/translation/date-time/date-time-patterns#basic-date-formats">CLDR short date</a>.`,
			ExampleTIK:  `Today is {7/16/99}`,
			ICU:         `Today is {var0, date, short}`,
			Examples: []Example{
				{
					Data: `{"var0": "2025-01-26T16:20:50Z"}`,
					Text: `Today is 1/26/25`,
				},
				{
					Data: `{"var0": "1998-09-02T20:20:50Z"}`,
					Text: `Today is 9/2/98`,
				},
			},
		},
		{
			ID:          "date-medium",
			Header:      `<code>{Jul 16, 1999}</code> Date Medium`,
			Description: `Represents a <a href="https://cldr.unicode.org/translation/date-time/date-time-patterns#basic-date-formats">CLDR medium date</a>.`,
			ExampleTIK:  `Today is {Jul 16, 1999}`,
			ICU:         `Today is {var0, date, medium}`,
			Examples: []Example{
				{
					Data: `{"var0": "2025-01-26T16:20:50Z"}`,
					Text: `Today is Jan 26, 2025`,
				},
				{
					Data: `{"var0": "1998-09-02T20:20:50Z"}`,
					Text: `Today is Sep 2, 1998`,
				},
			},
		},
		{
			ID:          "date-long",
			Header:      `<code>{July 16, 1999}</code> Date Long`,
			Description: `Represents a <a href="https://cldr.unicode.org/translation/date-time/date-time-patterns#basic-date-formats">CLDR long date</a>.`,
			ExampleTIK:  `Today is {July 16, 1999}`,
			ICU:         `Today is {var0, date, long}`,
			Examples: []Example{
				{
					Data: `{"var0": "2025-01-26T16:20:50Z"}`,
					Text: `Today is January 26, 2025`,
				},
				{
					Data: `{"var0": "1998-09-02T20:20:50Z"}`,
					Text: `Today is September 2, 1998`,
				},
			},
		},
		{
			ID:          "date-full",
			Header:      `<code>{Friday, July 16, 1999}</code> Date Full`,
			Description: `Represents a <a href="https://cldr.unicode.org/translation/date-time/date-time-patterns#basic-date-formats">CLDR full date</a>.`,
			ExampleTIK:  `Today is {Friday, July 16, 1999}`,
			ICU:         `Today is {var0, date, full}`,
			Examples: []Example{
				{
					Data: `{"var0": "2025-01-26T16:20:50Z"}`,
					Text: `Today is Sunday, January 26, 2025`,
				},
				{
					Data: `{"var0": "1998-09-02T20:20:50Z"}`,
					Text: `Today is Wednesday, September 2, 1998`,
				},
			},
		},
		{
			ID:          "time-short",
			Header:      `<code>{10:30 pm}</code> Time Short`,
			Description: `Represents a <a href="https://cldr.unicode.org/translation/date-time/date-time-patterns#basic-time-formats">CLDR short time</a>.`,
			ExampleTIK:  `It's {10:30 pm}`,
			ICU:         `It's {var0, time, short}`,
			Examples: []Example{
				{
					Data: `{"var0": "2025-01-26T16:20:50Z"}`,
					Text: `It's 4:20 pm`,
				},
				{
					Data: `{"var0": "1998-09-02T20:20:50-07:00"}`,
					Text: `It's 8:20 pm`,
				},
			},
		},
		{
			ID:          "time-medium",
			Header:      `<code>{10:30:45 pm}</code> Time Medium`,
			Description: `Represents a <a href="https://cldr.unicode.org/translation/date-time/date-time-patterns#basic-time-formats">CLDR medium time</a>.`,
			ExampleTIK:  `It's {10:30:45 pm}`,
			ICU:         `It's {var0, time, medium}`,
			Examples: []Example{
				{
					Data: `{"var0": "2025-01-26T16:20:50Z"}`,
					Text: `It's 4:20:50 pm`,
				},
				{
					Data: `{"var0": "1998-09-02T20:20:50-07:00"}`,
					Text: `It's 8:20:50 pm`,
				},
			},
		},
		{
			ID:          "time-long",
			Header:      `<code>{10:30:45 pm PDT}</code> Time Long`,
			Description: `Represents a <a href="https://cldr.unicode.org/translation/date-time/date-time-patterns#basic-time-formats">CLDR long time</a>.`,
			ExampleTIK:  `It's {10:30:45 pm PDT}`,
			ICU:         `It's {var0, time, long}`,
			Examples: []Example{
				{
					Data: `{"var0": "2025-01-26T16:20:50Z"}`,
					Text: `It's 4:20:50 pm UTC`,
				},
				{
					Data: `{"var0": "1998-09-02T20:20:50-07:00"}`,
					Text: `It's 8:20:50 pm PDT`,
				},
			},
		},
		{
			ID:          "time-full",
			Header:      `<code>{10:30:45 pm Pacific Daylight Time}</code> Time Full`,
			Description: `Represents a <a href="https://cldr.unicode.org/translation/date-time/date-time-patterns#basic-time-formats">CLDR full time</a>.`,
			ExampleTIK:  `It's {10:30:45 pm Pacific Daylight Time}`,
			ICU:         `It's {var0, time, full}`,
			Examples: []Example{
				{
					Data: `{"var0": "2025-01-26T16:20:50Z"}`,
					Text: `It's 4:20:50 pm UTC`,
				},
				{
					Data: `{"var0": "1998-09-02T20:20:50-07:00"}`,
					Text: `It's 8:20:50 pm Pacific Daylight Time`,
				},
			},
		},
		{
			ID:          "currency",
			Header:      `<code>{$1}</code> Currency`,
			Description: `Represents an amount of money.`,
			ExampleTIK:  `The price is {$1}`,
			ICU:         `The price is {var0, number, ::currency/auto}`,
			Examples: []Example{
				{
					Data: `{"var0": {"currency": "USD", "amount": 12.99}}`,
					Text: `The price is $12.99`,
				},
			},
		},
	})
	if err := p.Render(context.Background(), of); err != nil {
		panic(err)
	}
	fmt.Println("written to", *fOutFile)
}
