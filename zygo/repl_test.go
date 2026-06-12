package zygo

import (
	"bufio"
	"strings"
	"testing"
)

func TestReplMultilineInfixIfElseDoesNotAccumulatePartialParses(t *testing.T) {
	env := NewZlisp()
	defer env.Close()
	env.StandardSetup()

	input := strings.NewReader(`if false {
  result = 1
} else {
  result = 2
}
`)

	pr := &Prompter{prompt: ""}
	line, exprs, err := pr.getExpressionWithLiner(env, bufio.NewReader(input), true)
	if err != nil {
		t.Fatalf("getExpressionWithLiner failed: %v", err)
	}

	const wantLine = `if false {
  result = 1
} else {
  result = 2
}`
	if line != wantLine {
		t.Fatalf("readin line mismatch\n got: %q\nwant: %q", line, wantLine)
	}

	if len(exprs) != 5 {
		t.Fatalf("expected one final if/else parse with 5 terms, got len=%d exprs=%s",
			len(exprs), (&SexpArray{Val: exprs, Env: env}).SexpString(nil))
	}

	infix := MakeList([]Sexp{env.MakeSymbol("infix"), &SexpArray{Val: exprs, Env: env}})
	_, err = env.EvalExpressions([]Sexp{infix})
	if err != nil {
		t.Fatalf("EvalExpressions failed: %v", err)
	}

	got := evalSelectorRvalueInfix(t, env, "result")
	if goti := recentInt(t, got); goti != 2 {
		t.Fatalf("expected multiline infix if/else to run else branch, got result=%d", goti)
	}
}
