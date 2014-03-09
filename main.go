package main

import (
	"bufio"
	"fmt"
	"os"
	"./interpreter"
)

func main() {
	env := glisp.NewGlisp()
	lexer := glisp.NewLexerFromStream(bufio.NewReader(os.Stdin))

	expressions, err := glisp.ParseTokens(env, lexer)
	if err != nil {
		fmt.Printf("Error on line %d: %v\n", lexer.Linenum(), err)
		os.Exit(-1)
	}

	for _, expr := range expressions {
		fmt.Println(expr.SexpString())
	}
}
