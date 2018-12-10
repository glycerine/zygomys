package zygo

// where we store our closure-supporing stack pointers
type Closing struct {
	Stack *Stack
	Name  string
	env   *Zlisp
}

func NewClosing(name string, env *Zlisp) *Closing {
	stk := env.linearstack.Clone()
	// be super strict: only store up to our
	// enclosing function definition, because after
	// that, the definition time of that function
	// should be what we use.

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
