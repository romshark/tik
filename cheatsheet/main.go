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
			ID:     "cardinal-pluralization",
			Header: `<code>{"2 ..."}</code> Cardinal Pluralization`,
			Description: `Pluralizes <code>...</code> to the correct <a
					href="https://www.unicode.org/cldr/charts/47/supplemental/language_plural_rules.html">CLDR
					plural form</a>.`,
			ExampleTIK: `You have {2 unread messages} in {2 groups}.`,
			ICU:        `You have {var0, plural, one{# unread message} other{# unread messages}} in {var1, plural, one{# group} other{# groups}}}`,
			Examples: []Example{
				{
					Data: `{"var0": 4, "var1": 1}`,
					Text: `You have 4 unread messages in 1 group`,
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
	})
	if err := p.Render(context.Background(), of); err != nil {
		panic(err)
	}
	fmt.Println("written to", *fOutFile)
}
