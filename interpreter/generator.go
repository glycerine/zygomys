package glisp

import (
	"errors"
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

func buildSexpFun(env *Glisp, name string, funcargs SexpArray,
	funcbody []Sexp) (SexpFunction, error) {
	gen := NewGenerator(env)
	gen.tail = true

	if len(name) == 0 {
		gen.funcname = env.GenSymbol("__anon").name
	} else {
		gen.funcname = name
	}

	for i := len(funcargs) - 1; i >= 0; i-- {
		var argsym SexpSymbol
		switch expr := funcargs[i].(type) {
		case SexpSymbol:
			argsym = expr
		default:
			return MissingFunction,
				errors.New("function argument must be symbol")
		}
		gen.AddInstruction(PutInstr{argsym})
	}
	err := gen.GenerateBegin(funcbody)
	if err != nil {
		return MissingFunction, err
	}
	gen.AddInstruction(ReturnInstr{nil})

	newfunc := GlispFunction(gen.instructions)
	return MakeFunction(gen.funcname, len(funcargs), newfunc), nil
}

func (gen *Generator) GenerateFn(args []Sexp) error {
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
	sfun, err := buildSexpFun(gen.env, "", funcargs, funcbody)
	if err != nil {
		return err
	}
	gen.AddInstruction(PushInstr{sfun})

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

	gen.tail = false
	err := gen.Generate(args[1])
	if err != nil {
		return err
	}
	gen.AddInstruction(PutInstr{sym})
	gen.AddInstruction(PushInstr{SexpNull})
	return nil
}

func (gen *Generator) GenerateDefn(args []Sexp) error {
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

	sfun, err := buildSexpFun(gen.env, sym.name, funcargs, args[2:])
	if err != nil {
		return err
	}

	gen.AddInstruction(PushInstr{sfun})
	gen.AddInstruction(PutInstr{sym})
	gen.AddInstruction(PushInstr{SexpNull})

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

func (gen *Generator) GenerateCallBySymbol(sym SexpSymbol, args []Sexp) error {
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
		return gen.GenerateFn(args)
	case "defn":
		return gen.GenerateDefn(args)
	case "begin":
		return gen.GenerateBegin(args)
	case "let":
		return gen.GenerateLet("let", args)
	case "let*":
		return gen.GenerateLet("let*", args)
	}
	oldtail := gen.tail
	gen.GenerateAll(args)
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
		return gen.GenerateCallBySymbol(head, arr)
	}
	return gen.GenerateDispatch(expr.head, arr)
}

func (gen *Generator) Generate(expr Sexp) error {
	switch e := expr.(type) {
	case SexpSymbol:
		gen.AddInstruction(GetInstr{e})
		return nil
	case SexpPair:
		if IsList(e) {
			return gen.GenerateCall(e)
		} else {
			gen.AddInstruction(PushInstr{expr})
		}
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
