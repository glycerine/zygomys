package zygo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// all Sexp are typed, and have a zero value corresponding to
// the type of the Sexp.

// Sexp is the central interface for all
// S-expressions (Symbol expressions ala lisp).
type Sexp interface {
	// SexpString: produce a string from our value.
	// Single-line strings can ignore indent.
	// Only multiline strings should follow every
	// newline with at least indent worth of spaces.
	SexpString(ps *PrintState) string

	// Type returns the type of the value.
	Type() *RegisteredType
}

type SexpPair struct {
	Head Sexp
	Tail Sexp
}

type SexpPointer struct {
	ReflectTarget reflect.Value
	Target        Sexp
	PointedToType *RegisteredType
	MyType        *RegisteredType
}

func NewSexpPointer(pointedTo Sexp) *SexpPointer {
	pointedToType := pointedTo.Type()

	var reftarg reflect.Value

	//Q("NewSexpPointer sees pointedTo of '%#v'", pointedTo)
	switch e := pointedTo.(type) {
	case *SexpReflect:
		//Q("SexpReflect.Val = '%#v'", e.Val)
		reftarg = e.Val
	default:
		reftarg = reflect.ValueOf(pointedTo)
	}

	ptrRt := GoStructRegistry.GetOrCreatePointerType(pointedToType)
	//Q("pointer type is ptrRt = '%#v'", ptrRt)
	p := &SexpPointer{
		ReflectTarget: reftarg,
		Target:        pointedTo,
		PointedToType: pointedToType,
		MyType:        ptrRt,
	}
	return p
}

func (p *SexpPointer) SexpString(ps *PrintState) string {
	return fmt.Sprintf("%p", &p.Target)
	//return fmt.Sprintf("(* %v) %p", p.PointedToType.RegisteredName, p.Target)
}

func (p *SexpPointer) Type() *RegisteredType {
	return p.MyType
}

type SexpInt struct {
	Val int64
	Typ *RegisteredType
}
type SexpUint64 struct {
	Val uint64
	Typ *RegisteredType
}
type SexpBool struct {
	Val bool
	Typ *RegisteredType
}
type SexpFloat struct {
	Val        float64
	Typ        *RegisteredType
	Scientific bool
}
type SexpChar struct {
	Val rune
	Typ *RegisteredType
}
type SexpStr struct {
	S        string
	backtick bool
	Typ      *RegisteredType
}

func (r SexpStr) Type() *RegisteredType {
	return GoStructRegistry.Registry["string"]
}

func (r *SexpInt) Type() *RegisteredType {
	return GoStructRegistry.Registry["int64"]
}

func (r *SexpUint64) Type() *RegisteredType {
	return GoStructRegistry.Registry["uint64"]
}

func (r *SexpFloat) Type() *RegisteredType {
	return GoStructRegistry.Registry["float64"]
}

func (r *SexpBool) Type() *RegisteredType {
	return GoStructRegistry.Registry["bool"]
}

func (r *SexpChar) Type() *RegisteredType {
	return GoStructRegistry.Registry["int32"]
}

func (r *RegisteredType) Type() *RegisteredType {
	return r
}

type SexpRaw struct {
	Val    []byte
	Typ    *RegisteredType
	Base64 bool
}

func (r *SexpRaw) Type() *RegisteredType {
	return r.Typ
}

type SexpReflect struct {
	Val reflect.Value
}

func (r *SexpReflect) Type() *RegisteredType {
	k := reflectName(reflect.Value(r.Val))
	//Q("SexpReflect.Type() looking up type named '%s'", k)
	ty, ok := GoStructRegistry.Registry[k]
	if !ok {
		//Q("SexpReflect.Type(): type named '%s' not found", k)
		return nil
	}
	//Q("SexpReflect.Type(): type named '%s' found as regtype '%v'", k, ty.SexpString(nil))
	return ty
}

type SexpError struct {
	error
}

func (r *SexpError) Type() *RegisteredType {
	return GoStructRegistry.Registry["error"]
}

