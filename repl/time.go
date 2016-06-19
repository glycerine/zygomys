package zygo

import (
	"errors"
	"fmt"
	"time"
)

type SexpTime struct {
	Tm time.Time
}

func (r *SexpTime) Type() *RegisteredType {
	return nil // TODO what should this be?
}

func (t *SexpTime) SexpString(ps *PrintState) string {
	return t.Tm.String()
}

func NowFunction(env *Glisp, name string,
	args []Sexp) (Sexp, error) {
	return &SexpTime{Tm: time.Now()}, nil
}

func TimeitFunction(env *Glisp, name string,
	args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	var fun *SexpFunction
	switch t := args[0].(type) {
	case *SexpFunction:
		fun = t
	default:
		return SexpNull,
			errors.New("argument of timeit should be function")
	}

	starttime := time.Now()
	elapsed := time.Since(starttime)
	maxseconds := 10.0
	var iterations int

	for iterations = 0; iterations < 10000; iterations++ {
		_, err := env.Apply(fun, []Sexp{})
		if err != nil {
			return SexpNull, err
		}
		elapsed = time.Since(starttime)
		if elapsed.Seconds() > maxseconds {
			break
		}
	}

	fmt.Printf("ran %d iterations in %f seconds\n",
		iterations, elapsed.Seconds())
	fmt.Printf("average %f seconds per run\n",
		elapsed.Seconds()/float64(iterations))

	return SexpNull, nil
}

func (env *Glisp) ImportTime() {
	env.AddFunction("now", NowFunction)
	env.AddFunction("timeit", TimeitFunction)
}
