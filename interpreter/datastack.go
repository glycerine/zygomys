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

func (stack *Stack) PopExpressions(n int) ([]Sexp, error) {
	arr := make([]Sexp, 0)
	for i := 0; i < n; i++ {
		elem, err := stack.Pop()
		if err != nil {
			return nil, err
		}
		arr = append(arr, elem.(DataStackElem).expr)
	}
	ReverseArray(arr)
	return arr, nil
}

func (stack *Stack) GetExpr(n int) (Sexp, error) {
	elem, err := stack.Get(n)
	if err != nil {
		return nil, err
	}
	return elem.(DataStackElem).expr, nil
}

// reverse array in-place
func ReverseArray(arr []Sexp) {
	size := len(arr)
	for i := 0; i < size / 2; i++ {
		temp := arr[i]
		arr[i] = arr[size - i - 1]
		arr[size - i - 1] = temp
	}
}
