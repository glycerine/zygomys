package glisp

import (
	"errors"
	"fmt"
)

type Generator struct {
	env          *Glisp
	funcname     string
	tail         bool
	scopes       int
	instructions []Instruction
}

func NewGenerator(env *Glisp) *Generator {
	gen := new(Generator)
	gen.env = env
	gen.instructions = make([]Instruction, 0)
	// tail marks whether or not we are in the tail position
	gen.tail = false
	// scopes is the number of extra (non-function) scopes we've created
	gen.scopes = 0
	return gen
}

func (gen *Generator) AddInstructions(instr []Instruction) {
	gen.instructions = append(gen.instructions, instr...)
}

func (gen *Generator) AddInstruction(instr Instruction) {
	gen.instructions = append(gen.instructions, instr)
}

func (gen *Generator) GenerateBegin(expressions []Sexp) error {
	size := len(expressions)
	oldtail := gen.tail
	gen.tail = false
	if size == 0 {
		return errors.New("No expressions found")
	}
	for _, expr := range expressions[:size-1] {
		err := gen.Generate(expr)
		if err != nil {
			return err
		}
		// insert pops after all but the last instruction
		// that way the stack remains clean
		gen.AddInstruction(PopInstr(0))
	}
	gen.tail = oldtail
	return gen.Generate(expressions[size-1])
}

func buildSexpFun(
	env *Glisp,
	name string,
	funcargs SexpArray,
	funcbody []Sexp,
	orig Sexp) (SexpFunction, error) {

	gen := NewGenerator(env)
	gen.tail = true

	if len(name) == 0 {
		gen.funcname = env.GenSymbol("__anon").name
	} else {
		gen.funcname = name
	}

	argsyms := make([]SexpSymbol, len(funcargs))

	for i, expr := range funcargs {
		switch t := expr.(type) {
		case SexpSymbol:
			argsyms[i] = t
		default:
			return MissingFunction,
				errors.New("function argument must be symbol")
		}
	}

	varargs := false
	nargs := len(funcargs)

	if len(argsyms) >= 2 && argsyms[len(argsyms)-2].name == "&" {
		argsyms[len(argsyms)-2] = argsyms[len(argsyms)-1]
		argsyms = argsyms[0 : len(argsyms)-1]
		varargs = true
		nargs = len(argsyms) - 1
	}

	for i := len(argsyms) - 1; i >= 0; i-- {
		gen.AddInstruction(PutInstr{argsyms[i]})
	}
	err := gen.GenerateBegin(funcbody)
	if err != nil {
		return MissingFunction, err
	}
	gen.AddInstruction(ReturnInstr{nil})

	newfunc := GlispFunction(gen.instructions)
	return MakeFunction(gen.funcname, nargs, varargs, newfunc, orig), nil
}

func (gen *Generator) GenerateFn(args []Sexp, orig Sexp) error {
	if len(args) < 2 {
		return errors.New("malformed function definition")
	}

	var funcargs SexpArray

	switch expr := args[0].(type) {
	case SexpArray:
		funcargs = expr
	default:
		return errors.New("function arguments must be in vector")
	}

	funcbody := args[1:]
	sfun, err := buildSexpFun(gen.env, "", funcargs, funcbody, orig)
	if err != nil {
		return err
	}
	gen.AddInstruction(PushInstrClosure{sfun})

	return nil
}

func (gen *Generator) GenerateDef(args []Sexp) error {
	if len(args) != 2 {
		return errors.New("Wrong number of arguments to def")
	}

	var sym SexpSymbol
	switch expr := args[0].(type) {
	case SexpSymbol:
		sym = expr
	default:
		return errors.New("Definition name must by symbol")
	}

	if gen.env.HasMacro(sym) {
		return fmt.Errorf("Already have macro named '%s': refusing "+
			"to define variable of same name.", sym.name)
	}

	gen.tail = false
	err := gen.Generate(args[1])
	if err != nil {
		return err
	}
	gen.AddInstruction(PutInstr{sym})
	gen.AddInstruction(PushInstr{SexpNull})
	return nil
}

