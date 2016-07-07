package zygo

import (
	"fmt"
)

// import a package, analagous to Golang.
func ImportPackageBuilder(env *Glisp, name string, args []Sexp) (Sexp, error) {
	//P("starting ImportPackageBuilder")
	n := len(args)
	if n != 1 && n != 2 {
		return SexpNull, WrongNargs
	}

	var path Sexp
	var alias string

	switch n {
	case 1:
		path = args[0]
	case 2:
		path = args[1]
		//P("import debug: alias position at args[0] is '%#v'", args[0])
		switch sy := args[0].(type) {
		case *SexpSymbol:
			//P("import debug: alias is symbol, ok: '%v'", sy.name)
			alias = sy.name
		default:
			return SexpNull, fmt.Errorf("import error: alias was not a symbol name")
		}
	}

	var pth string
	switch x := path.(type) {
	case *SexpStr:
		pth = x.S
	default:
		return SexpNull, fmt.Errorf("import error: path argument must be string")
	}
	if !FileExists(pth) {
		return SexpNull, fmt.Errorf("import error: path '%s' does not exist", pth)
	}

	pkg, err := SourceFileFunction(env, "source", []Sexp{path})
	if err != nil {
		return SexpNull, fmt.Errorf("import error: attempt to import path '%s' resulted in: '%s'", pth, err)
	}
	//P("pkg = '%#v'", pkg)

	asPkg, isPkg := pkg.(*Stack)
	if !isPkg || !asPkg.IsPackage {
		return SexpNull, fmt.Errorf("import error: attempt to import path '%s' resulted value that was not a package, but rather '%T'", pth, pkg)
	}

	if n == 1 {
		alias = asPkg.PackageName
	}
	//P("using alias = '%s'", alias)

	// now set alias in the current env
	err = env.LexicalBindSymbol(env.MakeSymbol(alias), asPkg)
	if err != nil {
		return SexpNull, err
	}

	return pkg, nil
}