func (r *SexpSentinel) Type() *RegisteredType {
	return nil // TODO what should this be?
}

type SexpClosureEnv Scope

func (r *SexpClosureEnv) Type() *RegisteredType {
	return nil // TODO what should this be?
}

func (c *SexpClosureEnv) SexpString(ps *PrintState) string {
	scop := (*Scope)(c)
	s, err := scop.Show(scop.env, ps, "")
	if err != nil {
		panic(err)
	}
	return s
}

type SexpSentinel struct {
	Val int
}

// these are values now so that they also have addresses.
var SexpNull = &SexpSentinel{Val: 0}
var SexpEnd = &SexpSentinel{Val: 1}
var SexpMarker = &SexpSentinel{Val: 2}

type SexpSemicolon struct{}
type SexpComma struct{}

func (r *SexpSemicolon) Type() *RegisteredType {
	return nil // TODO what should this be?
}

func (s *SexpSemicolon) SexpString(ps *PrintState) string {
	return ";"
}

func (r *SexpComma) Type() *RegisteredType {
	return nil // TODO what should this be?
}

func (s *SexpComma) SexpString(ps *PrintState) string {
	return ","
}

func (sent *SexpSentinel) SexpString(ps *PrintState) string {
	if sent == SexpNull {
		return "nil"
	}
	if sent == SexpEnd {
		return "End"
	}
	if sent == SexpMarker {
		return "Marker"
	}

	return ""
}

func Cons(a Sexp, b Sexp) *SexpPair {
	return &SexpPair{a, b}
}

func (pair *SexpPair) SexpString(ps *PrintState) string {
	str := "("

	for {
		switch pair.Tail.(type) {
		case *SexpPair:
			str += pair.Head.SexpString(ps) + " "
			pair = pair.Tail.(*SexpPair)
			continue
		}
		break
	}

	str += pair.Head.SexpString(ps)

	if pair.Tail == SexpNull {
		str += ")"
	} else {
		str += " \\ " + pair.Tail.SexpString(ps) + ")"
	}

	return str
}
func (r *SexpPair) Type() *RegisteredType {
	return nil // TODO what should this be?
}

type SexpArray struct {
	Val []Sexp

	Typ *RegisteredType

	IsFuncDeclTypeArray bool
	Infix               bool

	Env *Zlisp
}

func (r *SexpArray) Type() *RegisteredType {
	if r.Typ == nil {
		if len(r.Val) > 0 {
			// take type from first element
			ty := r.Val[0].Type()
			if ty != nil {
				r.Typ = GoStructRegistry.GetOrCreateSliceType(ty)
			}
		} else {
			// empty array
			r.Typ = GoStructRegistry.Lookup("[]")
			//P("lookup [] returned type %#v", r.Typ)
		}
	}
	return r.Typ
}

func (arr *SexpArray) SexpString(ps *PrintState) string {
	indInner := ""
	indent := ps.GetIndent()
	innerPs := ps.AddIndent(4) // generates a fresh new PrintState
	inner := indent + 4
	//prettyEnd := ""
	pretty := false
	if arr != nil && arr.Env != nil && arr.Env.Pretty {
		pretty = true
		//prettyEnd = "\n"
		indInner = strings.Repeat(" ", inner)
		ps = innerPs
	}

	comma := ""
	if ps != nil && ps.PrintJSON {
		comma = "," // be valid JSON
	}

	opn := "["
	cls := "]"

	if arr.Infix {
		opn = "{"
		cls = "}"
	}
	if pretty {
		opn += "\n"
		indInner = strings.Repeat(" ", inner)
	}

	n := len(arr.Val)
	if n == 0 {
		return opn + cls
	}
	ta := arr.IsFuncDeclTypeArray
	str := opn

	last := n - 1
	var comma2 string
	for i, sexp := range arr.Val {
		str += indInner + sexp.SexpString(ps)
		comma2 = comma
		if i == last {
			comma2 = ""
		}
		if ta {
			if i%2 == 0 {
				str += ":"
			} else {
				str += comma2 + " "
			}
		} else {
			if pretty {
				str += comma2 + "\n"
			} else {
				str += comma2 + " "
			}
		}
	}
	m := len(str)
	if pretty {
		str = str[:m-1] + "\n" + indInner + cls
	} else {
		str = str[:m-1] + indInner + cls
	}
	return str
}

