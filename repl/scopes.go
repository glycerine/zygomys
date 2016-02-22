package zygo

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Scope struct {
	Map        map[int]Sexp
	IsGlobal   bool
	Name       string
	Parent     *Scope
	IsFunction bool // if true, read-only.
	env        *Glisp
}

func (env *Glisp) NewScope() *Scope {
	return &Scope{
		Map: make(map[int]Sexp),
		env: env,
	}
}

func (env *Glisp) NewNamedScope(name string) *Scope {
	return &Scope{
		Map:  make(map[int]Sexp),
		Name: name,
		env:  env,
	}
}

func (s *Scope) CloneScope() *Scope {
	n := s.env.NewScope()
	for k, v := range s.Map {
		n.Map[k] = v
	}
	return n
}

func (s Scope) IsStackElem() {}

func (stack *Stack) PushScope() {
	s := stack.env.NewScope()
	if stack.Size() > 0 {
		s.Parent = stack.GetTop().(*Scope)
	}
	stack.Push(s)
}

func (stack *Stack) PopScope() error {
	_, err := stack.Pop()
	return err
}

// dynamic scoping lookup. See env.LexicalLookupSymbol() for the lexically
// scoped equivalent.
func (stack *Stack) lookupSymbol(sym SexpSymbol, minFrame int) (Sexp, error, *Scope) {
	if !stack.IsEmpty() {
		for i := 0; i <= stack.tos-minFrame; i++ {
			elem, err := stack.Get(i)
			if err != nil {
				return SexpNull, err, nil
			}
			switch scope := elem.(type) {
			case (*Scope):
				expr, ok := scope.Map[sym.number]
				if ok {
					return expr, nil, scope
				}
			}
		}
	}

	if stack.env != nil && stack.env.debugSymbolNotFound {
		stack.env.ShowStackStackAndScopeStack()
	}
	return SexpNull, fmt.Errorf("symbol `%s` not found", sym.name), nil
}

func (stack *Stack) LookupSymbol(sym SexpSymbol) (Sexp, error, *Scope) {
	return stack.lookupSymbol(sym, 0)
}

// LookupSymbolNonGlobal  - closures use this to only find symbols below the global scope, to avoid copying globals it'll always be-able to ref
func (stack *Stack) LookupSymbolNonGlobal(sym SexpSymbol) (Sexp, error, *Scope) {
	return stack.lookupSymbol(sym, 1)
}

var SymNotFound = errors.New("symbol not found")

// lookup symbols, but don't go beyond a function boundary
func (stack *Stack) LookupSymbolUntilFunction(sym SexpSymbol) (Sexp, error, *Scope) {

	if !stack.IsEmpty() {
	doneSearching:
		for i := 0; i <= stack.tos; i++ {
			elem, err := stack.Get(i)
			if err != nil {
				return SexpNull, err, nil
			}
			switch scope := elem.(type) {
			case (*Scope):
				VPrintf("   ...looking up in scope '%s'\n", scope.Name)
				expr, ok := scope.Map[sym.number]
				if ok {
					return expr, nil, scope
				}
				if scope.IsFunction {
					VPrintf("   ...scope '%s' was a function, halting up search\n",
						scope.Name)
					break doneSearching
				}
			}
		}
	}

	if stack != nil && stack.env != nil && stack.env.debugSymbolNotFound {
		fmt.Printf("debugSymbolNotFound is true, here are scopes:\n")
		stack.env.ShowStackStackAndScopeStack()
	}
	return SexpNull, SymNotFound, nil
}

