package glisp

type Address int

func (a Address) IsStackElem() {}

func (stack *Stack) PushAddress(addr int) {
	stack.Push(Address(addr))
}

func (stack *Stack) PopAddress() (int, error) {
	elem, err := stack.Pop()
	if err != nil {
		return 0, err
	}
	return int(elem.(Address)), nil
}
