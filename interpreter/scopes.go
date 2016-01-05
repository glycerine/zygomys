package gdsl

import (
	"errors"
	"fmt"
)

type Scope map[int]Sexp

func (s Scope) IsStackElem() {}

func (stack *Stack) PushScope() {
	stack.Push(Scope(make(map[int]Sexp)))
}

func (stack *Stack) PopScope() error {
	_, err := stack.Pop()
	return err
}

func (stack *Stack) lookupSymbol(sym SexpSymbol, minFrame int) (Sexp, error) {
	if !stack.IsEmpty() {
		for i := 0; i <= stack.tos-minFrame; i++ {
			elem, err := stack.Get(i)
			if err != nil {
				return SexpNull, err
			}
			scope := map[int]Sexp(elem.(Scope))
			expr, ok := scope[sym.number]
			if ok {
				return expr, nil
			}
		}
	}
	return SexpNull, errors.New(fmt.Sprint("symbol ", sym, " not found"))
}

func (stack *Stack) LookupSymbol(sym SexpSymbol) (Sexp, error) {
	return stack.lookupSymbol(sym, 0)
}

// LookupSymbolNonGlobal  - closures use this to only find symbols below the global scope, to avoid copying globals it'll always be-able to ref
func (stack *Stack) LookupSymbolNonGlobal(sym SexpSymbol) (Sexp, error) {
	return stack.lookupSymbol(sym, 1)
}

func (stack *Stack) BindSymbol(sym SexpSymbol, expr Sexp) error {
	if stack.IsEmpty() {
		return errors.New("no scope available")
	}
	stack.elements[stack.tos].(Scope)[sym.number] = expr
	return nil
}
