package glisp

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

func (stack *Stack) LookupSymbol(sym SexpSymbol) (Sexp, error) {
	if !stack.IsEmpty() {
		for i := 0; i <= stack.tos; i++ {
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
	errmsg := fmt.Sprintf("symbol %s not found", sym.name)
	return SexpNull, errors.New(errmsg)
}

func (stack *Stack) BindSymbol(sym SexpSymbol, expr Sexp) error {
	if stack.IsEmpty() {
		return errors.New("no scope available")
	}
	stack.elements[stack.tos].(Scope)[sym.number] = expr
	return nil
}
