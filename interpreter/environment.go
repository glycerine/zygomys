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

type PreHook func(*Glisp, string, []Sexp)
type PostHook func(*Glisp, string, Sexp)

type Glisp struct {
	datastack   *Stack
	scopestack  *Stack
	addrstack   *Stack
	stackstack  *Stack
	symtable    map[string]int
	revsymtable map[int]string
	builtins    map[int]SexpFunction
	macros      map[int]SexpFunction
	curfunc     SexpFunction
	mainfunc    SexpFunction
	pc          int
	nextsymbol  int
	before      []PreHook
	after       []PostHook
}

const CallStackSize = 25
const ScopeStackSize = 50
const DataStackSize = 100
const StackStackSize = 5

func NewGlisp() *Glisp {
	env := new(Glisp)
	env.datastack = NewStack(DataStackSize)
	env.scopestack = NewStack(ScopeStackSize)
	env.scopestack.PushScope()
	env.stackstack = NewStack(StackStackSize)
	env.addrstack = NewStack(CallStackSize)
	env.builtins = make(map[int]SexpFunction)
	env.macros = make(map[int]SexpFunction)
	env.symtable = make(map[string]int)
	env.revsymtable = make(map[int]string)
	env.nextsymbol = 1
	env.before = []PreHook{}
	env.after = []PostHook{}

	for key, function := range BuiltinFunctions {
		sym := env.MakeSymbol(key)
		env.builtins[sym.number] = MakeUserFunction(key, function)
		env.AddFunction(key, function)
	}

	env.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0), nil)
	env.curfunc = env.mainfunc
	env.pc = 0
	return env
}

func (env *Glisp) Clone() *Glisp {
	dupenv := new(Glisp)

	dupenv.datastack = env.datastack.Clone()
	dupenv.stackstack = env.stackstack.Clone()
	dupenv.scopestack = env.scopestack.Clone()
	dupenv.addrstack = env.addrstack.Clone()

	dupenv.builtins = env.builtins
	dupenv.macros = env.macros
	dupenv.symtable = env.symtable
	dupenv.revsymtable = env.revsymtable
	dupenv.nextsymbol = env.nextsymbol
	dupenv.before = env.before
	dupenv.after = env.after

	dupenv.scopestack.Push(env.scopestack.elements[0])

	dupenv.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0), nil)
	dupenv.curfunc = dupenv.mainfunc
	dupenv.pc = 0
	return dupenv
}

func (env *Glisp) Duplicate() *Glisp {
	dupenv := new(Glisp)
	dupenv.datastack = NewStack(DataStackSize)
	dupenv.scopestack = NewStack(ScopeStackSize)
	dupenv.stackstack = NewStack(StackStackSize)
	dupenv.addrstack = NewStack(CallStackSize)
	dupenv.builtins = env.builtins
	dupenv.macros = env.macros
	dupenv.symtable = env.symtable
	dupenv.revsymtable = env.revsymtable
	dupenv.nextsymbol = env.nextsymbol
	dupenv.before = env.before
	dupenv.after = env.after

	dupenv.scopestack.Push(env.scopestack.elements[0])

	dupenv.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0), nil)
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

func (env *Glisp) wrangleOptargs(fnargs, nargs int) error {
	if nargs < fnargs {
		return errors.New(
			fmt.Sprintf("Expected >%d arguments, got %d",
				fnargs, nargs))
	}
	if nargs > fnargs {
		optargs, err := env.datastack.PopExpressions(nargs - fnargs)
		if err != nil {
			return err
		}
		env.datastack.PushExpr(MakeList(optargs))
	} else {
		env.datastack.PushExpr(SexpNull)
	}
	return nil
}

func (env *Glisp) CallFunction(function SexpFunction, nargs int) error {
	for _, prehook := range env.before {
		expressions, err := env.datastack.GetExpressions(nargs)
		if err != nil {
			return err
		}
		prehook(env, function.name, expressions)
	}

	if function.varargs {
		err := env.wrangleOptargs(function.nargs, nargs)
		if err != nil {
			return err
		}
	} else if nargs != function.nargs {
		return errors.New(
			fmt.Sprintf("%s expected %d arguments, got %d",
				function.name, function.nargs, nargs))
	}

	if env.scopestack.IsEmpty() {
		panic("where's the global scope?")
	}
	globalScope := env.scopestack.elements[0]
	env.stackstack.Push(env.scopestack)
	env.scopestack = NewStack(ScopeStackSize)
	env.scopestack.Push(globalScope)

	if function.closeScope != nil {
		function.closeScope.PushAllTo(env.scopestack)
	}

	env.addrstack.PushAddr(env.curfunc, env.pc+1)
	env.scopestack.PushScope()
	env.curfunc = function
	env.pc = 0

	return nil
}

func (env *Glisp) ReturnFromFunction() error {
	for _, posthook := range env.after {
		retval, err := env.datastack.GetExpr(0)
		if err != nil {
			return err
		}
		posthook(env, env.curfunc.name, retval)
	}

	var err error
	env.curfunc, env.pc, err = env.addrstack.PopAddr()
	if err != nil {
		return err
	}
	scopestack, err := env.stackstack.Pop()
	if err != nil {
		return err
	}
	env.scopestack = scopestack.(*Stack)

	return nil
}

