package zygo

import (
	"errors"
)

type SexpGoroutine struct {
	env *Glisp
}

func (goro *SexpGoroutine) SexpString() string {
	return "[coroutine]"
}
func (goro *SexpGoroutine) Type() *RegisteredType {
	return nil // TODO what goes here
}

func StartGoroutineFunction(env *Glisp, name string,
	args []Sexp) (Sexp, error) {
	switch t := args[0].(type) {
	case *SexpGoroutine:
		go t.env.Run()
	default:
		return SexpNull, errors.New("not a goroutine")
	}
	return SexpNull, nil
}

func CreateGoroutineMacro(env *Glisp, name string,
	args []Sexp) (Sexp, error) {
	goroenv := env.Duplicate()
	err := goroenv.LoadExpressions(args)
	if err != nil {
		return SexpNull, nil
	}
	goro := &SexpGoroutine{goroenv}

	// (apply StartGoroutineFunction [goro])
	return MakeList([]Sexp{env.MakeSymbol("apply"),
		MakeUserFunction("__start", StartGoroutineFunction),
		&SexpArray{Val: []Sexp{goro}}}), nil
}

func (env *Glisp) ImportGoroutines() {
	env.AddMacro("go", CreateGoroutineMacro)
}
