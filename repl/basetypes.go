package zygo

var BaseTypes = []string{"byte", "defbuild", "builder", "field", "and", "or", "cond", "quote", "def", "mdef", "fn", "defn", "begin", "let", "let*", "assert", "defmac", "macexpand", "syntax-quote", "include", "for", "set", "break", "continue", "new-scope", "_ls", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "complex64", "complex128", "bool", "string", "any", "break", "case", "chan", "const", "continue", "default", "else", "defer", "fallthrough", "for", "func", "go", "goto", "if", "import", "interface", "map", "package", "range", "return", "select", "struct", "switch", "type", "var", "append", "cap", "close", "complex", "copy", "delete", "imag", "len", "make", "new", "panic", "print", "println", "real", "recover", "null", "nil"}

func (env *Glisp) ImportBaseTypes() {
	for _, e := range GoStructRegistry.Builtin {
		env.AddGlobal(e.RegisteredName, e)
	}
}
