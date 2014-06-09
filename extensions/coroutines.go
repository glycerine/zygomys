package glispext

import (
	"github.com/zhemao/glisp/interpreter"
	"errors"
)

type SexpCoroutine struct {
	env *glisp.Glisp
}

func (coro SexpCoroutine) SexpString() string {
	return "[coroutine]"
}

func StartCoroutineFunction(env *glisp.Glisp, name string,
		args []glisp.Sexp) (glisp.Sexp, error) {
	switch t := args[0].(type) {
	case SexpCoroutine:
		go t.env.Run()
	default:
		return glisp.SexpNull, errors.New("not a coroutine")
	}
	return glisp.SexpNull, nil
}

func CreateCoroutineMacro(env *glisp.Glisp, name string,
		args []glisp.Sexp) (glisp.Sexp, error) {
	coroenv := env.Duplicate()
	err := coroenv.LoadExpressions(args)
	if err != nil {
		return glisp.SexpNull, nil
	}
	coro := SexpCoroutine{coroenv}

	// (apply StartCoroutineFunction [coro])
	return glisp.MakeList([]glisp.Sexp{env.MakeSymbol("apply"),
		glisp.MakeUserFunction("__start", StartCoroutineFunction),
		glisp.SexpArray([]glisp.Sexp{coro})}), nil
}

func ImportCoroutines(env *glisp.Glisp) {
	env.AddMacro("go", CreateCoroutineMacro)
}
