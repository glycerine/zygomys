package zygo

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
	datastack  *Stack
	scopestack *Stack
	addrstack  *Stack
	stackstack *Stack

	// linearstack: push on scope enter, pop on scope exit. runtime dynamic.
	linearstack *Stack

	// loopstack: let break and continue find the nearest enclosing loop.
	loopstack *Stack

	// lexicalstack: track the scope where a function was defined.
	lexicalstack *Stack

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

	debugExec           bool
	debugSymbolNotFound bool
	oldStyleLookups     bool
	showGlobalScope     bool
}

const CallStackSize = 25
const ScopeStackSize = 50
const DataStackSize = 100
const StackStackSize = 5
const LoopStackSize = 5

func NewGlisp() *Glisp {
	env := new(Glisp)
	env.datastack = env.NewStack(DataStackSize)
	env.scopestack = env.NewStack(ScopeStackSize)
	env.linearstack = env.NewStack(ScopeStackSize)
	env.lexicalstack = env.NewStack(ScopeStackSize)
	glob := NewScope()
	glob.IsGlobal = true
	env.scopestack.Push(glob)
	env.linearstack.Push(glob)
	env.lexicalstack.Push(glob)
	env.stackstack = env.NewStack(StackStackSize)
	env.addrstack = env.NewStack(CallStackSize)
	env.loopstack = env.NewStack(LoopStackSize)
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
	env.debugSymbolNotFound = false
	return env
}

func (env *Glisp) Clone() *Glisp {
	dupenv := new(Glisp)

	dupenv.datastack = env.datastack.Clone()
	dupenv.stackstack = env.stackstack.Clone()
	dupenv.scopestack = env.scopestack.Clone()
	dupenv.linearstack = env.linearstack.Clone()
	dupenv.lexicalstack = env.lexicalstack.Clone()
	dupenv.addrstack = env.addrstack.Clone()

	dupenv.builtins = env.builtins
	dupenv.macros = env.macros
	dupenv.symtable = env.symtable
	dupenv.revsymtable = env.revsymtable
	dupenv.nextsymbol = env.nextsymbol
	dupenv.before = env.before
	dupenv.after = env.after

	dupenv.scopestack.Push(env.scopestack.elements[0])
	dupenv.linearstack.Push(env.scopestack.elements[0])
	dupenv.lexicalstack.Push(env.scopestack.elements[0])

	dupenv.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0), nil)
	dupenv.curfunc = dupenv.mainfunc
	dupenv.pc = 0
	dupenv.debugExec = env.debugExec
	dupenv.debugSymbolNotFound = env.debugSymbolNotFound
	dupenv.oldStyleLookups = env.oldStyleLookups
	dupenv.showGlobalScope = env.showGlobalScope

	return dupenv
}

func (env *Glisp) Duplicate() *Glisp {
	dupenv := new(Glisp)
	dupenv.datastack = dupenv.NewStack(DataStackSize)
	dupenv.scopestack = dupenv.NewStack(ScopeStackSize)
	dupenv.linearstack = dupenv.NewStack(ScopeStackSize)
	dupenv.lexicalstack = dupenv.NewStack(ScopeStackSize)
	dupenv.stackstack = dupenv.NewStack(StackStackSize)
	dupenv.addrstack = dupenv.NewStack(CallStackSize)
	dupenv.builtins = env.builtins
	dupenv.macros = env.macros
	dupenv.symtable = env.symtable
	dupenv.revsymtable = env.revsymtable
	dupenv.nextsymbol = env.nextsymbol
	dupenv.before = env.before
	dupenv.after = env.after

	dupenv.scopestack.Push(env.scopestack.elements[0])
	dupenv.linearstack.Push(env.scopestack.elements[0])
	dupenv.lexicalstack.Push(env.scopestack.elements[0])

	dupenv.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0), nil)
	dupenv.curfunc = dupenv.mainfunc
	dupenv.pc = 0
	dupenv.debugExec = env.debugExec
	dupenv.debugSymbolNotFound = env.debugSymbolNotFound
	dupenv.oldStyleLookups = env.oldStyleLookups
	dupenv.showGlobalScope = env.showGlobalScope

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

	env.linearstack.PushScope()
	globalScope := env.scopestack.elements[0]
	env.stackstack.Push(env.scopestack)
	env.scopestack = env.NewStack(ScopeStackSize)
	env.scopestack.Push(globalScope)

	if function.closeScope != nil {
		function.closeScope.PushAllTo(env.scopestack)
		function.closeScope.PushAllTo(env.linearstack)
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
	/* old style:
	scopestack, err := env.stackstack.Pop()
	if err != nil {
		return err
	}
	env.scopestack = scopestack.(*Stack)
	*/
	// new style
	_, err = env.stackstack.Pop()
	if err != nil {
		return err
	}
	_, err = env.linearstack.Pop()

	return err
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
	env.scopestack.elements[0].(*Scope).Map[sym.number] = obj
	env.linearstack.elements[0].(*Scope).Map[sym.number] = obj
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
		return errors.New("dump by name error: not a function")
	}
	DumpFunction(fun, -1)
	return nil
}

