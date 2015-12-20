package glisp

import (
	"errors"
	"fmt"
)

type StackElem interface {
	IsStackElem()
}

type Stack struct {
	tos      int
	elements []StackElem
}

func NewStack(size int) *Stack {
	stack := new(Stack)
	stack.tos = -1
	stack.elements = make([]StackElem, size)
	return stack
}

func (stack *Stack) Clone() *Stack {
	ret := &Stack{}
	ret.tos = stack.tos
	ret.elements = make([]StackElem, len(stack.elements))
	for i := range stack.elements {
		ret.elements[i] = stack.elements[i]
	}

	return ret
}

func (stack *Stack) Top() int {
	return stack.tos
}

func (stack *Stack) PushAllTo(target *Stack) int {
	if stack.tos < 0 {
		return 0
	}

	for _, v := range stack.elements[0 : stack.tos+1] {
		target.Push(v)
	}

	return stack.tos + 1
}

func (stack *Stack) IsEmpty() bool {
	return stack.tos == -1
}

func (stack *Stack) Push(elem StackElem) {
	stack.tos++

	if stack.tos == len(stack.elements) {
		stack.elements = append(stack.elements, elem)
	} else {
		stack.elements[stack.tos] = elem
	}
}

func (stack *Stack) Get(n int) (StackElem, error) {
	if stack.tos-n < 0 {
		return nil, errors.New(fmt.Sprint("invalid stack access asked for ", n, " Top was ", stack.tos))
	}
	return stack.elements[stack.tos-n], nil
}

func (stack *Stack) Pop() (StackElem, error) {
	elem, err := stack.Get(0)
	if err != nil {
		return nil, err
	}
	stack.tos--
	return elem, nil
}
