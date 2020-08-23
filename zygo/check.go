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
		switch sym := exprs[i].(type) {
		case *SexpSymbol:
			if sym.colonTail {
				//Q("in env.CallFunction, have symbol.colonTail: exprs[%v]='%#v'", i, sym)
				typ, err := f.inputTypes.HashGet(env, sym)
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
				valtyp := val.Type()
				if typ != nil && typ != valtyp {
					return fmt.Errorf("type mismatch for parameter '%s': expected '%s', got '%s'",
						sym.name, typ.SexpString(nil), valtyp.SexpString(nil))
				}
			} else {
				//Q("in env.CallFunction, exprs[%v]='%v'/type=%T", i, exprs[i].SexpString(nil), exprs[i])
			}
		default:
			//Q("in env.CallFunction, exprs[%v]='%v'/type=%T", i, exprs[i].SexpString(nil), exprs[i])
		}
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
	*nargs = len(finalArgs)
	return env.datastack.PushExpressions(finalArgs)
}
