package zygo

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
)

type PreHook func(*Zlisp, string, []Sexp)
type PostHook func(*Zlisp, string, Sexp)

type Zlisp struct {
	parser    ParserI
	datastack *Stack
	addrstack *Stack

	// linearstack: push on scope enter, pop on scope exit. runtime dynamic.
	linearstack *Stack

	// loopstack: let break and continue find the nearest enclosing loop.
	loopstack *Stack

	symtable    map[string]int
	revsymtable map[int]string
	builtins    map[int]*SexpFunction
	reserved    map[int]bool
	macros      map[int]*SexpFunction
	curfunc     *SexpFunction
	mainfunc    *SexpFunction
	pc          int
	nextsymbol  int
	before      []PreHook
	after       []PostHook

	debugExec           bool
	debugSymbolNotFound bool

	showGlobalScope bool
	baseTypeCtor    *SexpFunction

	infixOps map[string]*InfixOp
	Pretty   bool

	booter Booter

	// API use, since infix is already default at repl
	WrapLoadExpressionsInInfix bool
}

// allow clients to establish a callback to
// happen after reinflating a Go struct. These
// structs need to be "booted" to be ready to go.
func (env *Zlisp) SetBooter(b Booter) {
	env.booter = b
}

// Booter provides for registering a callback
// for any new Go struct created by the ToGoFunction (togo).
type Booter func(s interface{})

const CallStackSize = 25
const ScopeStackSize = 50
const DataStackSize = 100
const StackStackSize = 5
const LoopStackSize = 5

var ReservedWords = []string{"byte", "defbuild", "builder", "field", "and", "or", "cond", "quote", "def", "mdef", "fn", "defn", "begin", "let", "letseq", "assert", "defmac", "macexpand", "syntaxQuote", "include", "for", "set", "break", "continue", "newScope", "_ls", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "complex64", "complex128", "bool", "string", "any", "break", "case", "chan", "const", "continue", "default", "else", "defer", "fallthrough", "for", "func", "go", "goto", "if", "import", "interface", "map", "package", "range", "return", "select", "struct", "switch", "type", "var", "append", "cap", "close", "complex", "copy", "delete", "imag", "len", "make", "new", "panic", "print", "println", "real", "recover", "null", "nil", "-", "+", "--", "++", "-=", "+=", ":=", "=", ">", "<", ">=", "<=", "send", "NaN", "nan"}

func NewZlisp() *Zlisp {
	return NewZlispWithFuncs(AllBuiltinFunctions())
}

// Close cleans up the allocated env resources;
// it stops the parser goroutine
// ands frees it. Close should
// be called when you are done using the env
// to avoid having the parser goroutine hang around
// until process end.
func (env *Zlisp) Close() error {
	return env.parser.Stop()
}

// NewZlispSandbox returns a new *Zlisp instance that does not allow the
// user to get to the outside world
func NewZlispSandbox() *Zlisp {
	return NewZlispWithFuncs(SandboxSafeFunctions())
}

// NewZlispWithFuncs returns a new *Zlisp instance with access to only the given builtin functions
func NewZlispWithFuncs(funcs map[string]ZlispUserFunction) *Zlisp {
	env := new(Zlisp)
	env.baseTypeCtor = MakeUserFunction("__basetype_ctor", BaseTypeConstructorFunction)
	env.parser = env.NewLazyParser()
	env.parser.Start()
	env.datastack = env.NewStack(DataStackSize)
	env.linearstack = env.NewStack(ScopeStackSize)

	glob := env.NewNamedScope("global")
	glob.IsGlobal = true
	env.linearstack.Push(glob)
	env.addrstack = env.NewStack(CallStackSize)
	env.loopstack = env.NewStack(LoopStackSize)
	env.builtins = make(map[int]*SexpFunction)
	env.reserved = make(map[int]bool)
	env.macros = make(map[int]*SexpFunction)
	env.symtable = make(map[string]int)
	env.revsymtable = make(map[int]string)
	env.nextsymbol = 1
	env.before = []PreHook{}
	env.after = []PostHook{}
	env.infixOps = make(map[string]*InfixOp)
	env.AddGlobal("null", SexpNull)
	env.AddGlobal("nil", SexpNull)

	for key, function := range funcs {
		sym := env.MakeSymbol(key)
		env.builtins[sym.number] = MakeUserFunction(key, function)
		env.AddFunction(key, function)
	}

	for _, word := range ReservedWords {
		sym := env.MakeSymbol(word)
		env.reserved[sym.number] = true
	}

	env.mainfunc = env.MakeFunction("__main", 0, false,
		make([]Instruction, 0), nil)
	env.curfunc = env.mainfunc
	env.pc = 0
	env.debugSymbolNotFound = false
	//env.debugSymbolNotFound = true
	//env.debugExec = true
	env.InitInfixOps()

	return env

}

