package tik_test

import (
	"fmt"

	tik "github.com/romshark/tik/tik-go"
)

func ExampleParser() {
	const input = `{name} had {# messages} on {date-medium} at {time-full}`

	conf := tik.DefaultConfig
	parser := tik.NewParser(conf)

	tk, err := parser.Parse(input)
	if err != nil {
		fmt.Println("ERR:", err)
		return
	}

	fmt.Println(" ")
	fmt.Println("TOKENS:", len(tk.Tokens))
	for _, x := range tk.Tokens {
		fmt.Printf("%d-%d: %q (%s)\n", x.IndexStart, x.IndexEnd,
			x.String(input), x.Type.String())
	}

	icu := tik.NewICUTranslator(conf)

	fmt.Println("")
	fmt.Println("ICU Message:")
	fmt.Println(icu.TIK2ICU(tk))

	// Output:
	// TOKENS: 9
	// 0-6: "{name}" (text with gender)
	// 6-11: " had " (literal)
	// 11-14: "{# " (pluralization)
	// 14-22: "messages" (literal)
	// 22-23: "}" (pluralization block end)
	// 23-27: " on " (literal)
	// 27-40: "{date-medium}" (date medium)
	// 40-44: " at " (literal)
	// 44-55: "{time-full}" (time full)
	//
	// ICU Message:
	// {var0} had {var1, plural, other {# messages}} on {var2, date, medium} at {var3, time, full}
}