func (stack *Stack) BindSymbol(sym SexpSymbol, expr Sexp) error {
	if stack.IsEmpty() {
		panic("empty stack!!")
		//return errors.New("no scope available")
	}
	cur, already := stack.elements[stack.tos].(*Scope).Map[sym.number]
	if already {
		Q("BindSymbol already sees symbol %v, currently bound to '%v'", sym.name, cur.SexpString())

		lhsTy := cur.Type()
		rhsTy := expr.Type()
		if lhsTy == nil {
			// for backcompat with closure.zy, just do the binding for now if the LHS isn't typed.
			//return fmt.Errorf("left-hand-side had nil type")
			// TODO: fix this? or require removal of previous symbol binding to avoid type errors?
			stack.elements[stack.tos].(*Scope).Map[sym.number] = expr
			return nil
		}
		if rhsTy == nil {
			return fmt.Errorf("right-hand-side had nil type")
		}

		// both sides have type
		Q("BindSymbol: both sides have type. rhs=%v, lhs=%v", rhsTy.SexpString(), lhsTy.SexpString())

		if lhsTy == rhsTy {
			Q("BindSymbol: YES types match exactly. Good.")
			stack.elements[stack.tos].(*Scope).Map[sym.number] = expr
			return nil
		}

		if rhsTy.UserStructDefn != nil && rhsTy.UserStructDefn != lhsTy.UserStructDefn {
			return fmt.Errorf("cannot assign %v to %v", rhsTy.ShortName(), lhsTy.ShortName())
		}

		if lhsTy.UserStructDefn != nil && lhsTy.UserStructDefn != rhsTy.UserStructDefn {
			return fmt.Errorf("cannot assign %v to %v", rhsTy.ShortName(), lhsTy.ShortName())
		}

		// TODO: problem with this implementation is that it may narrow the possible
		// types assignments to this variable. To fix we'll need to keep around the
		// type of the symbol in the symbol table, separately from the value currently
		// bound to it.
		if lhsTy.TypeCache != nil && rhsTy.TypeCache != nil {
			if rhsTy.TypeCache.AssignableTo(lhsTy.TypeCache) {
				Q("BindSymbol: YES: rhsTy.TypeCache (%v) is AssigntableTo(lhsTy.TypeCache) (%v). Good.", rhsTy.TypeCache, lhsTy.TypeCache)
				stack.elements[stack.tos].(*Scope).Map[sym.number] = expr
				return nil
			}
		}
		Q("BindSymbol: at end, defaulting to deny")
		return fmt.Errorf("cannot assign %v to %v", rhsTy.ShortName(), lhsTy.ShortName())
	} else {
		Q("BindSymbol: new symbol %v", sym.name)
	}
	stack.elements[stack.tos].(*Scope).Map[sym.number] = expr
	return nil
}

func (stack *Stack) DeleteSymbolFromTopOfStackScope(sym SexpSymbol) error {
	if stack.IsEmpty() {
		panic("empty stack!!")
		//return errors.New("no scope available")
	}
	_, present := stack.elements[stack.tos].(*Scope).Map[sym.number]
	if !present {
		return fmt.Errorf("symbol `%s` not found", sym.name)
	}
	delete(stack.elements[stack.tos].(*Scope).Map, sym.number)
	return nil
}

// used to implement (set v 10)
func (scope *Scope) UpdateSymbolInScope(sym SexpSymbol, expr Sexp) error {

	_, found := scope.Map[sym.number]
	if !found {
		return fmt.Errorf("symbol `%s` not found", sym.name)
	}
	scope.Map[sym.number] = expr
	return nil
}

func (scope *Scope) DeleteSymbolInScope(sym SexpSymbol) error {

	_, found := scope.Map[sym.number]
	if !found {
		return fmt.Errorf("symbol `%s` not found", sym.name)
	}
	delete(scope.Map, sym.number)
	return nil
}

type SymtabE struct {
	Key string
	Val string
}

type SymtabSorter []*SymtabE

func (a SymtabSorter) Len() int           { return len(a) }
func (a SymtabSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SymtabSorter) Less(i, j int) bool { return a[i].Key < a[j].Key }

func (scop Scope) Show(env *Glisp, indent int, label string) (s string, err error) {
	rep := strings.Repeat(" ", indent)
	rep4 := strings.Repeat(" ", indent+4)
	s += fmt.Sprintf("%s %s  %s\n", rep, label, scop.Name)
	if scop.IsGlobal && !env.showGlobalScope {
		s += fmt.Sprintf("%s (global scope - omitting content for brevity)\n", rep4)
		return
	}
	if len(scop.Map) == 0 {
		s += fmt.Sprintf("%s empty-scope: no symbols\n", rep4)
		return
	}
	sortme := []*SymtabE{}
	for symbolNumber, val := range scop.Map {
		symbolName := env.revsymtable[symbolNumber]
		sortme = append(sortme, &SymtabE{Key: symbolName, Val: val.SexpString()})
	}
	sort.Sort(SymtabSorter(sortme))
	for i := range sortme {
		s += fmt.Sprintf("%s %s -> %s\n", rep4,
			sortme[i].Key, sortme[i].Val)
	}
	return
}

type Showable interface {
	Show(env *Glisp, indent int, label string) (string, error)
}
