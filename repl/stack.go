package zygo

import (
	"fmt"
	"strings"
)

type StackElem interface {
	IsStackElem()
}

type Stack struct {
	tos      int
	elements []StackElem
	env      *Glisp

	Name string // type name

	// package support:
	PackageName string
	IsPackage   bool
}

// SexpString satisfies the Sexp interface, producing a string presentation of the value.
func (s *Stack) SexpString(indent int) string {
	var label string
	head := ""
	if s.IsPackage {
		head = "(stackpackage " + s.PackageName
	} else {
		label = "scope " + s.Name
	}

	str, err := s.Show(s.env, indent, s.Name)
	if err != nil {
		return "(" + label + ")"
	}

	return head + " " + str + " )"
}

// Type() satisfies the Sexp interface, returning the type of the value.
func (s *Stack) Type() *RegisteredType {
	return GoStructRegistry.Lookup("packageScopeStack")
}

func (env *Glisp) NewStack(size int) *Stack {
	return &Stack{
		tos:      -1,
		elements: make([]StackElem, size),
		env:      env,
	}
}

func (stack *Stack) Clone() *Stack {
	ret := &Stack{}
	ret.tos = stack.tos
	ret.env = stack.env
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
	return stack.tos < 0
}

func (stack *Stack) Push(elem StackElem) {
	stack.tos++

	if stack.tos == len(stack.elements) {
		stack.elements = append(stack.elements, elem)
	} else {
		stack.elements[stack.tos] = elem
	}
}

func (stack *Stack) GetTop() StackElem {
	s, err := stack.Get(0)
	if err != nil {
		panic(err)
	}
	return s
}
func (stack *Stack) Size() int {
	return stack.tos + 1
}

var StackUnderFlowErr = fmt.Errorf("invalid stack access: underflow")

func (stack *Stack) Get(n int) (StackElem, error) {
	if stack.tos-n < 0 {
		err := StackUnderFlowErr
		//panic(err)
		return nil, err
	}
	return stack.elements[stack.tos-n], nil
}

func (stack *Stack) Pop() (StackElem, error) {
	// always make a new array,
	// so we can use for the closure stack-of-scopes.

	elem, err := stack.Get(0)
	if err != nil {
		return nil, err
	}

	// invar n > 0
	n := stack.Size()
	if n == 0 {
		return nil, fmt.Errorf("Stack.Pop() on emtpy stack")
	}

	el := make([]StackElem, n-1)
	copy(el, stack.elements)
	stack.elements = el
	stack.tos--
	return elem, nil
}

func (stack *Stack) PopAndDiscard() {
	stack.tos--
	if stack.tos < -1 {
		stack.tos = -1
	}
}

func (stack *Stack) IsStackElem() {}

func (stack Stack) Show(env *Glisp, indent int, label string) (string, error) {
	s := ""
	rep := strings.Repeat(" ", indent)
	s += fmt.Sprintf("%s %s\n", rep, label)
	n := stack.Top()
	for i := 0; i <= n; i++ {
		ele, err := stack.Get(n - i)
		if err != nil {
			panic(fmt.Errorf("stack access error on %v: %v", i, err))
		}
		showme, canshow := ele.(Showable)
		if canshow {
			r, err := showme.Show(env, indent+4,
				fmt.Sprintf("elem %v (%#v) of %s:", i, showme, label))
			if err != nil {
				return "", err
			}
			s += r
		}
	}
	return s, nil
}

// set newsize to 0 to truncate everything
func (s *Stack) TruncateToSize(newsize int) {
	s.tos = newsize - 1
}

// nestedPathGetSet does a top-down lookup, as opposed to LexicalLookupSymbol which is bottom up
func (s *Stack) nestedPathGetSet(env *Glisp, dotpaths []string, setVal *Sexp) (Sexp, error) {

	if len(dotpaths) == 0 {
		return SexpNull, fmt.Errorf("internal error: in nestedPathGetSet() dotpaths" +
			" had zero length")
	}

	curStack := s

	var ret Sexp = SexpNull
	var err error
	var scop *Scope
	lenpath := len(dotpaths)
	P("\n in nestedPathGetSet, dotpaths=%#v\n", dotpaths)
	for i := range dotpaths {

		curSym := env.MakeSymbol(stripAnyDotPrefix(dotpaths[i]))
		if !curStack.IsPackage {
			return SexpNull, fmt.Errorf("error locating symbol '%s': current Stack is not a package", curSym.name)
		}

		ret, err, scop = curStack.LookupSymbol(curSym, nil)
		if err != nil {
			return SexpNull, fmt.Errorf("could not find symbol '%s' in current package '%v'",
				curSym.name, curStack.PackageName)
		}
		if setVal != nil && i == lenpath-1 {
			// assign now
			scop.Map[curSym.number] = *setVal
			// done with SET
			return *setVal, nil
		}

		if i == lenpath-1 {
			// done with GET
			return ret, nil
		}
		// invar: i < lenpath-1, so go deeper
		switch x := ret.(type) {
		case *SexpHash:
			P("\n found hash in x at i=%d, looping to next i\n", i)
			return x.nestedPathGetSet(env, dotpaths[1:], setVal)
		case *Stack:
			curStack = x
		default:
			return SexpNull, fmt.Errorf("not a record or scope: cannot get field '%s'"+
				" out of type %T)", dotpaths[i+1][1:], x)
		}

	}
	return ret, nil
}
