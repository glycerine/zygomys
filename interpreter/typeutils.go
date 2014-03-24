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

func IsHash(expr Sexp) bool {
	switch expr.(type) {
	case SexpHash:
		return true
	}
	return false
}

func IsZero(expr Sexp) bool {
	switch e := expr.(type) {
	case SexpInt:
		return int(e) == 0
	case SexpChar:
		return int(e) == 0
	case SexpFloat:
		return float64(e) == 0.0
	}
	return false
}

func IsEmpty(expr Sexp) bool {
	if expr == SexpNull {
		return true
	}

	switch e := expr.(type) {
	case SexpArray:
		return len(e) == 0
	case SexpHash:
		return HashIsEmpty(e)
	}

	return false
}