func (gen *Generator) GenerateDefn(args []Sexp, orig Sexp) error {
	if len(args) < 3 {
		return errors.New("Wrong number of arguments to defn")
	}

	var funcargs SexpArray
	switch expr := args[1].(type) {
	case SexpArray:
		funcargs = expr
	default:
		return errors.New("function arguments must be in vector")
	}

	var sym SexpSymbol
	switch expr := args[0].(type) {
	case SexpSymbol:
		sym = expr
	default:
		return errors.New("Definition name must by symbol")
	}
	if gen.env.HasMacro(sym) {
		return fmt.Errorf("Already have macro named '%s': refusing"+
			" to define function of same name.", sym.name)
	}

	sfun, err := buildSexpFun(gen.env, sym.name, funcargs, args[2:], orig)
	if err != nil {
		return err
	}

	gen.AddInstruction(PushInstr{sfun})
	gen.AddInstruction(PutInstr{sym})
	gen.AddInstruction(PushInstr{SexpNull})

	return nil
}

func (gen *Generator) GenerateDefmac(args []Sexp, orig Sexp) error {
	if len(args) < 3 {
		return errors.New("Wrong number of arguments to defmac")
	}

	var funcargs SexpArray
	switch expr := args[1].(type) {
	case SexpArray:
		funcargs = expr
	default:
		return errors.New("function arguments must be in vector")
	}

	var sym SexpSymbol
	switch expr := args[0].(type) {
	case SexpSymbol:
		sym = expr
	default:
		return errors.New("Definition name must by symbol")
	}

	sfun, err := buildSexpFun(gen.env, sym.name, funcargs, args[2:], orig)
	if err != nil {
		return err
	}

	gen.env.macros[sym.number] = sfun
	gen.AddInstruction(PushInstr{SexpNull})

	return nil
}

func (gen *Generator) GenerateMacexpand(args []Sexp) error {
	if len(args) != 1 {
		return WrongNargs
	}

	var list SexpPair
	var islist bool
	var ismacrocall bool

	switch t := args[0].(type) {
	case SexpPair:
		if IsList(t.tail) {
			list = t
			islist = true
		}
	default:
		islist = false
	}

	var macro SexpFunction
	if islist {
		switch t := list.head.(type) {
		case SexpSymbol:
			macro, ismacrocall = gen.env.macros[t.number]
		default:
			ismacrocall = false
		}
	}

	if !ismacrocall {
		gen.AddInstruction(PushInstr{args[0]})
		return nil
	}

	macargs, err := ListToArray(list.tail)
	if err != nil {
		return err
	}
	expr, err := gen.env.Apply(macro, macargs)
	if err != nil {
		return err
	}
	gen.AddInstruction(PushInstr{expr})
	return nil
}

func (gen *Generator) GenerateShortCircuit(or bool, args []Sexp) error {
	size := len(args)

	subgen := NewGenerator(gen.env)
	subgen.scopes = gen.scopes
	subgen.tail = gen.tail
	subgen.funcname = gen.funcname
	subgen.Generate(args[size-1])
	instructions := subgen.instructions

	for i := size - 2; i >= 0; i-- {
		subgen = NewGenerator(gen.env)
		subgen.Generate(args[i])
		subgen.AddInstruction(DupInstr(0))
		subgen.AddInstruction(BranchInstr{or, len(instructions) + 2})
		subgen.AddInstruction(PopInstr(0))
		instructions = append(subgen.instructions, instructions...)
	}
	gen.AddInstructions(instructions)

	return nil
}

func (gen *Generator) GenerateCond(args []Sexp) error {
	if len(args)%2 == 0 {
		return errors.New("missing default case")
	}

	subgen := NewGenerator(gen.env)
	subgen.tail = gen.tail
	subgen.scopes = gen.scopes
	subgen.funcname = gen.funcname
	err := subgen.Generate(args[len(args)-1])
	if err != nil {
		return err
	}
	instructions := subgen.instructions

	for i := len(args)/2 - 1; i >= 0; i-- {
		subgen.Reset()
		err := subgen.Generate(args[2*i])
		if err != nil {
			return err
		}
		pred_code := subgen.instructions

		subgen.Reset()
		subgen.tail = gen.tail
		subgen.scopes = gen.scopes
		subgen.funcname = gen.funcname
		err = subgen.Generate(args[2*i+1])
		if err != nil {
			return err
		}
		body_code := subgen.instructions

		subgen.Reset()
		subgen.AddInstructions(pred_code)
		subgen.AddInstruction(BranchInstr{false, len(body_code) + 2})
		subgen.AddInstructions(body_code)
		subgen.AddInstruction(JumpInstr{len(instructions) + 1})
		subgen.AddInstructions(instructions)
		instructions = subgen.instructions
	}

	gen.AddInstructions(instructions)
	return nil
}

