// make even the Repl() available from the glisp library
package gdsl

import (
	"bufio"
	"fmt"
	"os"
	"runtime/pprof"
	"strings"
)

var precounts map[string]int
var postcounts map[string]int

func CountPreHook(env *Glisp, name string, args []Sexp) {
	precounts[name] += 1
}

func CountPostHook(env *Glisp, name string, retval Sexp) {
	postcounts[name] += 1
}

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

// NB at the moment this doesn't track comment and strings state,
// so it will fail if unbalanced '(' are found in either.
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

func processDumpCommand(env *Glisp, args []string) {
	if len(args) == 0 {
		env.DumpEnvironment()
	} else {
		err := env.DumpFunctionByName(args[0])
		if err != nil {
			fmt.Println(err)
		}
	}
}

func Repl(env *Glisp, cfg *GlispConfig) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("gdsl version %s\n", Version())
	fmt.Printf("gdslext version %s\n", cfg.ExtensionsVersion)

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

		if expr != SexpNull {
			fmt.Println(expr.SexpString())
		}
	}
}

func runScript(env *Glisp, fname string, cfg *GlispConfig) {
	file, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	err = env.LoadFile(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = env.Run()
	if cfg.CountFuncCalls {
		fmt.Println("Pre:")
		for name, count := range precounts {
			fmt.Printf("\t%s: %d\n", name, count)
		}
		fmt.Println("Post:")
		for name, count := range postcounts {
			fmt.Printf("\t%s: %d\n", name, count)
		}
	}
	if err != nil {
		fmt.Print(env.GetStackTrace(err))
		if cfg.ExitOnFailure {
			os.Exit(-1)
		}
		Repl(env, cfg)
	}
}

func (env *Glisp) StandardSetup() {
	env.ImportEval()
	env.ImportRequire()
	env.ImportTime()
	env.ImportMsgpackMap()

	defmap := `(defmac defmap [name] ^(defn ~name [& rest] (msgmap (quote ~name) rest)))`
	_, err := env.EvalString(defmap)
	panicOn(err)

	colonOp := `(defmac : [key hmap & def] ^(hget ~hmap (quote ~key) ~@def))`
	_, err = env.EvalString(colonOp)
	panicOn(err)
}

// like main() for a standalone repl, now in library
func ReplMain(cfg *GlispConfig, registerExtsFunc func(env *Glisp)) {
	env := NewGlisp()
	env.StandardSetup()

	registerExtsFunc(env)

	if cfg.CpuProfile != "" {
		f, err := os.Create(cfg.CpuProfile)
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

	precounts = make(map[string]int)
	postcounts = make(map[string]int)

	if cfg.CountFuncCalls {
		env.AddPreHook(CountPreHook)
		env.AddPostHook(CountPostHook)
	}

	args := cfg.Flags.Args()
	if len(args) > 0 {
		runScript(env, args[0], cfg)
	} else {
		Repl(env, cfg)
	}

	if cfg.MemProfile != "" {
		f, err := os.Create(cfg.MemProfile)
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
