package glisp

type Address struct {
	function GlispFunction
	position int
}

func (a Address) IsStackElem() {}

func (stack *Stack) PushAddr(function GlispFunction, pc int) {
	stack.Push(Address{function, pc})
}

func (stack *Stack) PopAddr() (GlispFunction, int, error) {
	elem, err := stack.Pop()
	if err != nil {
		return nil, 0, err
	}
	addr := elem.(Address)
	return addr.function, addr.position, nil
}
