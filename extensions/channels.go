package gdslext

import (
	"errors"
	"fmt"
	"github.com/glycerine/godiesel/interpreter"
)

type SexpChannel chan gdsl.Sexp

func (ch SexpChannel) SexpString() string {
	return "[chan]"
}

func MakeChanFunction(env *gdsl.Glisp, name string,
	args []gdsl.Sexp) (gdsl.Sexp, error) {
	if len(args) > 1 {
		return gdsl.SexpNull, gdsl.WrongNargs
	}

	size := 0
	if len(args) == 1 {
		switch t := args[0].(type) {
		case gdsl.SexpInt:
			size = int(t)
		default:
			return gdsl.SexpNull, errors.New(
				fmt.Sprintf("argument to %s must be int", name))
		}
	}

	return SexpChannel(make(chan gdsl.Sexp, size)), nil
}

func ChanTxFunction(env *gdsl.Glisp, name string,
	args []gdsl.Sexp) (gdsl.Sexp, error) {
	if len(args) < 1 {
		return gdsl.SexpNull, gdsl.WrongNargs
	}
	var channel chan gdsl.Sexp
	switch t := args[0].(type) {
	case SexpChannel:
		channel = chan gdsl.Sexp(t)
	default:
		return gdsl.SexpNull, errors.New(
			fmt.Sprintf("argument 0 of %s must be channel", name))
	}

	if name == "send!" {
		if len(args) != 2 {
			return gdsl.SexpNull, gdsl.WrongNargs
		}
		channel <- args[1]
		return gdsl.SexpNull, nil
	}

	return <-channel, nil
}

func ImportChannels(env *gdsl.Glisp) {
	env.AddFunction("make-chan", MakeChanFunction)
	env.AddFunction("send!", ChanTxFunction)
	env.AddFunction("<!", ChanTxFunction)
}
