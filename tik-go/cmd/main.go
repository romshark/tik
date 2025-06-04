package main

import (
	"fmt"
	"os"

	"github.com/romshark/tik/tik-go"
)

const input = `{"John"} has {2 messages} with a similar {"status"}.`

func main() {
	conf := tik.DefaultConfig()
	parser := tik.NewParser(conf)

	tk, err := parser.Parse(input)
	if err != nil {
		fmt.Println("ERR:", err)
		os.Exit(1)
	}

	fmt.Println(" ")
	fmt.Println("TOKENS: ", len(tk.Tokens))
	for _, x := range tk.Tokens {
		fmt.Printf("%d-%d: %q (%s)\n", x.IndexStart, x.IndexEnd, x.String(input), x.Type.String())
	}

	icu := tik.NewICUTranslator(conf)

	fmt.Println("")
	fmt.Println("ICU Message:")
	fmt.Println(icu.TIK2ICU(tk, map[int]tik.ICUModifier{
		0: {Gender: true}, // John
		2: {Plural: true}, // "status"
	}))
}
