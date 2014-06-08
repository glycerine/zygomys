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
	builtins    map[int]SexpFunction
	macros      map[int]SexpFunction
	curfunc     SexpFunction
	mainfunc    SexpFunction
	pc          int
	nextsymbol  int
}

const CallStackSize = 25
const ScopeStackSize = 50
const DataStackSize = 100

func NewGlisp() *Glisp {
	env := new(Glisp)
	env.datastack = NewStack(DataStackSize)
	env.scopestack = NewStack(ScopeStackSize)
	env.scopestack.PushScope()
	env.addrstack = NewStack(CallStackSize)
	env.builtins = make(map[int]SexpFunction)
	env.macros = make(map[int]SexpFunction)
	env.symtable = make(map[string]int)
	env.revsymtable = make(map[int]string)
	env.nextsymbol = 1

	for key, function := range BuiltinFunctions {
		sym := env.MakeSymbol(key)
		env.builtins[sym.number] = MakeUserFunction(key, function)
		env.AddFunction(key, function)
	}

	env.mainfunc = MakeFunction("__main", 0, make([]Instruction, 0))
	env.curfunc = env.mainfunc
	env.pc = 0
	return env
}

func (env *Glisp) Duplicate() *Glisp {
	dupenv := new(Glisp)
	dupenv.datastack = NewStack(DataStackSize)
	dupenv.scopestack = NewStack(ScopeStackSize)
	dupenv.addrstack = NewStack(CallStackSize)
	dupenv.builtins = env.builtins
	dupenv.macros = env.macros
	dupenv.symtable = env.symtable
	dupenv.revsymtable = env.revsymtable
	dupenv.nextsymbol = env.nextsymbol

	dupenv.scopestack.Push(env.scopestack.elements[0])

	env.mainfunc = MakeFunction("__main", 0, make([]Instruction, 0))
	dupenv.curfunc = dupenv.mainfunc
	dupenv.pc = 0
	return dupenv
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

	args, err := env.datastack.PopExpressions(nargs)
	if err != nil {
		return err
	}

	env.addrstack.PushAddr(env.curfunc, env.pc+1)

	env.curfunc = function
	env.pc = -1

	res, err := function.userfun(env, name, args)
	if err != nil {
		return err
	}
	env.datastack.PushExpr(res)

	env.curfunc, env.pc, _ = env.addrstack.PopAddr()
	return nil
}

func (env *Glisp) LoadStream(stream io.RuneReader) error {
	lexer := NewLexerFromStream(stream)

	expressions, err := ParseTokens(env, lexer)
	if err != nil {
		return errors.New(fmt.Sprintf(
			"Error on line %d: %v\n", lexer.Linenum(), err))
	}

	gen := NewGenerator(env)
	if !env.ReachedEnd() {
		gen.AddInstruction(PopInstr(0))
	}
	err = gen.GenerateBegin(expressions)
	if err != nil {
		return err
	}

	env.mainfunc.fun = append(env.mainfunc.fun, gen.instructions...)
	env.curfunc = env.mainfunc

	return nil
}

func (env *Glisp) EvalString(str string) (Sexp, error) {
	err := env.LoadString(str)
	if err != nil {
		return SexpNull, err
	}

	return env.Run()
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

func (env *Glisp) AddMacro(name string, function GlispUserFunction) {
	sym := env.MakeSymbol(name)
	env.macros[sym.number] = MakeUserFunction(name, function)
}

func (env *Glisp) ImportEval() {
	env.AddFunction("eval", EvalFunction)
}

func (env *Glisp) DumpFunctionByName(name string) error {
	obj, found := env.FindObject(name)
	if !found {
		return errors.New(fmt.Sprintf("%q not found", name))
	}

	var fun GlispFunction
	switch t := obj.(type) {
	case SexpFunction:
		if !t.user {
			fun = t.fun
		} else {
			return errors.New("not a glisp function")
		}
	default:
		return errors.New("not a function")
	}
	env.DumpFunction(fun)
	return nil
}

func (env *Glisp) DumpFunction(fun GlispFunction) {
	for _, instr := range fun {
		fmt.Println("\t" + instr.InstrString())
	}
}

func (env *Glisp) DumpEnvironment() {
	fmt.Println("Instructions:")
	if !env.curfunc.user {
		env.DumpFunction(env.curfunc.fun)
	}
	fmt.Println("Stack:")
	for i := 0; i <= env.datastack.tos; i++ {
		expr, _ := env.datastack.GetExpr(i)
		fmt.Println("\t" + expr.SexpString())
	}
	fmt.Printf("PC: %d\n", env.pc)
}

func (env *Glisp) ReachedEnd() bool {
	return env.pc == env.CurrentFunctionSize()
}

func (env *Glisp) GetStackTrace(err error) string {
	str := fmt.Sprintf("error in %s:%d: %v\n",
		env.curfunc.name, env.pc, err)
	for !env.addrstack.IsEmpty() {
		fun, pos, _ := env.addrstack.PopAddr()
		str += fmt.Sprintf("in %s:%d\n", fun.name, pos)
	}
	return str
}

func (env *Glisp) Clear() {
	env.datastack.tos = -1
	env.scopestack.tos = 0
	env.addrstack.tos = -1
	env.mainfunc = MakeFunction("__main", 0, make([]Instruction, 0))
	env.curfunc = env.mainfunc
	env.pc = 0
}

func (env *Glisp) FindObject(name string) (Sexp, bool) {
	sym := env.MakeSymbol(name)
	obj, err := env.scopestack.LookupSymbol(sym)
	if err != nil {
		return SexpNull, false
	}
	return obj, true
}

func (env *Glisp) Apply(fun SexpFunction, args []Sexp) (Sexp, error) {
	if fun.user {
		return fun.userfun(env, fun.name, args)
	}
	if len(args) != fun.nargs {
		return SexpNull, WrongNargs
	}
	env.pc = -2
	for _, expr := range args {
		env.datastack.PushExpr(expr)
	}
	err := env.CallFunction(fun, fun.nargs)
	if err != nil {
		return SexpNull, err
	}
	return env.Run()
}

func (env *Glisp) Run() (Sexp, error) {
	for env.pc != -1 && !env.ReachedEnd() {
		instr := env.curfunc.fun[env.pc]
		err := instr.Execute(env)
		if err != nil {
			return SexpNull, err
		}
	}

	return env.datastack.PopExpr()
}