func (e *SexpError) SexpString(ps *PrintState) string {
	return e.error.Error()
}

type EmbedPath struct {
	ChildName     string
	ChildFieldNum int
}

func GetEmbedPath(e []EmbedPath) string {
	r := ""
	last := len(e) - 1
	for i, s := range e {
		r += s.ChildName
		if i < last {
			r += ":"
		}
	}
	return r
}

type HashFieldDet struct {
	FieldNum     int
	FieldType    reflect.Type
	StructField  reflect.StructField
	FieldName    string
	FieldJsonTag string
	EmbedPath    []EmbedPath // we are embedded if len(EmbedPath) > 0
}
type SexpHash struct {
	TypeName         string
	Map              map[int][]*SexpPair
	KeyOrder         []Sexp
	GoStructFactory  *RegisteredType
	NumKeys          int
	GoMethods        []reflect.Method
	GoFields         []reflect.StructField
	GoMethSx         SexpArray
	GoFieldSx        SexpArray
	GoType           reflect.Type
	NumMethod        int
	GoShadowStruct   interface{}
	GoShadowStructVa reflect.Value
	ShadowSet        bool

	// json tag name -> pointers to example values, as factories for SexpToGoStructs()
	JsonTagMap map[string]*HashFieldDet
	DetOrder   []*HashFieldDet

	// for using these as a scoping model
	DefnEnv    *SexpHash
	SuperClass *SexpHash
	ZMain      SexpFunction
	ZMethods   map[string]*SexpFunction
	Env        *Zlisp
}

var MethodNotFound = fmt.Errorf("method not found")

func (h *SexpHash) RunZmethod(method string, args []Sexp) (Sexp, error) {
	f, ok := (h.ZMethods)[method]
	if !ok {
		return SexpNull, MethodNotFound
	}

	panic(fmt.Errorf("not done calling %s", f.name))
	//return SexpNull, nil
}

func CallZMethodOnRecordFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	narg := len(args)
	if narg < 2 {
		return SexpNull, WrongNargs
	}
	var hash *SexpHash
	switch h := args[0].(type) {
	case *SexpHash:
		hash = h
	default:
		return SexpNull, fmt.Errorf("can only _call on a record")
	}

	method := ""
	switch s := args[1].(type) {
	case *SexpSymbol:
		method = s.name
	case *SexpStr:
		method = s.S
	default:
		return SexpNull, fmt.Errorf("can only _call with a " +
			"symbol or string as the method name. example: (_call record method:)")
	}

	return hash.RunZmethod(method, args[2:])
}

func (h *SexpHash) SetMain(p *SexpFunction) {
	h.BindSymbol(h.Env.MakeSymbol(".main"), p)
}

func (h *SexpHash) SetDefnEnv(p *SexpHash) {
	h.DefnEnv = p
	h.BindSymbol(h.Env.MakeSymbol(".parent"), p)
}

func (h *SexpHash) Lookup(env *Zlisp, key Sexp) (expr Sexp, err error) {
	return h.HashGet(env, key)
}

func (h *SexpHash) BindSymbol(key *SexpSymbol, val Sexp) error {
	return h.HashSet(key, val)
}

func (h *SexpHash) SetGoStructFactory(factory *RegisteredType) {
	h.GoStructFactory = factory
}

var SexpIntSize = 64
var SexpFloatSize = 64

func (r *SexpReflect) SexpString(ps *PrintState) string {
	//Q("in SexpReflect.SexpString(indent); top; type = %T", r)
	if reflect.Value(r.Val).Type().Kind() == reflect.Ptr {
		iface := reflect.Value(r.Val).Interface()
		switch iface.(type) {
		case *string:
			return fmt.Sprintf("`%v`", reflect.Value(r.Val).Elem().Interface())
		default:
			return fmt.Sprintf("%v", reflect.Value(r.Val).Elem().Interface())
		}
	}
	iface := reflect.Value(r.Val).Interface()
	//Q("in SexpReflect.SexpString(indent); type = %T", iface)
	switch iface.(type) {
	default:
		return fmt.Sprintf("%v", iface)
	}
}

