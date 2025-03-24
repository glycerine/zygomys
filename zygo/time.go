package zygo

import (
	"errors"
	"fmt"
	"time"
)

var UtcTz *time.Location
var NYC *time.Location

func init() {
	var err error
	UtcTz, err = time.LoadLocation("UTC")
	panicOn(err)
	NYC, err = time.LoadLocation("America/New_York")
	panicOn(err)
}

type SexpTime struct {
	Tm time.Time
}

func (r *SexpTime) Type() *RegisteredType {
	return nil // TODO what should this be?
}

func (t *SexpTime) SexpString(ps *PrintState) string {
	return t.Tm.String()
}

func NowFunction(env *Zlisp, name string,
	args []Sexp) (Sexp, error) {
	return &SexpTime{Tm: time.Now()}, nil
}

// string -> time.Time
func AsTmFunction(env *Zlisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	var str *SexpStr
	switch t := args[0].(type) {
	case *SexpStr:
		str = t
	default:
		return SexpNull,
			errors.New("argument of astm should be a string RFC3999Nano timestamp that we want to convert to time.Time")
	}

	tm, err := time.ParseInLocation(time.RFC3339Nano, str.S, NYC)
	if err != nil {
		return SexpNull, err
	}
	return &SexpTime{Tm: tm.In(NYC)}, nil
}

func TimeitFunction(env *Zlisp, name string,
	args []Sexp) (Sexp, error) {
	nargs := len(args)
	if nargs != 1 && nargs != 2 {
		return SexpNull, WrongNargs
	}

	var fun *SexpFunction
	switch t := args[0].(type) {
	case *SexpFunction:
		fun = t
	default:
		return SexpNull,
			errors.New("1st argument of timeit should be function")
	}

	starttime := time.Now()
	maxseconds := 10.0
	iterations := int64(1)
	if nargs == 2 {
		switch t := args[1].(type) {
		case *SexpInt:
			iterations = t.Val
		case *SexpUint64:
			iterations = int64(t.Val)
		case *SexpFloat:
			iterations = int64(t.Val)
		default:
			return SexpNull,
				fmt.Errorf("2nd argument to timeit should be the iteration count (default 1); got type '%T'", args[1])
		}
	}
	//fmt.Printf("nargs = %v; iterations = %v\n", nargs, iterations)
	for i := int64(0); i < iterations; i++ {
		_, err := env.Apply(fun, []Sexp{})
		if err != nil {
			return SexpNull, err
		}
		if nargs == 1 {
			// only limit to 10 seconds if using default iteration count.
			elapsed := time.Since(starttime)
			if elapsed.Seconds() > maxseconds {
				break
			}
		}
	}

	elapsed := time.Since(starttime)
	fmt.Printf("ran %d iterations in %f seconds\n",
		iterations, elapsed.Seconds())
	fmt.Printf("average %f seconds per run\n",
		elapsed.Seconds()/float64(iterations))

	return SexpNull, nil
}

func MillisFunction(env *Zlisp, name string,
	args []Sexp) (Sexp, error) {
	millis := time.Now().UnixNano() / 1000000
	return &SexpInt{Val: int64(millis)}, nil
}

func (env *Zlisp) ImportTime() {
	env.AddFunction("now", NowFunction)
	env.AddFunction("timeit", TimeitFunction)
	env.AddFunction("astm", AsTmFunction)
	env.AddFunction("millis", MillisFunction)
}
