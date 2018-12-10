package zygo

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Scopes map names to values. Scope nesting avoids variable name collisions and
// allows namespace maintainance. Most scopes (inside loops, inside functions)
// are implicitly created. Packages are scopes that the user can manipulate
// explicitly.
type Scope struct {
	Map         map[int]Sexp
	IsGlobal    bool
	Name        string
	PackageName string
	Parent      *Scope
	IsFunction  bool          // if true, read-only.
	MyFunction  *SexpFunction // so we can query captured closure scopes.
	IsPackage   bool
	env         *Zlisp
}

// SexpString satisfies the Sexp interface, producing a string presentation of the value.
func (s *Scope) SexpString(ps *PrintState) string {
	var label string
	head := ""
	if s.IsPackage {
		head = "(package " + s.PackageName
	} else {
		label = "scope " + s.Name
		if s.IsGlobal {
			label += " (global)"
		}
	}

	str, err := s.Show(s.env, ps, s.Name)
	if err != nil {
		return "(" + label + ")"
	}

	return head + " " + str + " )"
}

// Type() satisfies the Sexp interface, returning the type of the value.
func (s *Scope) Type() *RegisteredType {
	return GoStructRegistry.Lookup("packageScope")
}

func (env *Zlisp) NewScope() *Scope {
	return &Scope{
		Map: make(map[int]Sexp),
		env: env,
	}
}