func (b *SexpBool) SexpString(ps *PrintState) string {
	if bool(b.Val) {
		return "true"
	}
	return "false"
}

func (i *SexpInt) SexpString(ps *PrintState) string {
	return strconv.Itoa(int(i.Val))
}

func (i *SexpUint64) SexpString(ps *PrintState) string {
	return strconv.FormatUint(i.Val, 10) + "ULL"
}

func (f *SexpFloat) SexpString(ps *PrintState) string {
	if f.Scientific {
		return strconv.FormatFloat(f.Val, 'e', -1, SexpFloatSize)
	}
	return strconv.FormatFloat(f.Val, 'f', -1, SexpFloatSize)
}

func (c *SexpChar) SexpString(ps *PrintState) string {
	return strconv.QuoteRune(c.Val)
}

func (s *SexpStr) SexpString(ps *PrintState) string {
	if s.backtick {
		return "`" + s.S + "`"
	}
	return strconv.Quote(string(s.S))
}

//func (r *SexpRaw) SexpString(ps *PrintState) string {
//	return fmt.Sprintf("%#v", []byte(r.Val))
//}

type SexpSymbol struct {
	name      string
	number    int
	isDot     bool
	isSigil   bool
	colonTail bool
	sigil     string
}

func (sym *SexpSymbol) RHS(env *Zlisp) (Sexp, error) {
	if sym.isDot && env != nil {
		return dotGetSetHelper(env, sym.name, nil)
	}
	return sym, nil
}

func (sym *SexpSymbol) AssignToSelection(env *Zlisp, rhs Sexp) error {
	if sym.isDot && env != nil {
		_, err := dotGetSetHelper(env, sym.name, &rhs)
		return err
	}
	panic("not implemented yet")
}

func (sym *SexpSymbol) SexpString(ps *PrintState) string {
	if sym.colonTail {
		//		return sym.name + ":"
	}
	return sym.name
}

func (r *SexpSymbol) Type() *RegisteredType {
	return GoStructRegistry.Registry["symbol"]
}

func (sym SexpSymbol) Name() string {
	return sym.name
}

func (sym SexpSymbol) Number() int {
	return sym.number
}

// SexpInterfaceDecl
type SexpInterfaceDecl struct {
	name    string
	methods []*SexpFunction
}

func (r *SexpInterfaceDecl) SexpString(ps *PrintState) string {
	indent := ps.GetIndent()
	space := strings.Repeat(" ", indent)
	space4 := strings.Repeat(" ", indent+4)
	s := space + "(interface " + r.name + " ["
	if len(r.methods) > 0 {
		s += "\n"
	}
	for i := range r.methods {
		s += space4 + r.methods[i].SexpString(ps.AddIndent(4)) + "\n"
	}
	s += space + "])"
	return s
}

func (r *SexpInterfaceDecl) Type() *RegisteredType {
	// todo: how to register/what to register?
	return GoStructRegistry.Registry[r.name]
}

// SexpFunction
type SexpFunction struct {
	name              string
	user              bool
	nargs             int
	varargs           bool
	fun               ZlispFunction
	userfun           ZlispUserFunction
	orig              Sexp
	closingOverScopes *Closing
	parent            *SexpFunction
	isBuilder         bool // see defbuild; builders are builtins that receive un-evaluated expressions
	inputTypes        *SexpHash
	returnTypes       *SexpHash
	hasBody           bool // could just be declaration in an interface, without a body
}

func (sf *SexpFunction) Type() *RegisteredType {
	return nil // TODO what goes here
}

func (sf *SexpFunction) Copy() *SexpFunction {
	cp := *sf
	return &cp
}

