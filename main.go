package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	tokens, err := LexStream(bufio.NewReader(os.Stdin))

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	expressions, err := ParseTokens(tokens)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	fmt.Println(tokens)

	for _, expr := range expressions {
		fmt.Println(expr.SexpString())
	}
}
