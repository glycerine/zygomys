package glisp

import (
	"os"
	"errors"
	"io"
	"bufio"
	"fmt"
)

type Glisp struct {
	datastack *Stack
	scopestack *Stack
	addrstack *Stack
	symtable map[string]int
	user_functions map[int]GlispUserFunction
	curfunc GlispFunction
	mainfunc GlispFunction
	pc int
	nextsymbol int
}

func NewGlisp() *Glisp {
	env := new(Glisp)
	env.datastack = NewStack()
	env.scopestack = NewStack()
	env.scopestack.PushScope()
	env.addrstack = NewStack()
	env.symtable = make(map[string]int)
	env.nextsymbol = 1

	env.user_functions = make(map[int]GlispUserFunction)
	for key, function := range BuiltinFunctions {
		sym := env.MakeSymbol(key)
		env.user_functions[sym.number] = function
	}

	env.curfunc = nil
	env.mainfunc = make([]Instruction, 0)
	env.pc = -1
	return env
}

func (env *Glisp) MakeSymbol(name string) SexpSymbol {
	symnum, ok := env.symtable[name]
	if ok {
		return SexpSymbol{name, symnum}
	}
	symbol := SexpSymbol{name, env.nextsymbol}
	env.symtable[name] = env.nextsymbol
	env.nextsymbol++
	return symbol
}

func (env *Glisp) CurrentFunctionSize() int {
	if env.curfunc == nil {
		return 0
	}
	return len(env.curfunc)
}

func (env *Glisp) CallFunction(function GlispFunction) {
	env.addrstack.PushAddr(env.curfunc, env.pc + 1)
	env.scopestack.PushScope()
	env.pc = 0
	env.curfunc = function
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
		function GlispUserFunction, name string, nargs int) error {
	env.addrstack.PushAddr(env.curfunc, env.pc + 1)
	env.scopestack.PushScope()

	env.curfunc = nil
	env.pc = -1

	args, err := env.datastack.PopExpressions(nargs)
	if err != nil {
		return err
	}

	res, err := function(env, name, args)
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
	env.mainfunc = append(env.mainfunc, gen.instructions...)
	return nil
}

func (env *Glisp) LoadFile(file *os.File) error {
	return env.LoadStream(bufio.NewReader(file))
}

func (env *Glisp) DumpEnvironment() {
	fmt.Println("instructions:")
	for _, instr := range env.mainfunc {
		fmt.Println("\t" + instr.InstrString())
	}
	fmt.Println("stack:")
	for !env.datastack.IsEmpty() {
		expr, _ := env.datastack.PopExpr()
		fmt.Println("\t" + expr.SexpString())
	}
}

func (env *Glisp) ReachedEnd() bool {
	return env.pc == env.CurrentFunctionSize()
}

func (env *Glisp) Run() error {
	if env.pc == -1 {
		env.pc = 0
		env.curfunc = env.mainfunc
	}

	for env.pc != -1 && !env.ReachedEnd() {
		instr := env.curfunc[env.pc]
		err := instr.Execute(env)
		if err != nil {
			return err
		}
	}

	return nil
}

func (env *Glisp) PopResult() (Sexp, error) {
	return env.datastack.PopExpr()
}
