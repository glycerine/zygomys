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
