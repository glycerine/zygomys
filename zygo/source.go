package zygo

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

// alternative. simpler, currently panics.
func SimpleSourceFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	src, isStr := args[0].(*SexpStr)
	if !isStr {
		return SexpNull, fmt.Errorf("-> error: first argument must be a string")
	}

	file := src.S
	if !FileExists(file) {
		return SexpNull, fmt.Errorf("path '%s' does not exist", file)
	}

	env2 := env.Duplicate()

	f, err := os.Open(file)
	if err != nil {
		return SexpNull, err
	}
	defer f.Close()

	err = env2.LoadFile(f)
	if err != nil {
		return SexpNull, err
	}

	_, err = env2.Run()

	return SexpNull, err
}

// existing

// SourceExpressions, this should be called from a user func context
func (env *Zlisp) SourceExpressions(expressions []Sexp) error {
	gen := NewGenerator(env)

	err := gen.GenerateBegin(expressions)
	if err != nil {
		return err
	}
	//P("debug: in SourceExpressions, FROM expressions='%s'", (&SexpArray{Val: expressions, Env: env}).SexpString(0))
	//P("debug: in SourceExpressions, gen=")
	//DumpFunction(ZlispFunction(gen.instructions), -1)
	curfunc := env.curfunc
	curpc := env.pc

	env.curfunc = env.MakeFunction("__source", 0, false,
		gen.instructions, nil)
	env.pc = 0

	result, err := env.Run()
	if err != nil {
		return err
	}

	//P("end of SourceExpressions, result going onto datastack is: '%s'", result.SexpString(0))
	env.datastack.PushExpr(result)

	//P("debug done with Run in source, now stack is:")
	//env.datastack.PrintStack()

	env.pc = curpc
	env.curfunc = curfunc

	return nil
}

func (env *Zlisp) SourceStream(stream io.RuneScanner) error {
	env.parser.ResetAddNewInput(stream)
	expressions, err := env.parser.ParseTokens()
	if err != nil {
		return errors.New(fmt.Sprintf(
			"Error parsing on line %d: %v\n", env.parser.Linenum(), err))
	}

	// like LoadExpressions in environment.go, remove comments.
	expressions = env.FilterArray(expressions, RemoveCommentsFilter)
	
	return env.SourceExpressions(expressions)
}

func (env *Zlisp) SourceFile(file *os.File) error {
	return env.SourceStream(bufio.NewReader(file))
}

func SourceFileFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}

	for _, v := range args {
		if err := env.sourceItem(v); err != nil {
			return SexpNull, err
		}
	}

	result, err := env.datastack.PopExpr()
	if err != nil {
		return SexpNull, err
	}
	return result, nil
}

// helper for SourceFileFunction recursion
func (env *Zlisp) sourceItem(item Sexp) error {
	switch t := item.(type) {
	case *SexpArray:
		for _, v := range t.Val {
			if err := env.sourceItem(v); err != nil {
				return err
			}
		}
	case *SexpPair:
		expr := item
		for expr != SexpNull {
			list := expr.(*SexpPair)
			if err := env.sourceItem(list.Head); err != nil {
				return err
			}
			expr = list.Tail
		}
	case *SexpStr:
		var f *os.File
		var err error

		if f, err = os.Open(t.S); err != nil {
			return err
		}
		defer f.Close()
		if err = env.SourceFile(f); err != nil {
			return err
		}

	default:
		return fmt.Errorf("source: Expected `string`, `list`, `array`. Instead found type %T val %v", item, item)
	}

	return nil
}