func (env *Zlisp) NewNamedScope(name string) *Scope {
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
// If setVal is not nil, and if we find the symbol, we set it in the scope
// where it was found. This is equivalent to scope.UpdateSymbolInScope.
//
func (stack *Stack) lookupSymbol(sym *SexpSymbol, minFrame int, setVal *Sexp) (Sexp, error, *Scope) {
	if !stack.IsEmpty() {
		for i := 0; i <= stack.tos-minFrame; i++ {
			//P("lookupSymbol checking stack %v of %v", i, (stack.tos-minFrame)+1)
			elem, err := stack.Get(i)
			if err != nil {
				//P("lookupSymbol bailing (early?) at i=%v on err='%v'", i, err)
				return SexpNull, err, nil
			}
			switch scope := elem.(type) {
			case (*Scope):
				expr, ok := scope.Map[sym.number]
				if ok {
					//P("lookupSymbol at stack scope# i=%v, we found sym '%s' with value '%s'", i, sym.name, expr.SexpString(0))
					if setVal != nil {
						scope.Map[sym.number] = *setVal
					}
					return expr, nil, scope
				}
			}
		}
	}
	//P("lookupSymbol finished stack scan without finding it")
	if stack.env != nil && stack.env.debugSymbolNotFound {
		stack.env.ShowStackStackAndScopeStack()
	}
	return SexpNull, fmt.Errorf("alas, symbol `%s` not found", sym.name), nil
}

func (stack *Stack) LookupSymbol(sym *SexpSymbol, setVal *Sexp) (Sexp, error, *Scope) {
	return stack.lookupSymbol(sym, 0, setVal)
}

// LookupSymbolNonGlobal  - closures use this to only find symbols below the global scope, to avoid copying globals it'll always be-able to ref
func (stack *Stack) LookupSymbolNonGlobal(sym *SexpSymbol) (Sexp, error, *Scope) {
	return stack.lookupSymbol(sym, 1, nil)
}

var SymNotFound = errors.New("symbol not found")

// lookup symbols, but don't go beyond a function boundary -- a user-defined
// function boundary that is. We certainly have to go up beyond
// all built-in operators like '+' and '-', '*' and '/'.
func (stack *Stack) LookupSymbolUntilFunction(sym *SexpSymbol, setVal *Sexp) (Sexp, error, *Scope) {

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
					if setVal != nil {
						scope.UpdateSymbolInScope(sym, *setVal)
					}
					return expr, nil, scope
				}
				if scope.IsFunction {
					//P("   ...scope '%s' was a function, halting up search and checking captured closures\n", scope.Name)

					// check the parent function, if avail.
					if scope.MyFunction.parent != nil {
						//P("checking non-nil parent...")
						exp, err, whichScope := scope.MyFunction.parent.ClosingLookupSymbol(sym, setVal)
						switch err {
						case nil:
							P("LookupSymbolUntilFunction('%s') found in parent scope '%s'\n", sym.name, whichScope.Name)
							return exp, err, whichScope
						}
					} else {
						//P("parent of '%s' was nil", scope.MyFunction.name)
					}

					// then check the captured closure scope stack

					exp, err, whichScope := scope.MyFunction.ClosingLookupSymbol(sym, setVal)
					switch err {
					case nil:
						//P("LookupSymbolUntilFunction('%s') found in scope '%s'\n", sym.name, whichScope.Name)
						return exp, err, whichScope
					case SymNotFound:
						//P("LookupSymbolUntilFunction('%s') not found in scope '%s'\n", sym.name, whichScope.Name)
						break doneSearching
					default:
						//P("unrecognized error '%v'", err)
						break doneSearching
					}

					// no luck inside the captured closure scopes.
					// unreachable: break doneSearching
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

func (stack *Stack) BindSymbol(sym *SexpSymbol, expr Sexp) error {
	if stack.IsEmpty() {
		panic("empty stack!!")
	}
	cur, already := stack.elements[stack.tos].(*Scope).Map[sym.number]
	if already {
		Q("BindSymbol already sees symbol %v, currently bound to '%v'", sym.name, cur.SexpString(nil))

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
			// meh, we need to be able to assign nil to stuff without freaking out,
			// so force type match
			rhsTy = lhsTy

			//return fmt.Errorf("right-hand-side had nil type back from Type() call; val = '%s'/%T", expr.SexpString(nil), expr)
		}

		// both sides have type
		Q("BindSymbol: both sides have type. rhs=%v, lhs=%v", rhsTy.SexpString(nil), lhsTy.SexpString(nil))

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

func (stack *Stack) DeleteSymbolFromTopOfStackScope(sym *SexpSymbol) error {
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
func (scope *Scope) UpdateSymbolInScope(sym *SexpSymbol, expr Sexp) error {

	_, found := scope.Map[sym.number]
	if !found {
		return fmt.Errorf("symbol `%s` not found", sym.name)
	}
	scope.Map[sym.number] = expr
	return nil
}

func (scope *Scope) DeleteSymbolInScope(sym *SexpSymbol) error {

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

func (scop *Scope) Show(env *Zlisp, ps *PrintState, label string) (s string, err error) {
	//P("scop %p Show() starting, PackageName: '%s'  IsGlobal: %v", scop, scop.PackageName, scop.IsGlobal)
	if ps == nil {
		ps = NewPrintState()
	}
	if ps.GetSeen(scop) {
		// This check is critical to prevent infinite looping in a cycle.
		// Scopes like global are referenced by every package, and
		// nested scopes refer to their paranets, so nesting
		// two packages will loop forever without this check.

		// debug version: return fmt.Sprintf("already-saw Scope %p with scop.PackageName='%s'\n", scop, scop.PackageName), nil
		return "", nil
	} else {
		ps.SetSeen(scop, "Scope")
	}
	indent := ps.GetIndent()
	rep := strings.Repeat(" ", indent)
	rep4 := strings.Repeat(" ", indent+4)
	s += fmt.Sprintf("%s %s  %s (%p)\n", rep, label, scop.Name, scop)
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
		sortme = append(sortme, &SymtabE{Key: symbolName, Val: val.SexpString(ps)})
	}
	sort.Sort(SymtabSorter(sortme))
	for i := range sortme {
		s += fmt.Sprintf("%s %s -> %s\n", rep4,
			sortme[i].Key, sortme[i].Val)
	}
	return
}

type Showable interface {
	Show(env *Zlisp, ps *PrintState, label string) (string, error)
}
