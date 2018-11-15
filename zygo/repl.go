package zygo

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
)

var precounts map[string]int
var postcounts map[string]int

func CountPreHook(env *Zlisp, name string, args []Sexp) {
	precounts[name] += 1
}

func CountPostHook(env *Zlisp, name string, retval Sexp) {
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

var continuationPrompt = "... "

func (pr *Prompter) getExpressionOrig(reader *bufio.Reader) (readin string, err error) {
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

// liner reads Stdin only. If noLiner, then we read from reader.
func (pr *Prompter) getExpressionWithLiner(env *Zlisp, reader *bufio.Reader, noLiner bool) (readin string, xs []Sexp, err error) {

	var line, nextline string

	if noLiner {
		fmt.Printf(pr.prompt)
		line, err = getLine(reader)
	} else {
		line, err = pr.Getline(nil)
	}
	if err != nil {
		return "", nil, err
	}

	err = UnexpectedEnd
	var x []Sexp

	// test parse, but don't load or generate bytecode
	env.parser.ResetAddNewInput(bytes.NewBuffer([]byte(line + "\n")))
	x, err = env.parser.ParseTokens()
	//P("\n after ResetAddNewInput, err = %v. x = '%s'\n", err, SexpArray(x).SexpString())

	if len(x) > 0 {
		xs = append(xs, x...)
	}

	for err == ErrMoreInputNeeded || err == UnexpectedEnd || err == ResetRequested {
		if noLiner {
			fmt.Printf(continuationPrompt)
			nextline, err = getLine(reader)
		} else {
			nextline, err = pr.Getline(&continuationPrompt)
		}
		if err != nil {
			return "", nil, err
		}
		env.parser.NewInput(bytes.NewBuffer([]byte(nextline + "\n")))
		x, err = env.parser.ParseTokens()
		if len(x) > 0 {
			for i := range x {
				if x[i] == SexpEnd {
					P("found an SexpEnd token, omitting it")
					continue
				}
				xs = append(xs, x[i])
			}
		}
		switch err {
		case nil:
			line += "\n" + nextline
			Q("no problem parsing line '%s' into '%s', proceeding...\n", line, (&SexpArray{Val: x, Env: env}).SexpString(nil))
			return line, xs, nil
		case ResetRequested:
			continue
		case ErrMoreInputNeeded:
			continue
		default:
			return "", nil, fmt.Errorf("Error on line %d: %v\n", env.parser.lexer.Linenum(), err)
		}
	}
	return line, xs, nil
}

func processDumpCommand(env *Zlisp, args []string) {
	if len(args) == 0 {
		env.DumpEnvironment()
	} else {
		err := env.DumpFunctionByName(args[0])
		if err != nil {
			fmt.Println(err)
		}
	}
}

func Repl(env *Zlisp, cfg *ZlispConfig) {
	var reader *bufio.Reader
	if cfg.NoLiner {
		// reader is used if one wishes to drop the liner library.
		// Useful for not full terminal env, like under test.
		reader = bufio.NewReader(os.Stdin)
	}

	if cfg.Trace {
		// debug tracing
		env.debugExec = true
	}

	if !cfg.Quiet {
		if cfg.Sandboxed {
			fmt.Printf("zygo [sandbox mode] version %s\n", Version())
		} else {
			fmt.Printf("zygo version %s\n", Version())
		}
		fmt.Printf("press tab (repeatedly) to get completion suggestions. Shift-tab goes back. Ctrl-d to exit.\n")
	}
	var pr *Prompter // can be nil if noLiner
	if !cfg.NoLiner {
		pr = NewPrompter(cfg.Prompt)
		defer pr.Close()
	} else {
		pr = &Prompter{prompt: cfg.Prompt}
	}
	infixSym := env.MakeSymbol("infix")

	for {
		line, exprsInput, err := pr.getExpressionWithLiner(env, reader, cfg.NoLiner)
		//Q("\n exprsInput(len=%d) = '%v'\n line = '%s'\n", len(exprsInput), (&SexpArray{Val: exprsInput}).SexpString(nil), line)
		if err != nil {
			fmt.Println(err)
			if err == io.EOF {
				os.Exit(0)
			}
			env.Clear()
			continue
		}

		parts := strings.Split(strings.Trim(line, " "), " ")
		//parts := strings.Split(line, " ")
		if len(parts) == 0 {
			continue
		}
		first := strings.Trim(parts[0], " ")

		if first == ".quit" {
			break
		}

		if first == ".cd" {
			if len(parts) < 2 {
				fmt.Printf("provide directory path to change to.\n")
				continue
			}
			err := os.Chdir(parts[1])
			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			pwd, err := os.Getwd()
			if err == nil {
				fmt.Printf("cur dir: %s\n", pwd)
			} else {
				fmt.Printf("error: %s\n", err)
			}
			continue
		}

		// allow & at the repl to take the address of an expression
		if len(first) > 0 && first[0] == '&' {
			//P("saw & at repl, first='%v', parts='%#v'. exprsInput = '%#v'", first, parts, exprsInput)
			exprsInput = []Sexp{MakeList(exprsInput)}
		}

		// allow * at the repl to dereference a pointer and print
		if len(first) > 0 && first[0] == '*' {
			//P("saw * at repl, first='%v', parts='%#v'. exprsInput = '%#v'", first, parts, exprsInput)
			exprsInput = []Sexp{MakeList(exprsInput)}
		}

		if first == ".dump" {
			processDumpCommand(env, parts[1:])
			continue
		}

		if first == ".gls" {
			fmt.Printf("\nScopes:\n")
			prev := env.showGlobalScope
			env.showGlobalScope = true
			err = env.ShowStackStackAndScopeStack()
			env.showGlobalScope = prev
			if err != nil {
				fmt.Printf("%s\n", err)
			}
			continue
		}

		if first == ".ls" {
			err := env.ShowStackStackAndScopeStack()
			if err != nil {
				fmt.Println(err)
			}
			continue
		}

		if first == ".verb" {
			Verbose = !Verbose
			fmt.Printf("verbose: %v.\n", Verbose)
			continue
		}

		if first == ".debug" {
			env.debugExec = true
			fmt.Printf("instruction debugging on.\n")
			continue
		}

		if first == ".undebug" {
			env.debugExec = false
			fmt.Printf("instruction debugging off.\n")
			continue
		}

		var expr Sexp
		n := len(exprsInput)
		if n > 0 {
			infixWrappedSexp := MakeList([]Sexp{infixSym, &SexpArray{Val: exprsInput, Env: env}})
			expr, err = env.EvalExpressions([]Sexp{infixWrappedSexp})
		} else {
			line = env.ReplLineInfixWrap(line)
			expr, err = env.EvalString(line + " ") // print standalone variables
		}
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
			// try to print strings more elegantly!
			switch e := expr.(type) {
			case *SexpStr:
				if e.backtick {
					fmt.Printf("`%s`\n", e.S)
				} else {
					fmt.Printf("%s\n", strconv.Quote(e.S))
				}
			default:
				switch sym := expr.(type) {
				case Selector:
					Q("repl calling RHS() on Selector")
					rhs, err := sym.RHS(env)
					if err != nil {
						Q("repl problem in call to RHS() on SexpSelector: '%v'", err)
						fmt.Print(env.GetStackTrace(err))
						env.Clear()
						continue
					} else {
						Q("got back rhs of type %T", rhs)
						fmt.Println(rhs.SexpString(nil))
						continue
					}
				case *SexpSymbol:
					if sym.isDot {
						resolved, err := dotGetSetHelper(env, sym.name, nil)
						if err != nil {
							fmt.Print(env.GetStackTrace(err))
							env.Clear()
							continue
						}
						fmt.Println(resolved.SexpString(nil))
						continue
					}
				}
				fmt.Println(expr.SexpString(nil))
			}
		}
	}
}

func runScript(env *Zlisp, fname string, cfg *ZlispConfig) {
	file, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	err = env.LoadFile(file)
	if err != nil {
		fmt.Println(err)
		if cfg.ExitOnFailure {
			os.Exit(-1)
		}
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

func (env *Zlisp) StandardSetup() {
	env.ImportBaseTypes()
	env.ImportEval()
	env.ImportTime()
	env.ImportPackageBuilder()
	env.ImportMsgpackMap()

	defmap := `(defmac defmap [name] ^(defn ~name [& rest] (msgmap (quote ~name) rest)))`
	_, err := env.EvalString(defmap)
	panicOn(err)

	//	colonOp := `(defmac : [key hmap & def] ^(hget ~hmap (quote ~key) ~@def))`
	//	_, err = env.EvalString(colonOp)
	//	panicOn(err)

	rangeMacro := `(defmac range [key value myhash & body]
  ^(let [n (len ~myhash)]
      (for [(def i 0) (< i n) (def i (+ i 1))]
        (begin
          (mdef (quote ~key) (quote ~value) (hpair ~myhash i))
          ~@body))))`
	_, err = env.EvalString(rangeMacro)
	panicOn(err)

	reqMacro := `(defmac req [a] ^(source (sym2str (quote ~a))))`
	_, err = env.EvalString(reqMacro)
	panicOn(err)

	incrMacro := `(defmac ++ [a] ^(set ~a (+ ~a 1)))`
	_, err = env.EvalString(incrMacro)
	panicOn(err)

	incrEqMacro := `(defmac += [a b] ^(set ~a (+ ~a ~b)))`
	_, err = env.EvalString(incrEqMacro)
	panicOn(err)

	decrMacro := `(defmac -- [a] ^(set ~a (- ~a 1)))`
	_, err = env.EvalString(decrMacro)
	panicOn(err)

	decrEqMacro := `(defmac -= [a b] ^(set ~a (- ~a ~b)))`
	_, err = env.EvalString(decrEqMacro)
	panicOn(err)

	env.ImportChannels()
	env.ImportGoroutines()
	env.ImportRegex()
	env.ImportRandom()

	gob.Register(SexpHash{})
	gob.Register(SexpArray{})
}

// like main() for a standalone repl, now in library
func ReplMain(cfg *ZlispConfig) {
	var env *Zlisp
	if cfg.LoadDemoStructs {
		RegisterDemoStructs()
	}
	if cfg.Sandboxed {
		env = NewZlispSandbox()
	} else {
		env = NewZlisp()
	}
	env.StandardSetup()
	if cfg.LoadDemoStructs {
		// avoid data conflicts by only loading these in demo mode.
		env.ImportDemoData()
	}

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

	if cfg.Command != "" {
		_, err := env.EvalString(cfg.Command)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	runRepl := true
	args := cfg.Flags.Args()
	if len(args) > 0 {
		runRepl = false
		runScript(env, args[0], cfg)
		if cfg.AfterScriptDontExit {
			runRepl = true
		}
	}
	if runRepl {
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

func (env *Zlisp) ReplLineInfixWrap(line string) string {
	return "{" + line + "}"
}
