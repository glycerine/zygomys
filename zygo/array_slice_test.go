package zygo

import "testing"

func arraySliceInts(t *testing.T, expr Sexp) []int64 {
	t.Helper()

	arr, ok := expr.(*SexpArray)
	if !ok {
		t.Fatalf("expected *SexpArray, got %T/%v", expr, expr.SexpString(nil))
	}
	got := make([]int64, len(arr.Val))
	for i, sx := range arr.Val {
		x, ok := sx.(*SexpInt)
		if !ok {
			t.Fatalf("expected slice element %d to be *SexpInt, got %T/%v", i, sx, sx.SexpString(nil))
		}
		got[i] = x.Val
	}
	return got
}

func assertIntSlice(t *testing.T, got, want []int64) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestInfixArrayGoStyleSlicing(t *testing.T) {
	env := NewZlisp()
	defer env.Close()
	env.StandardSetup()

	if _, err := env.EvalString("(def a [3 4 5])"); err != nil {
		t.Fatalf("def a failed: %v", err)
	}

	cases := []struct {
		name string
		src  string
		want []int64
	}{
		{name: "tail", src: "a[1:]", want: []int64{4, 5}},
		{name: "prefix", src: "a[:2]", want: []int64{3, 4}},
		{name: "bounded", src: "a[1:2]", want: []int64{4}},
		{name: "whole", src: "a[:]", want: []int64{3, 4, 5}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			evalSelectorRvalueInfix(t, env, "got := "+tc.src)
			got := arraySliceInts(t, evalSelectorRvalueInfix(t, env, "got"))
			assertIntSlice(t, got, tc.want)
		})
	}
}

func TestInfixArrayGoStyleSlicingWithExpressionBounds(t *testing.T) {
	env := NewZlisp()
	defer env.Close()
	env.StandardSetup()

	if _, err := env.EvalString("(def a [0 1 2 3 4 5])"); err != nil {
		t.Fatalf("def a failed: %v", err)
	}
	if _, err := env.EvalString("(def i 2)"); err != nil {
		t.Fatalf("def i failed: %v", err)
	}
	if _, err := env.EvalString("(def j 5)"); err != nil {
		t.Fatalf("def j failed: %v", err)
	}

	cases := []struct {
		name string
		src  string
		want []int64
	}{
		{name: "whole", src: "a[:]", want: []int64{0, 1, 2, 3, 4, 5}},
		{name: "variable start", src: "a[i:]", want: []int64{2, 3, 4, 5}},
		{name: "variable end", src: "a[:i]", want: []int64{0, 1}},
		{name: "variable bounds", src: "a[i:j]", want: []int64{2, 3, 4}},
		{name: "expression start", src: "a[i+1:j]", want: []int64{3, 4}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			evalSelectorRvalueInfix(t, env, "got := "+tc.src)
			got := arraySliceInts(t, evalSelectorRvalueInfix(t, env, "got"))
			assertIntSlice(t, got, tc.want)
		})
	}
}

func TestInfixArrayGoStyleSlicingWithSubtractionStartBound(t *testing.T) {
	env := NewZlisp()
	defer env.Close()
	env.StandardSetup()

	if _, err := env.EvalString("(def a [0 2 4 6 8])"); err != nil {
		t.Fatalf("def a failed: %v", err)
	}
	if _, err := env.EvalString("(def x 3)"); err != nil {
		t.Fatalf("def x failed: %v", err)
	}

	evalSelectorRvalueInfix(t, env, "got := a[x-1:x]")
	got := arraySliceInts(t, evalSelectorRvalueInfix(t, env, "got"))
	assertIntSlice(t, got, []int64{4})
}

func TestInfixHashArrayKeySelectorStillWorks(t *testing.T) {
	env := NewZlisp()
	defer env.Close()
	env.StandardSetup()

	_, err := env.EvalString(`
(def h (hash))
(hset h [0 0] %a)
(assert (== {h[0 0]} %a))
`)
	if err != nil {
		t.Fatalf("hash array-key infix selector failed: %v", err)
	}
}
