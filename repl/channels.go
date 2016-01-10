package zygo

import (
	"errors"
	"fmt"
)

type SexpChannel chan Sexp

func (ch SexpChannel) SexpString() string {
	return "[chan]"
}

func MakeChanFunction(env *Glisp, name string,
	args []Sexp) (Sexp, error) {
	if len(args) > 1 {
		return SexpNull, WrongNargs
	}

	size := 0
	if len(args) == 1 {
		switch t := args[0].(type) {
		case SexpInt:
			size = int(t)
		default:
			return SexpNull, errors.New(
				fmt.Sprintf("argument to %s must be int", name))
		}
	}

	return SexpChannel(make(chan Sexp, size)), nil
}

func ChanTxFunction(env *Glisp, name string,
	args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}
	var channel chan Sexp
	switch t := args[0].(type) {
	case SexpChannel:
		channel = chan Sexp(t)
	default:
		return SexpNull, errors.New(
			fmt.Sprintf("argument 0 of %s must be channel", name))
	}

	if name == "send!" {
		if len(args) != 2 {
			return SexpNull, WrongNargs
		}
		channel <- args[1]
		return SexpNull, nil
	}

	return <-channel, nil
}

func (env *Glisp) ImportChannels() {
	env.AddFunction("make-chan", MakeChanFunction)
	env.AddFunction("send!", ChanTxFunction)
	env.AddFunction("<!", ChanTxFunction)
}
