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
	case *SexpInt:
		tm := time.Unix(0, int64(t.Val)).In(NYC)
		return &SexpTime{Tm: tm}, nil
	case *SexpStr:
		str = t
	case *SexpDate:
		return &SexpTime{Tm: t.Date.ToGoTimeNYC()}, nil

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
	env.AddFunction("date", AsDateFunction)
	env.AddFunction("nextBusinessDay", NextBusinessDayFunction)
	env.AddFunction("dur", AsDurationFunction)
}

// Date

type SexpDate struct {
	Date Date
}

func (r *SexpDate) Type() *RegisteredType {
	return nil // TODO what should this be?
}

func (t *SexpDate) SexpString(ps *PrintState) string {
	return t.Date.String()
}

// string -> date
func AsDateFunction(env *Zlisp, name string,
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
			errors.New(`argument of (date "YYYY/MM/DD") constructor should be a YYYY/MM/DD string such as "2017/12/25"`)
	}

	dt, err := ParseDate(str.S, "/")
	if err != nil {
		return SexpNull, err
	}
	return &SexpDate{Date: *dt}, nil
}

func NextBusinessDayFunction(env *Zlisp, name string,
	args []Sexp) (Sexp, error) {

	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	var d *SexpDate
	switch t := args[0].(type) {
	case *SexpDate:
		d = t
	default:
		return SexpNull,
			errors.New(`argument of (nextBusinessDay) must be a date.`)
	}

	return &SexpDate{Date: *d.Date.NextBusinessDate()}, nil
}

// time.Duration

type SexpDur struct {
	Dur time.Duration
}

func (r *SexpDur) Type() *RegisteredType {
	return nil // TODO what should this be?
}

func (t *SexpDur) SexpString(ps *PrintState) string {
	//return t.Dur.String() // can give "-1ns"; hard to parse
	return fmt.Sprintf("%v", int(t.Dur))
}

// int/string -> time.Duration
func AsDurationFunction(env *Zlisp, name string,
	args []Sexp) (Sexp, error) {

	narg := len(args)
	if narg == 0 {
		// return the zero duration
		return &SexpDur{}, nil
	}

	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpInt:
		return &SexpDur{Dur: time.Duration(t.Val)}, nil
	case *SexpStr:
		dur, err := time.ParseDuration(t.S)
		panicOn(err)
		return &SexpDur{Dur: dur}, nil
	default:
		return SexpNull,
			errors.New("argument of dur must be string or int")
	}
}
