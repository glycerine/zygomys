package zygo

import (
	"bytes"
	"strings"
	"testing"
)

func evalUnderflowReplInput(t *testing.T, env *Zlisp, line string) (Sexp, error) {
	t.Helper()

	env.parser.ResetAddNewInput(bytes.NewBuffer([]byte(line + "\n")))
	exprs, err := env.parser.ParseTokens()
	if err != nil {
		return SexpNull, err
	}
	if len(exprs) == 0 {
		return env.EvalString(env.ReplLineInfixWrap(line) + " ")
	}

	infixExpr := MakeList([]Sexp{
		env.MakeSymbol("infix"),
		&SexpArray{Val: exprs, Env: env},
	})
	return env.EvalExpressions([]Sexp{infixExpr})
}

func TestUnderflowAfterSourceThenInfixSymbol(t *testing.T) {
	env := NewZlisp()
	defer env.Close()
	env.StandardSetup()

	if _, err := evalUnderflowReplInput(t, env, `(source "../tests/set.zy")`); err != nil {
		t.Fatalf("source set.zy failed: %v", err)
	}

	res, err := evalUnderflowReplInput(t, env, `x`)
	if err != nil {
		if strings.Contains(err.Error(), "we've shrunk the datastack during eval") {
			t.Fatalf("infix lookup after source corrupted the data stack: %v", err)
		}
		t.Fatalf("infix lookup after source failed: %v", err)
	}

	x, ok := res.(*SexpInt)
	if !ok {
		t.Fatalf("{x} returned %T/%v, want int 2", res, res.SexpString(nil))
	}
	if x.Val != 2 {
		t.Fatalf("{x} returned %d, want 2", x.Val)
	}
}
