package zygo

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

var continuationPrompt = ">> "

func (pr *Prompter) getExpressionOrig(reader *bufio.Reader) (string, error) {

	line, err := getLine(reader)
	if err != nil {
		return "", err
	}

	for !isBalanced(line) {
		fmt.Printf(continuationPrompt)
		nextline, err := getLine(reader)
		if err != nil {
			return "", err
		}
		line += "\n" + nextline
	}
	return line, nil
}

// reads Stdin only
func (pr *Prompter) getExpressionWithLiner() (string, error) {

	line, err := pr.Getline(nil)
	if err != nil {
		return "", err
	}

	for !isBalanced(line) {
		nextline, err := pr.Getline(&continuationPrompt)
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
	// used if one wishes to drop the liner library and use
	// pr.getExpressionOrig() instead.
	//reader := bufio.NewReader(os.Stdin)

	// debug
	// env.debugExec = true

	fmt.Printf("zygo version %s\n", Version())
	fmt.Printf("press tab (repeatedly) to get completion suggestions. Shift-tab goes back.\n")
	pr := NewPrompter()
	defer pr.Close()

	for {
		//line, err := pr.getExpressionOrig(reader)
		line, err := pr.getExpressionWithLiner()
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

		if parts[0] == "debug" {
			env.debugExec = true
			fmt.Printf("instruction debugging on.\n")
			continue
		}

		if parts[0] == "undebug" {
			env.debugExec = false
			fmt.Printf("instruction debugging off.\n")
			continue
		}

		expr, err := env.EvalString(line)
		switch err {
		case nil:
		case NoExpressionsFound:
			env.Clear()
			continue
		default:
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
	env.ImportTime()
	env.ImportMsgpackMap()

	defmap := `(defmac defmap [name] ^(defn ~name [& rest] (msgmap (quote ~name) rest)))`
	_, err := env.EvalString(defmap)
	panicOn(err)

	colonOp := `(defmac : [key hmap & def] ^(hget ~hmap (quote ~key) ~@def))`
	_, err = env.EvalString(colonOp)
	panicOn(err)

	rangeMacro := `(defmac range [key value my-hash & body]
  ^(let [n (len ~my-hash)]
      (for [(def i 0) (< i n) (def i (+ i 1))]
        (begin
          (mdef (quote ~key) (quote ~value) (hpair ~my-hash i))
          ~@body))))`
	_, err = env.EvalString(rangeMacro)
	panicOn(err)

	reqMacro := `(defmac req [a] ^(source (sym2str (quote ~a))))`
	_, err = env.EvalString(reqMacro)
	panicOn(err)

	slurpMacro := `(defmac slurp [a] ^(slurpf (sym2str (quote ~a))))`
	_, err = env.EvalString(slurpMacro)
	panicOn(err)

	owriteMacro := `(defmac owrite [array filepath] ^(owritef ~array (sym2str (quote ~filepath))))`
	_, err = env.EvalString(owriteMacro)
	panicOn(err)

	writeMacro := `(defmac write [array filepath] ^(writef ~array (sym2str (quote ~filepath))))`
	_, err = env.EvalString(writeMacro)
	panicOn(err)

	systemMacro := `(defmac $ [ & body] ^(system (map sym2str (quote ~@body))))`
	_, err = env.EvalString(systemMacro)
	panicOn(err)

	env.ImportChannels()
	env.ImportGoroutines()
	env.ImportRegex()
	env.ImportRandom()

}

// like main() for a standalone repl, now in library
func ReplMain(cfg *GlispConfig) {
	env := NewGlisp()
	env.StandardSetup()

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
