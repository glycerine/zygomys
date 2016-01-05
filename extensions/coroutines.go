package gdslext

import (
	"errors"
	"github.com/glycerine/godiesel/interpreter"
)

type SexpCoroutine struct {
	env *gdsl.Glisp
}

func (coro SexpCoroutine) SexpString() string {
	return "[coroutine]"
}

func StartCoroutineFunction(env *gdsl.Glisp, name string,
	args []gdsl.Sexp) (gdsl.Sexp, error) {
	switch t := args[0].(type) {
	case SexpCoroutine:
		go t.env.Run()
	default:
		return gdsl.SexpNull, errors.New("not a coroutine")
	}
	return gdsl.SexpNull, nil
}

func CreateCoroutineMacro(env *gdsl.Glisp, name string,
	args []gdsl.Sexp) (gdsl.Sexp, error) {
	coroenv := env.Duplicate()
	err := coroenv.LoadExpressions(args)
	if err != nil {
		return gdsl.SexpNull, nil
	}
	coro := SexpCoroutine{coroenv}

	// (apply StartCoroutineFunction [coro])
	return gdsl.MakeList([]gdsl.Sexp{env.MakeSymbol("apply"),
		gdsl.MakeUserFunction("__start", StartCoroutineFunction),
		gdsl.SexpArray([]gdsl.Sexp{coro})}), nil
}

func ImportCoroutines(env *gdsl.Glisp) {
	env.AddMacro("go", CreateCoroutineMacro)
}
