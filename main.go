package main

import (
	"fmt"
	"os"
	"./interpreter"
)

func main() {
	env := glisp.NewGlisp()
	err := env.LoadFile(os.Stdin)

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	env.DumpEnvironment()

	err = env.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	expr, err := env.PopResult()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	fmt.Println(expr.SexpString())
}
