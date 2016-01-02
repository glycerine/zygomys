package glisp

import (
	"fmt"
	"os"
	"errors"
)

func RunScript(env *Glisp, fname string) {
	fmt.Printf("\n RunScript() started!\n")

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
	
	fmt.Printf("\n RunScript() LoadFile() done\n")

	_, err = env.Run()
	fmt.Printf("\n RunScript() env.Run() done, err: '%v'\n", err)
	
	if err != nil {
		fmt.Print(env.GetStackTrace(err))
	}
}

func SourceFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	fmt.Printf("\n SourceFunction started!\n")
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

func (env *Glisp) ImportSource() {
	env.AddFunction("source", SourceFunction)
}
