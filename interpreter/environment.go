package glisp

type Glisp struct {
	datastack *Stack
	scopestack *Stack
	symtable map[string]int
	nextsymbol int
	running bool
}

func NewGlisp() *Glisp {
	env := new(Glisp)
	env.datastack = NewStack()
	env.scopestack = NewStack()
	env.symtable = make(map[string]int)
	env.nextsymbol = 0
	env.running = false
	return env
}

func (env *Glisp) MakeSymbol(name string) SexpSymbol {
	symnum, ok := env.symtable[name]
	if ok {
		return SexpSymbol{name, symnum}
	}
	symbol := SexpSymbol{name, env.nextsymbol}
	env.symtable[name] = env.nextsymbol
	env.nextsymbol++
	return symbol
}
