package main

import (
	"fmt"
	"os"
	"bufio"
	"./interpreter"
)

func getLine(reader *bufio.Reader) (string, error) {
	line := make([]byte, 0)
	for {
		linepart, hasMore, err := reader.ReadLine()
		if err != nil {
			return "", err
		}
		line = append(line, linepart...)
		if !hasMore {
			break
		}
	}
	return string(line), nil
}

func isBalanced(str string) bool {
	parens := 0
	squares := 0

	for _, c := range str {
		switch c {
		case '(':
			parens++
		case ')':
			parens--
		case '[':
			squares++
		case ']':
			squares--
		}
	}

	return parens == 0 && squares == 0
}

func getExpression(reader *bufio.Reader) (string, error) {
	fmt.Printf("> ")
	line, err := getLine(reader)
	if err != nil {
		return "", err
	}
	for !isBalanced(line) {
		fmt.Printf(">> ")
		nextline, err := getLine(reader)
		if err != nil {
			return "", err
		}
		line += "\n" + nextline
	}
	return line, nil
}

func main() {
	env := glisp.NewGlisp()
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("glisp version %s\n", glisp.Version())

	for {
		line, err := getExpression(reader)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		if line == "quit" {
			break
		}
		if line == "dump" {
			env.DumpEnvironment()
			continue
		}

		err = env.LoadString(line)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		err = env.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		expr, err := env.PopResult()
		if err != nil {
			fmt.Println("nothing on the stack")
			os.Exit(-1)
		}
		fmt.Println(expr.SexpString())
	}

}
