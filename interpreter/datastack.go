package glisp

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

func (stack *Stack) GetExpr(n int) (Sexp, error) {
	elem, err := stack.Get(n)
	if err != nil {
		return nil, err
	}
	return elem.(DataStackElem).expr, nil
}