func (gen *Generator) GenerateQuote(args []Sexp) error {
	for _, expr := range args {
		gen.AddInstruction(PushInstr{expr})
	}
	return nil
}

func (gen *Generator) GenerateSyntaxQuote(args []Sexp) error {
	if len(args) != 1 {
		return errors.New("syntax-quote takes 1 argument")
	}

	if args[0] == SexpNull || !IsList(args[0]) {
		gen.AddInstruction(PushInstr{args[0]})
		return nil
	}
	quotebody, _ := ListToArray(args[0])

	if len(quotebody) == 2 {
		var issymbol bool
		var sym SexpSymbol
		switch t := quotebody[0].(type) {
		case SexpSymbol:
			sym = t
			issymbol = true
		default:
			issymbol = false
		}
		if issymbol {
			if sym.name == "unquote" {
				gen.Generate(quotebody[1])
				return nil
			} else if sym.name == "unquote-splicing" {
				gen.Generate(quotebody[1])
				gen.AddInstruction(ExplodeInstr(0))
				return nil
			}
		}
	}

	gen.AddInstruction(PushInstr{SexpMarker})

	for _, expr := range quotebody {
		gen.GenerateSyntaxQuote([]Sexp{expr})
	}

	gen.AddInstruction(SquashInstr(0))

	return nil
}

func (gen *Generator) GenerateLet(name string, args []Sexp) error {
	if len(args) < 2 {
		return errors.New("malformed let statement")
	}

	lstatements := make([]SexpSymbol, 0)
	rstatements := make([]Sexp, 0)
	var bindings []Sexp

	switch expr := args[0].(type) {
	case SexpArray:
		bindings = []Sexp(expr)
	default:
		return errors.New("let bindings must be in array")
	}

	if len(bindings)%2 != 0 {
		return errors.New("uneven let binding list")
	}

	for i := 0; i < len(bindings)/2; i++ {
		switch t := bindings[2*i].(type) {
		case SexpSymbol:
			lstatements = append(lstatements, t)
		default:
			return errors.New("cannot bind to non-symbol")
		}
		rstatements = append(rstatements, bindings[2*i+1])
	}

	gen.AddInstruction(AddScopeInstr(0))
	gen.scopes++

	if name == "let*" {
		for i, rs := range rstatements {
			err := gen.Generate(rs)
			if err != nil {
				return err
			}
			gen.AddInstruction(PutInstr{lstatements[i]})
		}
	} else if name == "let" {
		for _, rs := range rstatements {
			err := gen.Generate(rs)
			if err != nil {
				return err
			}
		}
		for i := len(lstatements) - 1; i >= 0; i-- {
			gen.AddInstruction(PutInstr{lstatements[i]})
		}
	}
	err := gen.GenerateBegin(args[1:])
	if err != nil {
		return err
	}
	gen.AddInstruction(RemoveScopeInstr(0))
	gen.scopes--

	return nil
}

func (gen *Generator) GenerateAssert(args []Sexp) error {
	if len(args) != 1 {
		return WrongNargs
	}
	err := gen.Generate(args[0])
	if err != nil {
		return err
	}

	reterrmsg := fmt.Sprintf("Assertion failed: %s\n",
		args[0].SexpString())
	gen.AddInstruction(BranchInstr{true, 2})
	gen.AddInstruction(ReturnInstr{errors.New(reterrmsg)})
	gen.AddInstruction(PushInstr{SexpNull})
	return nil
}

