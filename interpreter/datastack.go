package glisp

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

func (stack *Stack) PopExpr() (Sexp, error) {
	elem, err := stack.Pop()
	if err != nil {
		return nil, err
	}
	return elem.(DataStackElem).expr, nil
}

func (stack *Stack) PopExpressions(n int) ([]Sexp, error) {
	stack_start := stack.tos - n + 1
	if stack_start < 0 {
		return nil, errors.New("not enough items on stack")
	}
	arr := make([]Sexp, n)
	for i := 0; i < n; i++ {
		arr[i] = stack.elements[stack_start + i].(DataStackElem).expr
	}
	stack.tos = stack_start - 1
	return arr, nil
}

func (stack *Stack) GetExpr(n int) (Sexp, error) {
	elem, err := stack.Get(n)
	if err != nil {
		return nil, err
	}
	return elem.(DataStackElem).expr, nil
}

func (stack *Stack) PrintStack() {
	fmt.Printf("%d elements\n", stack.tos + 1)
	for i := 0; i <= stack.tos; i++ {
		expr := stack.elements[i].(DataStackElem).expr
		fmt.Println(expr.SexpString())
	}
}
