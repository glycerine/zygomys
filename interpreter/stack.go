package glisp

import (
	"errors"
)

const StackStartSize = 10

type StackElem interface {
	IsStackElem()
}

type Stack struct {
	tos int
	elements []StackElem
}

func NewStack() *Stack {
	stack := new(Stack)
	stack.tos = -1
	stack.elements = make([]StackElem, StackStartSize)
	return stack
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
	if stack.tos - n < 0 {
		return nil, errors.New("invalid stack access")
	}
	return stack.elements[stack.tos - n], nil
}

func (stack *Stack) Pop() (StackElem, error) {
	elem, err := stack.Get(0)
	if err != nil {
		return nil, err
	}
	stack.tos--
	return elem, nil
}