func (env *Zlisp) Clone() *Zlisp {
	dupenv := new(Zlisp)
	dupenv.parser = env.parser
	dupenv.baseTypeCtor = env.baseTypeCtor
	dupenv.datastack = env.datastack.Clone()
	dupenv.linearstack = env.linearstack.Clone()
	dupenv.addrstack = env.addrstack.Clone()

	dupenv.builtins = env.builtins
	dupenv.reserved = env.reserved
	dupenv.macros = env.macros
	dupenv.symtable = env.symtable
	dupenv.revsymtable = env.revsymtable
	dupenv.nextsymbol = env.nextsymbol
	dupenv.before = env.before
	dupenv.after = env.after
	dupenv.infixOps = env.infixOps
	dupenv.linearstack.Push(env.linearstack.elements[0])

	dupenv.mainfunc = env.MakeFunction("__main", 0, false,
		make([]Instruction, 0), nil)
	dupenv.curfunc = dupenv.mainfunc
	dupenv.pc = 0
	dupenv.debugExec = env.debugExec
	dupenv.debugSymbolNotFound = env.debugSymbolNotFound
	dupenv.showGlobalScope = env.showGlobalScope
	return dupenv
}

func (env *Zlisp) Duplicate() *Zlisp {
	dupenv := new(Zlisp)
	dupenv.parser = env.parser
	dupenv.baseTypeCtor = env.baseTypeCtor
	dupenv.datastack = dupenv.NewStack(DataStackSize)
	dupenv.linearstack = dupenv.NewStack(ScopeStackSize)
	dupenv.addrstack = dupenv.NewStack(CallStackSize)
	dupenv.builtins = env.builtins
	dupenv.reserved = env.reserved
	dupenv.macros = env.macros
	dupenv.symtable = env.symtable
	dupenv.revsymtable = env.revsymtable
	dupenv.nextsymbol = env.nextsymbol
	dupenv.before = env.before
	dupenv.after = env.after
	dupenv.infixOps = env.infixOps

	dupenv.linearstack.Push(env.linearstack.elements[0])

	dupenv.mainfunc = env.MakeFunction("__main", 0, false,
		make([]Instruction, 0), nil)
	dupenv.curfunc = dupenv.mainfunc
	dupenv.pc = 0
	dupenv.debugExec = env.debugExec
	dupenv.debugSymbolNotFound = env.debugSymbolNotFound
	dupenv.showGlobalScope = env.showGlobalScope

	return dupenv
}

func (env *Zlisp) MakeDotSymbol(name string) *SexpSymbol {
	x := env.MakeSymbol(name)
	x.isDot = true
	return x
}

func (env *Zlisp) DetectSigils(sym *SexpSymbol) {
	if sym == nil {
		return
	}
	if len(sym.name) == 0 {
		return
	}
	switch sym.name[0] {
	case '$':
		sym.isSigil = true
		sym.sigil = "$"
	case '#':
		sym.isSigil = true
		sym.sigil = "#"
	case '?':
		sym.isSigil = true
		sym.sigil = "?"
	}
}

func (env *Zlisp) DumpSymTable() {
	for kk, vv := range env.symtable {
		fmt.Printf("symtable entry: kk: '%v' -> '%v'\n", kk, vv)
	}
}
func (env *Zlisp) MakeSymbol(name string) *SexpSymbol {
	if env == nil {
		panic("internal problem:  env.MakeSymbol called with nil env")
	}
	symnum, ok := env.symtable[name]
	if ok {
		symbol := &SexpSymbol{name: name, number: symnum}
		env.DetectSigils(symbol)
		return symbol
	}
	symbol := &SexpSymbol{name: name, number: env.nextsymbol}
	env.symtable[name] = symbol.number
	env.revsymtable[symbol.number] = name

	env.nextsymbol++
	env.DetectSigils(symbol)
	return symbol
}

func (env *Zlisp) GenSymbol(prefix string) *SexpSymbol {
	symname := prefix + strconv.Itoa(env.nextsymbol)
	return env.MakeSymbol(symname)
}

