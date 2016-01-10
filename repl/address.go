package zygo

type Address struct {
	function SexpFunction
	position int
}

func (a Address) IsStackElem() {}

func (stack *Stack) PushAddr(function SexpFunction, pc int) {
	stack.Push(Address{function, pc})
}

func (stack *Stack) PopAddr() (SexpFunction, int, error) {
	elem, err := stack.Pop()
	if err != nil {
		return MissingFunction, 0, err
	}
	addr := elem.(Address)
	return addr.function, addr.position, nil
}
