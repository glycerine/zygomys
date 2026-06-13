package zygo

import (
	"fmt"
)

// FunctionCallNameTypeCheck type checks a function call.
func (env *Zlisp) FunctionCallNameTypeCheck(f *SexpFunction, nargs *int) error {
	if f.inputTypes != nil {
		//Q("FunctionCallNameTypeCheck sees inputTypes: '%v'", f.inputTypes.SexpString(nil))
	} else {
		return nil // no type checking requested
	}
	if f.varargs {
		return nil // name/type checking for vargarg not currently implemented.
	}

	// our call arguments prepared -- will be pushed to the datastack
	finalArgs := make([]Sexp, f.inputTypes.NumKeys)

	// pop everything off the stack, will push finalArgs later
	exprs, err := env.datastack.PopExpressions(*nargs)
	if err != nil {
		return err
	}

	// map the named submitted args, for fast lookup by name
	submittedByName := make(map[string]Sexp)

	// prep submittedByName
	for i := 0; i < *nargs; i++ {
		sym, isNamed := namedArgSymbol(exprs[i])
		if !isNamed {
			//Q("in env.CallFunction, exprs[%v]='%v'/type=%T", i, exprs[i].SexpString(nil), exprs[i])
			continue
		}
		//Q("in env.CallFunction, have symbol.colonTail: exprs[%v]='%#v'", i, sym)
		_, err := f.inputTypes.HashGet(env, sym)
		if err != nil {
			return fmt.Errorf("%s takes no argument '%s'", f.name, sym.name)
		}
		if i == (*nargs)-1 {
			return fmt.Errorf("named parameter '%s' not followed by value", sym.name)
		}
		val := exprs[i+1]
		i++
		_, already := submittedByName[sym.name]
		if already {
			return fmt.Errorf("duplicate named parameter '%s'", sym.name)
		}

		submittedByName[sym.name] = val
	}

	// simplify name matching for now with this rule: all by name, or none.
	haveByName := len(submittedByName)
	if haveByName > 0 {
		if haveByName != f.inputTypes.NumKeys {
			return fmt.Errorf("named arguments count %v != expected %v", haveByName, f.inputTypes.NumKeys)
		}

		// prep finalArgs in the order dictated
		for i, key := range f.inputTypes.KeyOrder {
			switch sy := key.(type) {
			case *SexpSymbol:
				// search for sy.name in our submittedByName args
				a, found := submittedByName[sy.name]
				if found {
					//Q("%s call: matching %v-th argument named '%s': passing value '%s'",
					// f.name, i, sy.name, a.SexpString(nil))
					finalArgs[i] = a
				}
			default:
				return fmt.Errorf("unsupported argument-name type %T", key)
			}

		}
	} else {
		// not using named parameters, restore the arguments to the stack as they were.
		finalArgs = exprs
	}
	finalArgs, err = env.prepareLazyFinalArgs(f, finalArgs)
	if err != nil {
		return err
	}
	for i, val := range finalArgs {
		if i >= len(f.inputTypes.KeyOrder) {
			break
		}
		if f.IsLazyCallArg(i) {
			continue
		}
		typ, err := f.inputTypes.HashGet(env, f.inputTypes.KeyOrder[i])
		if err != nil {
			return err
		}
		valtyp := val.Type()
		if typ != nil && typ != valtyp {
			sym, _ := f.inputTypes.KeyOrder[i].(*SexpSymbol)
			name := f.inputTypes.KeyOrder[i].SexpString(nil)
			if sym != nil {
				name = sym.name
			}
			return fmt.Errorf("type mismatch for parameter '%s': expected '%s', got '%s'",
				name, typ.SexpString(nil), valtyp.SexpString(nil))
		}
	}
	*nargs = len(finalArgs)
	return env.datastack.PushExpressions(finalArgs)
}

func namedArgSymbol(expr Sexp) (*SexpSymbol, bool) {
	switch x := expr.(type) {
	case *SexpSymbol:
		return x, x.colonTail
	case *SexpLazyArg:
		sym, ok := x.Expr.(*SexpSymbol)
		return sym, ok && sym.colonTail
	default:
		return nil, false
	}
}