// if pc is -1, don't show it.
func DumpFunction(fun GlispFunction, pc int) {
	blank := "      "
	extra := blank
	for i, instr := range fun {
		if i == pc {
			extra = " PC-> "
		} else {
			extra = blank
		}
		fmt.Printf("%s %d: %s\n", extra, i, instr.InstrString())
	}
	if pc == len(fun) {
		fmt.Printf(" PC just past end at %d -----\n\n", pc)
	}
}

func (env *Glisp) DumpEnvironment() {
	fmt.Printf("PC: %d\n", env.pc)
	fmt.Println("Instructions:")
	if !env.curfunc.user {
		DumpFunction(env.curfunc.fun, env.pc)
	}
	fmt.Printf("Stack: (length %d)\n", env.datastack.tos+1)
	env.datastack.PrintStack()
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
	env.linearstack.tos = 0
	env.lexicalstack.tos = 0
	env.addrstack.tos = -1
	env.mainfunc = MakeFunction("__main", 0, false, make([]Instruction, 0), nil)
	env.curfunc = env.mainfunc
	env.pc = 0
}

func (env *Glisp) FindObject(name string) (Sexp, bool) {
	sym := env.MakeSymbol(name)
	if env.oldStyleLookups {
		obj, err, _ := env.scopestack.LookupSymbol(sym)
		if err != nil {
			return SexpNull, false
		}
		return obj, true
	} else {
		obj, err, _ := env.linearstack.LookupSymbol(sym)
		if err != nil {
			return SexpNull, false
		}
		return obj, true
	}
}

func (env *Glisp) Apply(fun SexpFunction, args []Sexp) (Sexp, error) {
	VPrintf("\n\n debug Apply not working on user funcs: fun = '%#v'   and args = '%#v'\n\n", fun, args)
	if fun.user {
		return fun.userfun(env, fun.name, args)
	}

	env.pc = -2
	for _, expr := range args {
		env.datastack.PushExpr(expr)
	}

	//fmt.Printf("\nApply Calling '%s'\n", fun.SexpString())
	err := env.CallFunction(fun, len(args))
	if err != nil {
		return SexpNull, err
	}

	return env.Run()
}

func (env *Glisp) Run() (Sexp, error) {
	for env.pc != -1 && !env.ReachedEnd() {
		instr := env.curfunc.fun[env.pc]
		if env.debugExec {
			fmt.Printf("\n ====== in '%s', about to run instr: %#v\n",
				env.curfunc.name, instr)
			env.DumpEnvironment()
			fmt.Printf("\n ====== in '%s', now running %#v\n",
				env.curfunc.name, instr)
		}
		err := instr.Execute(env)
		if err != nil {
			return SexpNull, err
		}
		if env.debugExec {
			fmt.Printf("\n ****** in '%s', after running '%#v', stack is: \n",
				env.curfunc.name, instr)
			env.DumpEnvironment()
			fmt.Printf("\n ****** \n")

		}
	}

	if env.datastack.IsEmpty() {
		fmt.Printf("env.datastack was empty!!\n")
		env.DumpEnvironment()
		panic("env.datastack was empty!!")
		//os.Exit(-1)
	}

	return env.datastack.PopExpr()
}

func (env *Glisp) AddPreHook(fun PreHook) {
	env.before = append(env.before, fun)
}

func (env *Glisp) AddPostHook(fun PostHook) {
	env.after = append(env.after, fun)
}

// scan the instruction stream to locate loop start
func (env *Glisp) FindLoop(target *Loop) (int, error) {
	if env.curfunc.user {
		panic(fmt.Errorf("impossible in user-defined-function to find a loop '%s'", target.stmtname.name))
	}
	instruc := env.curfunc.fun
	for i := range instruc {
		switch loop := instruc[i].(type) {
		case LoopStartInstr:
			if loop.loop == target {
				return i, nil
			}
		}
	}
	return -1, fmt.Errorf("could not find loop target '%s'", target.stmtname.name)
}