func (gen *Generator) GenerateInclude(args []Sexp) error {
	if len(args) < 1 {
		return WrongNargs
	}

	var err error
	var exps []Sexp

	var sourceItem func(item Sexp) error

	sourceItem = func(item Sexp) error {
		switch t := item.(type) {
		case SexpArray:
			for _, v := range t {
				if err := sourceItem(v); err != nil {
					return err
				}
			}
		case SexpPair:
			expr := item
			for expr != SexpNull {
				list := expr.(SexpPair)
				if err := sourceItem(list.head); err != nil {
					return err
				}
				expr = list.tail
			}
		case SexpStr:
			exps, err = gen.env.ParseFile(string(t))
			if err != nil {
				return err
			}

			err = gen.GenerateBegin(exps)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("include: Expected `string`, `list`, `array` given type %T val %v", item, item)
		}

		return nil
	}

	for _, v := range args {
		err = sourceItem(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gen *Generator) GenerateCallBySymbol(sym SexpSymbol, args []Sexp, orig Sexp) error {
	switch sym.name {
	case "and":
		return gen.GenerateShortCircuit(false, args)
	case "or":
		return gen.GenerateShortCircuit(true, args)
	case "cond":
		return gen.GenerateCond(args)
	case "quote":
		return gen.GenerateQuote(args)
	case "def":
		return gen.GenerateDef(args)
	case "fn":
		return gen.GenerateFn(args, orig)
	case "defn":
		return gen.GenerateDefn(args, orig)
	case "begin":
		return gen.GenerateBegin(args)
	case "let":
		return gen.GenerateLet("let", args)
	case "let*":
		return gen.GenerateLet("let*", args)
	case "assert":
		return gen.GenerateAssert(args)
	case "defmac":
		return gen.GenerateDefmac(args, orig)
	case "macexpand":
		return gen.GenerateMacexpand(args)
	case "syntax-quote":
		return gen.GenerateSyntaxQuote(args)
	case "include":
		return gen.GenerateInclude(args)
	}

	macro, found := gen.env.macros[sym.number]
	if found {
		// calling Apply on the current environment will screw up
		// the stack, creating a duplicate environment is safer
		env := gen.env.Duplicate()
		expr, err := env.Apply(macro, args)
		if err != nil {
			return err
		}
		return gen.Generate(expr)
	}

	oldtail := gen.tail
	gen.tail = false
	err := gen.GenerateAll(args)
	if err != nil {
		return err
	}
	if oldtail && sym.name == gen.funcname {
		// to do a tail call
		// pop off all the extra scopes
		// then jump to beginning of function
		for i := 0; i < gen.scopes; i++ {
			gen.AddInstruction(RemoveScopeInstr(0))
		}
		gen.AddInstruction(GotoInstr{0})
	} else {
		gen.AddInstruction(CallInstr{sym, len(args)})
	}
	gen.tail = oldtail
	return nil
}

func (gen *Generator) GenerateDispatch(fun Sexp, args []Sexp) error {
	gen.GenerateAll(args)
	gen.Generate(fun)
	gen.AddInstruction(DispatchInstr{len(args)})
	return nil
}

func (gen *Generator) GenerateCall(expr SexpPair) error {
	arr, _ := ListToArray(expr.tail)
	switch head := expr.head.(type) {
	case SexpSymbol:
		return gen.GenerateCallBySymbol(head, arr, expr)
	}
	return gen.GenerateDispatch(expr.head, arr)
}

func (gen *Generator) GenerateArray(arr SexpArray) error {
	err := gen.GenerateAll(arr)
	if err != nil {
		return err
	}
	gen.AddInstruction(CallInstr{gen.env.MakeSymbol("array"), len(arr)})
	return nil
}

func (gen *Generator) Generate(expr Sexp) error {
	switch e := expr.(type) {
	case SexpSymbol:
		gen.AddInstruction(GetInstr{e})
		return nil
	case SexpPair:
		if IsList(e) {
			err := gen.GenerateCall(e)
			if err != nil {
				return errors.New(
					fmt.Sprintf("Error generating %s:\n%v",
						expr.SexpString(), err))
			}
			return nil
		} else {
			gen.AddInstruction(PushInstr{expr})
		}
	case SexpArray:
		return gen.GenerateArray(e)
	default:
		gen.AddInstruction(PushInstr{expr})
		return nil
	}
	return nil
}

func (gen *Generator) GenerateAll(expressions []Sexp) error {
	for _, expr := range expressions {
		err := gen.Generate(expr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (gen *Generator) Reset() {
	gen.instructions = make([]Instruction, 0)
	gen.tail = false
	gen.scopes = 0
}
