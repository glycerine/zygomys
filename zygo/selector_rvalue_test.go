package zygo

import (
	"bytes"
	"testing"
)

func evalSelectorRvalueInfix(t *testing.T, env *Zlisp, line string) Sexp {
	t.Helper()

	env.parser.ResetAddNewInput(bytes.NewBuffer([]byte(line + "\n")))
	exprs, err := env.parser.ParseTokens()
	if err != nil {
		t.Fatalf("ParseTokens(%q) failed: %v", line, err)
	}
	if len(exprs) == 0 {
		res, err := env.EvalString(env.ReplLineInfixWrap(line) + " ")
		if err != nil {
			t.Fatalf("EvalString(%q) failed: %v", line, err)
		}
		return res
	}
	infixExpr := MakeList([]Sexp{
		env.MakeSymbol("infix"),
		&SexpArray{Val: exprs, Env: env},
	})
	res, err := env.EvalExpressions([]Sexp{infixExpr})
	if err != nil {
		t.Fatalf("EvalExpressions(%q) failed: %v", line, err)
	}
	return res
}

func selectorRvalueInt(t *testing.T, expr Sexp) int64 {
	t.Helper()

	x, ok := expr.(*SexpInt)
	if !ok {
		t.Fatalf("expected *SexpInt, got %T/%v", expr, expr.SexpString(nil))
	}
	return x.Val
}

func selectorRvalueTypeName(t *testing.T, expr Sexp) string {
	t.Helper()

	x, ok := expr.(*SexpStr)
	if !ok {
		t.Fatalf("expected *SexpStr, got %T/%v", expr, expr.SexpString(nil))
	}
	return x.S
}

func TestInfixArrayIndexRValueContextsDereferenceSelector(t *testing.T) {
	env := NewZlisp()
	defer env.Close()
	env.StandardSetup()

	if _, err := env.EvalString("(def a [0 1 2 3 4])"); err != nil {
		t.Fatalf("def a failed: %v", err)
	}

	evalSelectorRvalueInfix(t, env, "b := a[0]")

	if got := selectorRvalueTypeName(t, recentEval(t, env, "(type? b)")); got != "int64" {
		t.Fatalf("(type? b) = %q, want int64", got)
	}
	if got := selectorRvalueInt(t, evalSelectorRvalueInfix(t, env, "b")); got != 0 {
		t.Fatalf("b = %d, want 0", got)
	}

	if _, err := env.EvalString("(aset a 0 99)"); err != nil {
		t.Fatalf("aset failed: %v", err)
	}
	if got := selectorRvalueInt(t, evalSelectorRvalueInfix(t, env, "b")); got != 0 {
		t.Fatalf("b changed after mutating a[0]: got %d, want 0", got)
	}

	if got := selectorRvalueTypeName(t, recentEval(t, env, "(type? {a[0]})")); got != "int64" {
		t.Fatalf("(type? {a[0]}) = %q, want int64", got)
	}
}

func TestInfixArrayIndexLValueStillAssignsThroughSelector(t *testing.T) {
	env := NewZlisp()
	defer env.Close()
	env.StandardSetup()

	if _, err := env.EvalString("(def a [0 1 2])"); err != nil {
		t.Fatalf("def a failed: %v", err)
	}

	evalSelectorRvalueInfix(t, env, "a[0] := 42")

	got := recentEval(t, env, "(aget a 0)")
	if val := selectorRvalueInt(t, got); val != 42 {
		t.Fatalf("(aget a 0) = %d, want 42", val)
	}
}
