package zygo

import (
	"errors"
	"fmt"
)

type DataStackElem struct {
	expr Sexp
}

func (d DataStackElem) IsStackElem() {}

func (stack *Stack) PushExpr(expr Sexp) {
	stack.Push(DataStackElem{expr})
}

func (stack *Stack) PushExpressions(expr []Sexp) error {
	for _, x := range expr {
		stack.Push(DataStackElem{x})
	}
	return nil
}

func (stack *Stack) PopExpr() (Sexp, error) {
	elem, err := stack.Pop()
	if err != nil {
		return nil, err
	}
	return elem.(DataStackElem).expr, nil
}

func (stack *Stack) GetExpressions(n int) ([]Sexp, error) {
	stack_start := stack.tos - n + 1
	if stack_start < 0 {
		return nil, errors.New("not enough items on stack")
	}
	arr := make([]Sexp, n)
	for i := 0; i < n; i++ {
		arr[i] = stack.elements[stack_start+i].(DataStackElem).expr
	}
	return arr, nil
}

func (stack *Stack) PopExpressions(n int) ([]Sexp, error) {
	expressions, err := stack.GetExpressions(n)
	if err != nil {
		return nil, err
	}
	stack.tos -= n
	return expressions, nil
}

func (stack *Stack) GetExpr(n int) (Sexp, error) {
	elem, err := stack.Get(n)
	if err != nil {
		return nil, err
	}
	return elem.(DataStackElem).expr, nil
}

func (stack *Stack) PrintStack() {
	for i := 0; i <= stack.tos; i++ {
		expr := stack.elements[i].(DataStackElem).expr
		fmt.Println("\t" + expr.SexpString(0))
	}
}

func (stack *Stack) PrintScopeStack() {
	for i := 0; i <= stack.tos; i++ {
		scop := stack.elements[i].(*Scope)
		scop.Show(stack.env, 4, "")
	}
}
