package glispext

import (
	"errors"
	"fmt"
	"github.com/zhemao/glisp/interpreter"
)

type SexpChannel chan glisp.Sexp

func (ch SexpChannel) SexpString() string {
	return "[chan]"
}

func MakeChanFunction(env *glisp.Glisp, name string,
	args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) > 1 {
		return glisp.SexpNull, glisp.WrongNargs
	}

	size := 0
	if len(args) == 1 {
		switch t := args[0].(type) {
		case glisp.SexpInt:
			size = int(t)
		default:
			return glisp.SexpNull, errors.New(
				fmt.Sprintf("argument to %s must be int", name))
		}
	}

	return SexpChannel(make(chan glisp.Sexp, size)), nil
}

func ChanTxFunction(env *glisp.Glisp, name string,
	args []glisp.Sexp) (glisp.Sexp, error) {
	if len(args) < 1 {
		return glisp.SexpNull, glisp.WrongNargs
	}
	var channel chan glisp.Sexp
	switch t := args[0].(type) {
	case SexpChannel:
		channel = chan glisp.Sexp(t)
	default:
		return glisp.SexpNull, errors.New(
			fmt.Sprintf("argument 0 of %s must be channel", name))
	}

	if name == "send!" {
		if len(args) != 2 {
			return glisp.SexpNull, glisp.WrongNargs
		}
		channel <- args[1]
		return glisp.SexpNull, nil
	}

	return <-channel, nil
}

func ImportChannels(env *glisp.Glisp) {
	env.AddFunction("make-chan", MakeChanFunction)
	env.AddFunction("send!", ChanTxFunction)
	env.AddFunction("<!", ChanTxFunction)
}