func (env *Zlisp) CurrentFunctionSize() int {
	if env.curfunc.user {
		return 0
	}
	return len(env.curfunc.fun)
}

func (env *Zlisp) wrangleOptargs(fnargs, nargs int) error {
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

func (env *Zlisp) CallFunction(function *SexpFunction, nargs int) error {
	for _, prehook := range env.before {
		expressions, err := env.datastack.GetExpressions(nargs)
		if err != nil {
			return err
		}
		prehook(env, function.name, expressions)
	}

	// do name and type checking
	err := env.FunctionCallNameTypeCheck(function, &nargs)
	if err != nil {
		return err
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

	if env.linearstack.IsEmpty() {
		panic("where's the global scope?")
	}

	env.addrstack.PushAddr(env.curfunc, env.pc+1)

	//P("DEBUG linearstack with this next:")
	//env.showStackHelper(env.linearstack, "linearstack")

	// this effectely *is* the call, because it sets the
	// next instructions to happen once we exit.
	env.curfunc = function
	env.pc = 0

	//Q("\n CallFunction starting with stack:\n")
	//env.ShowStackStackAndScopeStack()

	return nil
}

func (env *Zlisp) ReturnFromFunction() error {
	for _, posthook := range env.after {
		retval, err := env.datastack.GetExpr(0)
		if err != nil {
			return err
		}
		posthook(env, env.curfunc.name, retval)
	}
	var err error
	env.curfunc, env.pc, err = env.addrstack.PopAddr()
	return err
}

func (env *Zlisp) CallUserFunction(
	function *SexpFunction, name string, nargs int) (nargReturned int, err error) {
	//Q("CallUserFunction calling name '%s' with nargs=%v", name, nargs)
	for _, prehook := range env.before {
		expressions, err := env.datastack.GetExpressions(nargs)
		if err != nil {
			return 0, err
		}
		prehook(env, function.name, expressions)
	}

	args, err := env.datastack.PopExpressions(nargs)
	if err != nil {
		return 0, errors.New(
			fmt.Sprintf("Error calling '%s': %v", name, err))
	}

	env.addrstack.PushAddr(env.curfunc, env.pc+1)
	env.curfunc = function
	env.pc = -1

	//P("DEBUG linearstack with this next, just before calling function.userfun:")
	//env.showStackHelper(env.linearstack, "linearstack")

	// protect against bad calls/bad reflection in usercalls
	var wasPanic bool
	var recovered interface{}
	var trace []byte
	res, err := func() (Sexp, error) {
		defer func() {
			recovered = recover()
			if recovered != nil {
				wasPanic = true
				trace = make([]byte, 16384)
				nbyte := runtime.Stack(trace, false)
				trace = trace[:nbyte]
			}
		}()

		// the point we were getting to, before the panic protection:
		return function.userfun(env, name, args)
	}()

	//P("DEBUG linearstack with this next, just *after* calling function.userfun:")
	//env.showStackHelper(env.linearstack, "linearstack")

	if wasPanic {
		err = fmt.Errorf("CallUserFunction caught panic during call of "+
			"'%s': '%v'\n stack trace:\n%v\n",
			name, recovered, string(trace))
	}
	if err != nil {
		return 0, errors.New(
			fmt.Sprintf("Error calling '%s': %v", name, err))
	}

	env.datastack.PushExpr(res)

	for _, posthook := range env.after {
		posthook(env, name, res)
	}

	env.curfunc, env.pc, _ = env.addrstack.PopAddr()
	return len(args), nil
}

func (env *Zlisp) LoadExpressions(xs []Sexp) error {

	expressions := xs
	if env.WrapLoadExpressionsInInfix {
		infixSym := env.MakeSymbol("infix")
		expressions = []Sexp{MakeList([]Sexp{infixSym, &SexpArray{Val: xs, Env: env}})}
	}

	//P("expressions before RemoveCommentsFilter: '%s'", (&SexpArray{Val: expressions, Env: env}).SexpString(0))
	expressions = env.FilterArray(expressions, RemoveCommentsFilter)

	//P("expressions after RemoveCommentsFilter: '%s'", (&SexpArray{Val: expressions, Env: env}).SexpString(0))
	expressions = env.FilterArray(expressions, RemoveEndsFilter)

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

func (env *Zlisp) ParseFile(file string) ([]Sexp, error) {
	in, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	var exp []Sexp

	env.parser.Reset()
	env.parser.NewInput(bufio.NewReader(in))
	exp, err = env.parser.ParseTokens()
	if err != nil {
		return nil, fmt.Errorf("Error on line %d: %v (ParseFile err = '%#v')\n", env.parser.Linenum(), err, err)
	}

	in.Close()

	return exp, nil
}

func (env *Zlisp) LoadStream(stream io.RuneScanner) error {
	env.parser.ResetAddNewInput(stream)
	expressions, err := env.parser.ParseTokens()
	if err != nil {
		if err == ErrMoreInputNeeded {
			panic("where?")
		}
		return fmt.Errorf("Error on line %d: %v (LoadStream err='%#v')\n", env.parser.Linenum(), err, err)
	}
	return env.LoadExpressions(expressions)
}

func (env *Zlisp) EvalString(str string) (Sexp, error) {
	err := env.LoadString(str)
	if err != nil {
		return SexpNull, err
	}
	//VPrintf("\n EvalString: LoadString() done, now to Run():\n")
	return env.Run()
}

// for most things now (except the main repl), prefer EvalFunction() instead of EvalExpressions.
func (env *Zlisp) EvalExpressions(xs []Sexp) (Sexp, error) {
	//P("inside EvalExpressions with env %p: xs[0] = %s", env, xs[0].SexpString(0))
	err := env.LoadExpressions(xs)
	if err != nil {
		return SexpNull, err
	}
	return env.Run()
}

func (env *Zlisp) LoadFile(file io.Reader) error {
	return env.LoadStream(bufio.NewReader(file))
}

func (env *Zlisp) LoadString(str string) error {
	return env.LoadStream(bytes.NewBuffer([]byte(str)))
}

func (env *Zlisp) AddFunction(name string, function ZlispUserFunction) {
	env.AddGlobal(name, MakeUserFunction(name, function))
}

func (env *Zlisp) AddBuilder(name string, function ZlispUserFunction) {
	env.AddGlobal(name, MakeBuilderFunction(name, function))
}

func (env *Zlisp) AddGlobal(name string, obj Sexp) {
	sym := env.MakeSymbol(name)
	env.linearstack.elements[0].(*Scope).Map[sym.number] = obj
}

func (env *Zlisp) AddMacro(name string, function ZlispUserFunction) {
	sym := env.MakeSymbol(name)
	env.macros[sym.number] = MakeUserFunction(name, function)
}

func (env *Zlisp) HasMacro(sym *SexpSymbol) bool {
	_, found := env.macros[sym.number]
	return found
}

func (env *Zlisp) ImportEval() {
	env.AddFunction("eval", EvalFunction)
}

func (env *Zlisp) DumpFunctionByName(name string) error {
	obj, found := env.FindObject(name)
	if !found {
		return errors.New(fmt.Sprintf("%q not found", name))
	}

	var fun ZlispFunction
	switch t := obj.(type) {
	case *SexpFunction:
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
func DumpFunction(fun ZlispFunction, pc int) {
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

func (env *Zlisp) DumpEnvironment() {
	fmt.Printf("PC: %d\n", env.pc)
	fmt.Println("Instructions:")
	if !env.curfunc.user {
		DumpFunction(env.curfunc.fun, env.pc)
	}
	fmt.Printf("DataStack (%p): (length %d)\n", env.datastack, env.datastack.Size())
	env.datastack.PrintStack()
	fmt.Printf("Linear stack: (length %d)\n", env.linearstack.Size())
	//env.linearstack.PrintScopeStack()
	// instead of the above, try:
	env.showStackHelper(env.linearstack, "linearstack")
}

func (env *Zlisp) ReachedEnd() bool {
	return env.pc == env.CurrentFunctionSize()
}

func (env *Zlisp) GetStackTrace(err error) string {
	str := fmt.Sprintf("error in %s:%d: %v\n",
		env.curfunc.name, env.pc, err)
	for !env.addrstack.IsEmpty() {
		fun, pos, _ := env.addrstack.PopAddr()
		str += fmt.Sprintf("in %s:%d\n", fun.name, pos)
	}
	return str
}

func (env *Zlisp) Clear() {
	env.datastack.tos = -1
	env.linearstack.tos = 0
	env.addrstack.tos = -1

	env.mainfunc = env.MakeFunction("__main", 0, false,
		make([]Instruction, 0), nil)
	env.curfunc = env.mainfunc
	env.pc = 0
}

func (env *Zlisp) FindObject(name string) (Sexp, bool) {
	sym := env.MakeSymbol(name)
	obj, err, _ := env.linearstack.LookupSymbol(sym, nil)
	if err != nil {
		return SexpNull, false
	}
	return obj, true
}

func (env *Zlisp) Apply(fun *SexpFunction, args []Sexp) (Sexp, error) {
	//VPrintf("\n\n debug Apply not working on user funcs: fun = '%#v'   and args = '%#v'\n\n", fun, args)
	if fun.user {
		return fun.userfun(env, fun.name, args)
	}

	env.pc = -2
	for _, expr := range args {
		env.datastack.PushExpr(expr)
	}

	//VPrintf("\nApply Calling '%s'\n", fun.SexpString())
	err := env.CallFunction(fun, len(args))
	if err != nil {
		return SexpNull, err
	}

	return env.Run()
}

func (env *Zlisp) Run() (Sexp, error) {

	for env.pc != -1 && !env.ReachedEnd() {
		instr := env.curfunc.fun[env.pc]
		if env.debugExec {
			fmt.Printf("\n ====== in '%s', about to run: '%v'\n",
				env.curfunc.name, instr.InstrString())
			env.DumpEnvironment()
			fmt.Printf("\n ====== in '%s', now running the above.\n",
				env.curfunc.name)
		}
		err := instr.Execute(env)
		if err == StackUnderFlowErr {
			err = nil
		}
		if err != nil {
			return SexpNull, err
		}
		if env.debugExec {
			fmt.Printf("\n ****** in '%s', after running, stack is: \n",
				env.curfunc.name)
			env.DumpEnvironment()
			fmt.Printf("\n ****** \n")

		}
	}

	if env.datastack.IsEmpty() {
		// this does fire.
		//P("debug: *** detected empty datastack, adding a null")
		env.datastack.PushExpr(SexpNull)
	}

	return env.datastack.PopExpr()
}

func (env *Zlisp) AddPreHook(fun PreHook) {
	env.before = append(env.before, fun)
}

func (env *Zlisp) AddPostHook(fun PostHook) {
	env.after = append(env.after, fun)
}

// scan the instruction stream to locate loop start
func (env *Zlisp) FindLoop(target *Loop) (int, error) {
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

func (env *Zlisp) showStackHelper(stack *Stack, name string) {
	note := ""
	n := stack.Top()
	if n < 0 {
		note = "(empty)"
	}
	fmt.Printf(" ========  env(%p).%s is %v deep: %s\n", env, name, n+1, note)
	s := ""
	for i := 0; i <= n; i++ {
		ele, err := stack.Get(n - i)
		if err != nil {
			panic(fmt.Errorf("env.%s access error on %v: %v",
				name, i, err))
		}
		label := fmt.Sprintf("%s %v", name, i)
		switch x := ele.(type) {
		case *Stack:
			s, _ = x.Show(env, nil, label)

		case *Scope:
			s, _ = x.Show(env, nil, label)
		case Scope:
			s, _ = x.Show(env, nil, label)
		default:
			panic(fmt.Errorf("unrecognized element on %s: %T/val=%v",
				name, x, x))
		}
		fmt.Println(s)
	}
}

func (env *Zlisp) dumpParentChain(curfunc *SexpFunction) {

	cur := curfunc
	par := cur.parent
	for par != nil {
		fmt.Printf(" parent chain: cur:%v -> parent:%v\n", cur.name, par.name)
		fmt.Printf("        cur.closures = %s", ClosureToString(cur, env))
		cur = par
		par = par.parent
	}
}

func (env *Zlisp) ShowStackStackAndScopeStack() error {
	env.showStackHelper(env.linearstack, "linearstack")
	fmt.Println(" --- done with env.linearstack, now here is env.curfunc --- ")
	fmt.Println(ClosureToString(env.curfunc, env))
	fmt.Println(" --- done with env.curfunc closure, now here is parent chain: --- ")
	env.dumpParentChain(env.curfunc)
	return nil
}

func (env *Zlisp) ShowGlobalStack() error {
	prev := env.showGlobalScope
	env.showGlobalScope = true
	err := env.ShowStackStackAndScopeStack()
	env.showGlobalScope = prev
	return err
}

func (env *Zlisp) LexicalLookupSymbol(sym *SexpSymbol, setVal *Sexp) (Sexp, error, *Scope) {

	// DotSymbols always evaluate to themselves
	if sym.isDot || sym.isSigil || sym.colonTail {
		return sym, nil, nil
	}

	//P("LexicalLookupSymbol('%s', with setVal: %v)\n", sym.name, setVal)

	const maxFuncToScan = 1 // 1 or otherwise tests/{closure.zy, dynprob.zy, dynscope.zy} will fail.
	exp, err, scope := env.linearstack.LookupSymbolUntilFunction(sym, setVal, maxFuncToScan, false)
	switch err {
	case nil:
		//P("LexicalLookupSymbol('%s') found on env.linearstack(1, false) in scope '%s'\n", sym.name, scope.Name)
		return exp, err, scope
	case SymNotFound:
		break
	}

	// check the parent function lexical captured scopes, if parent available.
	if env.curfunc.parent != nil {
		//P("checking non-nil parent...")
		//exp, err, whichScope := env.curfunc.parent.ClosingLookupSymbol(sym, setVal)
		exp, err, whichScope := env.curfunc.LookupSymbolInParentChainOfClosures(sym, setVal, env)
		switch err {
		case nil:
			//P("LookupSymbolUntilFunction('%s') found in curfunc.parent.ClosingLookupSymbol() scope '%s'\n", sym.name, whichScope.Name)
			return exp, err, whichScope
		default:
			//P("not found  looking via env.curfunc.parent.ClosingLookupSymbol(sym='%s')", sym.name)
			//env.ShowStackStackAndScopeStack()
		}
	} else {

		//fmt.Printf(" *** env.curfunc has closure of: %s\n", ClosureToString(env.curfunc, env))
		//exp, err, scope = env.curfunc.ClosingLookupSymbol(sym, setVal)
		exp, err, scope = env.curfunc.ClosingLookupSymbolUntilFunc(sym, setVal, 1, false)
		switch err {
		case nil:
			//P("LexicalLookupSymbol('%s') found in env.curfunc.ClosingLookupSymbolUnfilFunc(1, false) in scope '%s'\n", sym.name, scope.Name)
			return exp, err, scope
		}
	}

	// with checkCaptures true, as tests/package.zy needs this.
	exp, err, scope = env.linearstack.LookupSymbolUntilFunction(sym, setVal, 2, true)
	switch err {
	case nil:
		//P("LexicalLookupSymbol('%s') found in env.linearstack.LookupSymbolUtilFunction(2, true) in parent runtime scope '%s'\n", sym.name, scope.Name)
		return exp, err, scope
	case SymNotFound:
		break
	}

	return SexpNull, fmt.Errorf("symbol `%s` not found", sym.name), nil
}

func (env *Zlisp) LexicalBindSymbol(sym *SexpSymbol, expr Sexp) error {
	return env.linearstack.BindSymbol(sym, expr)
}

// _closdump : show the closed over env attached to an *SexpFunction
func DumpClosureEnvFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch f := args[0].(type) {
	case *SexpFunction:
		s := ClosureToString(f, env)
		return &SexpStr{S: s}, nil
	default:
		return SexpNull, fmt.Errorf("_closdump needs an *SexpFunction to inspect")
	}
}

func ClosureToString(f *SexpFunction, env *Zlisp) string {
	s, err := f.ShowClosing(env, NewPrintState(),
		fmt.Sprintf("closedOverScopes of '%s'", f.name))
	if err != nil {
		return err.Error()
	}
	return s
}

func (env *Zlisp) IsBuiltinSym(sym *SexpSymbol) (builtin bool, typ string) {

	_, isBuiltin := env.builtins[sym.number]
	if isBuiltin {
		return true, "built-in function"
	}
	_, isBuiltin = env.macros[sym.number]
	if isBuiltin {
		return true, "macro"
	}
	_, isReserved := env.reserved[sym.number]
	if isReserved {
		return true, "reserved word"
	}

	return false, ""
}

func (env *Zlisp) ResolveDotSym(arg []Sexp) ([]Sexp, error) {
	r := []Sexp{}
	for i := range arg {
		switch sym := arg[i].(type) {
		case *SexpSymbol:
			resolved, err := dotGetSetHelper(env, sym.name, nil)
			if err != nil {
				return nil, err
			}
			r = append(r, resolved)
		default:
			r = append(r, arg[i])
		}
	}
	return r, nil
}
