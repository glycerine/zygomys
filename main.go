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

	gen := glisp.NewGenerator(env)
	err = gen.GenerateAll(expressions)
	if err != nil {
		fmt.Printf("generate error: %v\n", err)
		os.Exit(-1)
	}

	for _, instr := range gen.GetInstructions() {
		fmt.Println(instr.InstrString())
	}
}
