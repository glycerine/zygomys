package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"flag"
	"runtime/pprof"

	"github.com/zhemao/glisp/interpreter"
	"github.com/zhemao/glisp/extensions"
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

func processDumpCommand(env *glisp.Glisp, args []string) {
	if len(args) == 0 {
		env.DumpEnvironment()
	} else {
		err := env.DumpFunctionByName(args[0])
		if err != nil {
			fmt.Println(err)
		}
	}
}

func repl(env *glisp.Glisp) {
	fmt.Printf("glisp version %s\n", glisp.Version())
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := getExpression(reader)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		parts := strings.Split(line, " ")
		if len(parts) == 0 {
			continue
		}

		if parts[0] == "quit" {
			break
		}

		if parts[0] == "dump" {
			processDumpCommand(env, parts[1:])
			continue
		}

		expr, err := env.EvalString(line)
		if err != nil {
			fmt.Print(env.GetStackTrace(err))
			env.Clear()
			continue
		}

		if expr != glisp.SexpNull {
			fmt.Println(expr.SexpString())
		}
	}
}

func runScript(env *glisp.Glisp, fname string) {
	file, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer file.Close()

	err = env.LoadFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	_, err = env.Run()

	if err != nil {
		fmt.Print(env.GetStackTrace(err))
		if *exitOnFailure {
			os.Exit(-1)
		}
		repl(env)
	}
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write mem profile to file")
var exitOnFailure = flag.Bool("exitonfail", false,
	"exit on failure instead of starting repl")

func main() {
	env := glisp.NewGlisp()
	env.ImportEval()
	glispext.ImportRandom(env)
	glispext.ImportTime(env)

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		defer pprof.StopCPUProfile()
	}

	args := flag.Args()
	if len(args) > 0 {
		runScript(env, args[0])
	} else {
		repl(env)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		defer f.Close()

		err = pprof.Lookup("heap").WriteTo(f, 1)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}
}