func (env *Glisp) CallUserFunction(
	function SexpFunction, name string, nargs int) error {

	for _, prehook := range env.before {
		expressions, err := env.datastack.GetExpressions(nargs)
		if err != nil {
			return err
		}
		prehook(env, function.name, expressions)
	}

	args, err := env.datastack.PopExpressions(nargs)
	if err != nil {
		return errors.New(
			fmt.Sprintf("Error calling %s: %v", name, err))
	}

	env.addrstack.PushAddr(env.curfunc, env.pc+1)
	env.curfunc = function
	env.pc = -1

	res, err := function.userfun(env, name, args)
	if err != nil {
		return errors.New(
			fmt.Sprintf("Error calling %s: %v", name, err))
	}
	env.datastack.PushExpr(res)

	for _, posthook := range env.after {
		posthook(env, name, res)
	}

	env.curfunc, env.pc, _ = env.addrstack.PopAddr()
	return nil
}

// SourceExpressions, this should be called from a user func context
func (env *Glisp) SourceExpressions(expressions []Sexp) error {
	gen := NewGenerator(env)
	if !env.ReachedEnd() {
		gen.AddInstruction(PopInstr(0))
	}
	err := gen.GenerateBegin(expressions)
	if err != nil {
		return err
	}

	curfunc := env.curfunc
	curpc := env.pc

	env.curfunc = MakeFunction("__source", 0, false, gen.instructions, nil)
	env.pc = 0

	env.datastack.PushExpr(SexpNull)

	if _, err = env.Run(); err != nil {
		return err
	}

	env.datastack.PopExpr()

	env.pc = curpc
	env.curfunc = curfunc

	return nil
}

func (env *Glisp) SourceStream(stream io.RuneReader) error {
	lexer := NewLexerFromStream(stream)

	expressions, err := ParseTokens(env, lexer)
	if err != nil {
		return errors.New(fmt.Sprintf(
			"Error on line %d: %v\n", lexer.Linenum(), err))
	}

	return env.SourceExpressions(expressions)
}

func (env *Glisp) SourceFile(file *os.File) error {
	return env.SourceStream(bufio.NewReader(file))
}

func (env *Glisp) LoadExpressions(expressions []Sexp) error {
	gen := NewGenerator(env)
	if !env.ReachedEnd() {
		gen.AddInstruction(PopInstr(0))
	}
	err := gen.GenerateBegin(expressions)
	if err != nil {
		return err
	}

	env.mainfunc.fun = append(env.mainfunc.fun, gen.instructions...)
	env.curfunc = env.mainfunc

	return nil
}

func (env *Glisp) ParseFile(file string) ([]Sexp, error) {
	in, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	lexer := NewLexerFromStream(bufio.NewReader(in))

	var exp []Sexp

	exp, err = ParseTokens(env, lexer)
	if err != nil {
		return nil, fmt.Errorf("Error on line %d: %v\n", lexer.Linenum(), err)
	}

	in.Close()

	return exp, nil
}

func (env *Glisp) LoadStream(stream io.RuneReader) error {
	lexer := NewLexerFromStream(stream)

	expressions, err := ParseTokens(env, lexer)
	if err != nil {
		return fmt.Errorf("Error on line %d: %v\n", lexer.Linenum(), err)
	}

	return env.LoadExpressions(expressions)
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
	env.AddGlobal(name, MakeUserFunction(name, function))
}

func (env *Glisp) AddGlobal(name string, obj Sexp) {
	sym := env.MakeSymbol(name)
	env.scopestack.elements[0].(Scope)[sym.number] = obj
}

func (env *Glisp) AddMacro(name string, function GlispUserFunction) {
	sym := env.MakeSymbol(name)
	env.macros[sym.number] = MakeUserFunction(name, function)
}

func (env *Glisp) HasMacro(sym SexpSymbol) bool {
	_, found := env.macros[sym.number]
	return found
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
	DumpFunction(fun)
	return nil
}

func DumpFunction(fun GlispFunction) {
	for _, instr := range fun {
		fmt.Println("\t" + instr.InstrString())
	}
}

func (env *Glisp) DumpEnvironment() {
	fmt.Println("Instructions:")
	if !env.curfunc.user {
		DumpFunction(env.curfunc.fun)
	}
	fmt.Println("Stack:")
	env.datastack.PrintStack()
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
	env.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0), nil)
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

	env.pc = -2
	for _, expr := range args {
		env.datastack.PushExpr(expr)
	}

	//log.Print("Apply Calling ", fun, " with ", len(args))
	err := env.CallFunction(fun, len(args))
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

func (env *Glisp) AddPreHook(fun PreHook) {
	env.before = append(env.before, fun)
}

func (env *Glisp) AddPostHook(fun PostHook) {
	env.after = append(env.after, fun)
}
