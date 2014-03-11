package glisp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Glisp struct {
	datastack   *Stack
	scopestack  *Stack
	addrstack   *Stack
	symtable    map[string]int
	revsymtable map[int]string
	curfunc     SexpFunction
	mainfunc    SexpFunction
	pc          int
	nextsymbol  int
}

func NewGlisp() *Glisp {
	env := new(Glisp)
	env.datastack = NewStack()
	env.scopestack = NewStack()
	env.scopestack.PushScope()
	env.addrstack = NewStack()
	env.symtable = make(map[string]int)
	env.revsymtable = make(map[int]string)
	env.nextsymbol = 1

	for key, function := range BuiltinFunctions {
		env.AddFunction(key, function)
	}

	env.mainfunc = MissingFunction
	env.curfunc = MissingFunction
	env.pc = -1
	return env
}

func (env *Glisp) MakeSymbol(name string) SexpSymbol {
	symnum, ok := env.symtable[name]
	if ok {
		return SexpSymbol{name, symnum}
	}
	symbol := SexpSymbol{name, env.nextsymbol}
	env.symtable[name] = symbol.number
	env.revsymtable[symbol.number] = name
	env.nextsymbol++
	return symbol
}

func (env *Glisp) GenSymbol(prefix string) SexpSymbol {
	symname := prefix + strconv.Itoa(env.nextsymbol)
	return env.MakeSymbol(symname)
}

func (env *Glisp) CurrentFunctionSize() int {
	if env.curfunc.user {
		return 0
	}
	return len(env.curfunc.fun)
}

func (env *Glisp) CallFunction(function SexpFunction, nargs int) error {
	if nargs != function.nargs {
		return WrongNargs
	}
	env.addrstack.PushAddr(env.curfunc, env.pc+1)
	env.scopestack.PushScope()
	env.pc = 0
	env.curfunc = function
	return nil
}

func (env *Glisp) ReturnFromFunction() error {
	var err error
	env.curfunc, env.pc, err = env.addrstack.PopAddr()
	if err != nil {
		return err
	}
	return env.scopestack.PopScope()
}

func (env *Glisp) CallUserFunction(
	function SexpFunction, name string, nargs int) error {
	env.addrstack.PushAddr(env.curfunc, env.pc+1)
	env.scopestack.PushScope()

	env.curfunc = function
	env.pc = -1

	args, err := env.datastack.PopExpressions(nargs)
	if err != nil {
		return err
	}

	res, err := function.userfun(env, name, args)
	if err != nil {
		return err
	}
	env.datastack.PushExpr(res)

	return env.ReturnFromFunction()
}

func (env *Glisp) LoadStream(stream io.RuneReader) error {
	lexer := NewLexerFromStream(stream)

	expressions, err := ParseTokens(env, lexer)
	if err != nil {
		return errors.New(fmt.Sprintf(
			"Error on line %d: %v\n", lexer.Linenum(), err))
	}

	gen := NewGenerator(env)
	err = gen.GenerateBegin(expressions)
	if err != nil {
		return err
	}
	env.mainfunc = MakeFunction("__main", 0, gen.instructions)
	env.pc = -1
	return nil
}

func (env *Glisp) LoadFile(file *os.File) error {
	return env.LoadStream(bufio.NewReader(file))
}

func (env *Glisp) LoadString(str string) error {
	return env.LoadStream(bytes.NewBuffer([]byte(str)))
}

func (env *Glisp) AddFunction(name string, function GlispUserFunction) {
	sym := env.MakeSymbol(name)
	env.scopestack.elements[0].(Scope)[sym.number] =
		MakeUserFunction(name, function)
}

func (env *Glisp) ImportEval() {
	env.AddFunction("eval", EvalFunction)
}

func (env *Glisp) DumpEnvironment() {
	fmt.Println("Instructions:")
	if !env.curfunc.user {
		for _, instr := range env.curfunc.fun {
			fmt.Println("\t" + instr.InstrString())
		}
	}
	fmt.Println("Stack:")
	for i := 0; i <= env.datastack.tos; i++ {
		expr, _ := env.datastack.GetExpr(i)
		fmt.Println("\t" + expr.SexpString())
	}
	fmt.Printf("PC: %d\n", env.pc)

	fmt.Println("In Scope:")
	for i := 0; i <= env.scopestack.tos; i++ {
		scope := env.scopestack.elements[i].(Scope)
		for num, expr := range scope {
			name, _ := env.revsymtable[num]
			fmt.Printf("%s => %s\n", name, expr.SexpString())
		}
	}
}

func (env *Glisp) ReachedEnd() bool {
	return env.pc == env.CurrentFunctionSize()
}

func (env *Glisp) PrintStackTrace(err error) {
	fmt.Printf("error in %s:%d: %v\n",
		env.curfunc.name, env.pc, err)
	for !env.addrstack.IsEmpty() {
		fun, pos, _ := env.addrstack.PopAddr()
		fmt.Printf("in %s:%d\n", fun.name, pos)
	}
}

func (env *Glisp) Run() (Sexp, error) {
	if env.pc == -1 {
		env.pc = 0
		env.curfunc = env.mainfunc
	}

	for env.pc != -1 && !env.ReachedEnd() {
		instr := env.curfunc.fun[env.pc]
		err := instr.Execute(env)
		if err != nil {
			return SexpNull, err
		}
	}

	return env.datastack.PopExpr()
}
