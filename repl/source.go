package zygo

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

// alternative. simpler, currently panics.
func SimpleSourceFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	src, isStr := args[0].(SexpStr)
	if !isStr {
		return SexpNull, fmt.Errorf("-> error: first argument be a string")
	}

	file := string(src)
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
func (env *Glisp) SourceExpressions(expressions []Sexp) error {
	gen := NewGenerator(env)
	// not sure what the point of this was, but the extra pop
	// off the stack is causing source to fail because it is
	// popping off our loop stack mark by mistake.
	//if !env.ReachedEnd() {
	//  gen.AddInstruction(PopInstr(0))
	//}

	//scope := env.LexicalPush("__source")
	//	defer env.LexicalPop("__source")

	err := gen.GenerateBegin(expressions)
	if err != nil {
		return err
	}

	curfunc := env.curfunc
	curpc := env.pc

	env.curfunc = env.MakeFunction("__source", 0, false,
		gen.instructions, nil)
	env.pc = 0

	env.datastack.PushExpr(SexpNull)

	if _, err = env.Run(); err != nil {
		return err
	}

	//fmt.Printf("\n debug done with Run in source, now stack is:\n")
	//env.datastack.PrintStack()
	env.datastack.PopExpr()

	env.pc = curpc
	env.curfunc = curfunc

	return nil
}

func (env *Glisp) SourceStream(stream io.RuneScanner) error {
	env.parser.ResetAddNewInput(stream)
	expressions, err := env.parser.ParseTokens()
	if err != nil {
		return errors.New(fmt.Sprintf(
			"Error parsing on line %d: %v\n", env.parser.lexer.Linenum(), err))
	}

	return env.SourceExpressions(expressions)
}

func (env *Glisp) SourceFile(file *os.File) error {
	return env.SourceStream(bufio.NewReader(file))
}

func SourceFileFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 1 {
		return SexpNull, WrongNargs
	}

	var sourceItem func(item Sexp) error

	sourceItem = func(item Sexp) error {
		switch t := item.(type) {
		case SexpArray:
			for _, v := range t {
				if err := sourceItem(v); err != nil {
					return err
				}
			}
		case SexpPair:
			expr := item
			for expr != SexpNull {
				list := expr.(SexpPair)
				if err := sourceItem(list.Head); err != nil {
					return err
				}
				expr = list.Tail
			}
		case SexpStr:
			var f *os.File
			var err error

			if f, err = os.Open(string(t)); err != nil {
				return err
			}
			defer f.Close()
			if err = env.SourceFile(f); err != nil {
				return err
			}

		default:
			return fmt.Errorf("%v: Expected `string`, `list`, `array` given type %T val %v", name, item, item)
		}

		return nil
	}

	for _, v := range args {
		if err := sourceItem(v); err != nil {
			return SexpNull, err
		}
	}

	return SexpNull, nil
}