func (sf *SexpFunction) SetClosing(clos *Closing) {
	ps4 := NewPrintStateWithIndent(4)
	pre, err := sf.ShowClosing(clos.env, ps4, "prev")
	_ = pre
	panicOn(err)
	newnew, err := sf.ShowClosing(clos.env, ps4, "newnew")
	_ = newnew
	panicOn(err)
	//P("99999 for sfun = %p, in sfun.SetClosing(), prev value is %p = '%s'\n",
	//	sf, sf.closingOverScopes, pre)
	//P("88888 in sfun.SetClosing(), new  value is %p = '%s'\n", clos, newnew)
	sf.closingOverScopes = clos
	//P("in SetClosing() for '%s'/%p: my stack is: '%s'", sf.name, sf, clos.Stack.SexpString(nil))
}

func (sf *SexpFunction) ShowClosing(env *Zlisp, ps *PrintState, label string) (string, error) {
	if sf.closingOverScopes == nil {
		return sf.name + " has no captured scopes.", nil
	}
	return sf.closingOverScopes.Show(env, ps, label)
}

func (sf *SexpFunction) ClosingLookupSymbolUntilFunction(sym *SexpSymbol) (Sexp, error, *Scope) {
	if sf.closingOverScopes != nil {
		return sf.closingOverScopes.LookupSymbolUntilFunction(sym, nil, 1, false)
	}
	return SexpNull, SymNotFound, nil
}

func (sf *SexpFunction) ClosingLookupSymbol(sym *SexpSymbol, setVal *Sexp) (Sexp, error, *Scope) {
	if sf.closingOverScopes != nil {
		return sf.closingOverScopes.LookupSymbol(sym, setVal)
	}
	//P("sf.closingOverScopes was nil, no captured scopes. sf = '%v'", sf.SexpString(nil))
	return SexpNull, SymNotFound, nil
}

// chase parent pointers up the chain and check each of their immediate closures.
func (sf *SexpFunction) LookupSymbolInParentChainOfClosures(sym *SexpSymbol, setVal *Sexp, env *Zlisp) (Sexp, error, *Scope) {

	cur := sf
	par := sf.parent
	for par != nil {
		//fmt.Printf(" parent chain: cur:%v -> parent:%v\n", cur.name, par.name)
		//fmt.Printf("        cur.closures = %s", ClosureToString(cur, env))

		exp, err, scope := cur.ClosingLookupSymbolUntilFunc(sym, setVal, 1, false)
		if err == nil {
			//P("LookupSymbolInParentChainOfClosures(sym='%s') found in scope '%s'\n", sym.name, scope.Name)
			return exp, err, scope
		}

		cur = par
		par = par.parent
	}

	return SexpNull, SymNotFound, nil
}

func (sf *SexpFunction) ClosingLookupSymbolUntilFunc(sym *SexpSymbol, setVal *Sexp, maximumFuncToSearch int, checkCaptures bool) (Sexp, error, *Scope) {
	if sf.closingOverScopes != nil {
		return sf.closingOverScopes.LookupSymbolUntilFunction(sym, setVal, maximumFuncToSearch, checkCaptures)
	}
	//P("sf.closingOverScopes was nil, no captured scopes. sf = '%v'", sf.SexpString(nil))
	return SexpNull, SymNotFound, nil
}

func (sf *SexpFunction) SexpString(ps *PrintState) string {
	if sf.orig == nil {
		return "fn [" + sf.name + "]"
	}
	return sf.orig.SexpString(ps)
}

func IsTruthy(expr Sexp) bool {
	switch e := expr.(type) {
	case *SexpBool:
		return e.Val
	case *SexpInt:
		return e.Val != 0
	case *SexpUint64:
		return e.Val != 0
	case *SexpChar:
		return e.Val != 0
	case *SexpSentinel:
		return e != SexpNull
	}
	return true
}

type SexpStackmark struct {
	sym *SexpSymbol
}

func (r *SexpStackmark) Type() *RegisteredType {
	return nil // TODO what should this be?
}

func (mark *SexpStackmark) SexpString(ps *PrintState) string {
	return "stackmark " + mark.sym.name
}
