package glisp

func IsArray(expr Sexp) bool {
	switch expr.(type) {
	case SexpArray:
		return true
	}
	return false
}

func IsList(expr Sexp) bool {
	if expr == SexpNull {
		return true
	}
	switch list := expr.(type) {
	case SexpPair:
		return IsList(list.tail)
	}
	return false
}

func IsFloat(expr Sexp) bool {
	switch expr.(type) {
	case SexpFloat:
		return true
	}
	return false
}

func IsInt(expr Sexp) bool {
	switch expr.(type) {
	case SexpInt:
		return true
	}
	return false
}

func IsChar(expr Sexp) bool {
	switch expr.(type) {
	case SexpChar:
		return true
	}
	return false
}

func IsNumber(expr Sexp) bool {
	switch expr.(type) {
	case SexpFloat:
		return true
	case SexpInt:
		return true
	case SexpChar:
		return true
	}
	return false
}

func IsSymbol(expr Sexp) bool {
	switch expr.(type) {
	case SexpSymbol:
		return true
	}
	return false
}