// startat = 0 to show everything starting with global bindings
// startat = 1 to show one scope inside the global.
// startat = 2, 3, ... show deeper scopes only.
func (env *Glisp) ShowScopes(startat int) error {
	top := env.scopestack.Top()
	VPrintf("\n ShowScopes has top=%d\n", top)
	for i := startat; i <= top; i++ {
		scop, err := env.scopestack.Get(i)
		if err != nil {
			return err
		}

		switch sc := scop.(type) {
		case Scope:
			sc.Show(env, i*6, fmt.Sprintf("====  scope # %d  in  env %p\n", i, env))
		case *Scope:
			sc.Show(env, i*6, fmt.Sprintf("====  scope # %d  in  env %p\n", i, env))
		default:
			return fmt.Errorf("unexpected type %T / val = %#v on scopestack", sc, sc)
		}
	}
	return nil
}

func (env *Glisp) showStackHelper(stack *Stack, name string) {
	note := ""
	n := stack.Top()
	if n < 0 {
		note = "(empty)"
	}
	fmt.Printf(" ========  env.%s is %v deep: %s\n", name, n+1, note)
	for i := 0; i <= n; i++ {
		ele, err := stack.Get(n - i)
		if err != nil {
			panic(fmt.Errorf("env.%s access error on %i: %v",
				name, i, err))
		}
		label := fmt.Sprintf("%s %v", name, i)
		switch x := ele.(type) {
		case *Stack:
			x.Show(env, 0, label)
		case *Scope:
			x.Show(env, 0, label)
		case Scope:
			x.Show(env, 0, label)
		default:
			panic(fmt.Errorf("unrecognized element on %s: %T/val=%v",
				name, x, x))
		}
	}
}

/*
func (env *Glisp) OLD_ShowStackStackAndScopeStack() error {
	note := ""
	n := env.stackstack.Top()
	if n < 0 {
		note = "(empty)"
	}
	fmt.Printf(" ========  env.stackstack is %v deep: %s\n", n+1, note)
	for i := 0; i <= n; i++ {
		ele, err := env.stackstack.Get(n - i)
		if err != nil {
			panic(fmt.Errorf("env.stackstack access error on %i: %v", i, err))
		}
		label := fmt.Sprintf("stackstack %v", i)
		switch x := ele.(type) {
		case *Stack:
			x.Show(env, 0, label)
		default:
			panic(fmt.Errorf("unrecognized element on stackstack: %T/val=%v", x, x))
		}
	}
	n = env.scopestack.Top()
	fmt.Printf(" ++++++++  env.scopestack is %v deep:\n", n+1)
	for i := 0; i <= n; i++ {
		ele, err := env.scopestack.Get(n - i)
		if err != nil {
			panic(fmt.Errorf("env.scopestack access error on %i: %v", i, err))
		}
		label := fmt.Sprintf("scopestack %v", i)
		switch x := ele.(type) {
		case *Scope:
			x.Show(env, 0, label)
		case Scope:
			x.Show(env, 0, label)
		default:
			panic(fmt.Errorf("unrecognized element on scopestack: %T/val=%v", x, x))
		}
	}

	n = env.linearstack.Top()
	fmt.Printf(" ++++++++  env.linearstack is %v deep:\n", n+1)
	for i := 0; i <= n; i++ {
		ele, err := env.linearstack.Get(n - i)
		if err != nil {
			panic(fmt.Errorf("env.linearstack access error on %i: %v", i, err))
		}
		label := fmt.Sprintf("linearstack %v", i)
		switch x := ele.(type) {
		case *Scope:
			x.Show(env, 0, label)
		case Scope:
			x.Show(env, 0, label)
		default:
			panic(fmt.Errorf("unrecognized element on linearstack: %T/val=%v", x, x))
		}
	}

	fmt.Printf(" --------\n")
	return nil
}
*/

func (env *Glisp) ShowStackStackAndScopeStack() error {
	env.showStackHelper(env.stackstack, "stackstack")
	env.showStackHelper(env.scopestack, "scopestack")
	env.showStackHelper(env.linearstack, "linearstack")
	env.showStackHelper(env.lexicalstack, "lexicalstack")
	return nil
}

/*
zygo> (defn h [] (println "hi from h") (defn f [] (defn g [] (println "hi from g") (.ls)) (g)) (f))
zygo> (h)
hi from h
hi from g
 ========  env.stackstack is 3 deep:
 stackstack 0
     elem 0 of stackstack 0:
         (global scope - omitting content for brevity)
 stackstack 1
     elem 0 of stackstack 1:
         (global scope - omitting content for brevity)
     elem 1 of stackstack 1:
         f -> (defn f [] (defn g [] (println "hi from g") (.ls)) (g))
 stackstack 2
     elem 0 of stackstack 2:
         (global scope - omitting content for brevity)
     elem 1 of stackstack 2:
         g -> (defn g [] (println "hi from g") (.ls))
 ++++++++  env.scopestack is 2 deep:
 scopestack 0
     (global scope - omitting content for brevity)
 scopestack 1
     empty-scope: no symbols
 --------
zygo>
*/
