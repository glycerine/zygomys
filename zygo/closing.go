package zygo

// where we store our closure-supporing stack pointers
type Closing struct {
	Stack *Stack
	Name  string
	env   *Zlisp
}

func NewClosing(name string, env *Zlisp) *Closing {
	stk := env.linearstack.Clone()
	// Be strict: only store scopes from the current lexical function outward
	// to the current top. Caller scopes below that function are dynamic state.
	for i := stk.tos; i >= 0; i-- {
		scop, ok := stk.elements[i].(*Scope)
		if ok && scop.IsFunction {
			if i > 0 {
				trimmed := env.NewStack(stk.Size() - i)
				for _, elem := range stk.elements[i : stk.tos+1] {
					trimmed.Push(elem)
				}
				stk = trimmed
			}
			break
		}
	}

	return &Closing{
		Stack: stk,
		Name:  name,
		env:   env}
}

func NewEmptyClosing(name string, env *Zlisp) *Closing {
	return &Closing{
		Stack: env.NewStack(0),
		Name:  name,
		env:   env}
}

func (c *Closing) IsStackElem() {}

func (c *Closing) LookupSymbolUntilFunction(sym *SexpSymbol, setVal *Sexp, maximumFuncToSearch int, checkCaptures bool) (Sexp, error, *Scope) {
	return c.Stack.LookupSymbolUntilFunction(sym, setVal, maximumFuncToSearch, checkCaptures)
}
func (c *Closing) LookupSymbol(sym *SexpSymbol, setVal *Sexp) (Sexp, error, *Scope) {
	return c.Stack.LookupSymbol(sym, setVal)
}

func (c *Closing) Show(env *Zlisp, ps *PrintState, label string) (string, error) {
	return c.Stack.Show(env, ps, label)
}

func (c *Closing) TopScope() *Scope {
	return c.Stack.GetTop().(*Scope)
}
