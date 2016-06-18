package zygo

// where we store our closure-supporing stack pointers
type Closing struct {
	Stack *Stack
	Name  string
	env   *Glisp
}

func NewClosing(name string, env *Glisp) *Closing {
	return &Closing{
		Stack: env.linearstack.Clone(),
		Name:  name,
		env:   env}
}

func (c *Closing) IsStackElem() {}

func (c *Closing) LookupSymbolUntilFunction(sym *SexpSymbol) (Sexp, error, *Scope) {
	return c.Stack.LookupSymbolUntilFunction(sym, nil)
}
func (c *Closing) LookupSymbol(sym *SexpSymbol, setVal *Sexp) (Sexp, error, *Scope) {
	return c.Stack.LookupSymbol(sym, setVal)
}

func (c *Closing) Show(env *Glisp, indent int, label string) (string, error) {
	return c.Stack.Show(env, indent, label)
}

func (c *Closing) TopScope() *Scope {
	return c.Stack.GetTop().(*Scope)
}
