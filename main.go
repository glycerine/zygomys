package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	lexer := NewLexerFromStream(bufio.NewReader(os.Stdin))

	expressions, err := ParseTokens(lexer)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	for _, expr := range expressions {
		fmt.Println(expr.SexpString())
	}
}
