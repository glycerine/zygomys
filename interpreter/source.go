package glisp

import (
	"fmt"
	"os"
	"errors"
)

func RunScript(env *Glisp, fname string) {
	file, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	err = env.LoadFile(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	_, err = env.Run()
	
	if err != nil {
		fmt.Print(env.GetStackTrace(err))
		env.Clear()
	}
}

func SourceFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	env2 := env.Duplicate()
	
	switch t := args[0].(type) {
	case SexpStr:		
		RunScript(env2, string(t))
	default:
		return SexpNull, errors.New(
			fmt.Sprintf("argument to %s must be string: the path to source", name))
	}

	return SexpNull, nil
}

func (env *Glisp) ImportRequire() {
	env.AddMacro("require", RequireMacro)
}

// (require path) avoids the need to put quotes around path
func RequireMacro(env *Glisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) < 1 {
		return SexpNull, fmt.Errorf("path to source missing. use: "+
			"(require path-to-source)\n")
	}
	
	// (source "path")
	return MakeList([]Sexp{env.MakeSymbol("source"),
		SexpStr(args[0].(SexpSymbol).name)}), nil
}

