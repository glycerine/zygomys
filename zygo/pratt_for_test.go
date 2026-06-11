package zygo

import (
	"regexp"
	"strings"
	"testing"
)

func prattForEnv(t *testing.T) *Zlisp {
	t.Helper()

	env := NewZlisp()
	env.StandardSetup()
	t.Cleanup(func() { env.Close() })
	return env
}

func infixExpandString(t *testing.T, env *Zlisp, src string) string {
	t.Helper()

	got := recentEval(t, env, "(infixExpand {"+src+"})")
	return normalizePrattForGeneratedSymbols(got.SexpString(nil))
}

func normalizePrattForGeneratedSymbols(s string) string {
	replacements := []struct {
		pattern string
		name    string
	}{
		{`__range_src[0-9]+`, "__range_srcN"},
		{`__range_len[0-9]+`, "__range_lenN"},
		{`__range_i[0-9]+`, "__range_iN"},
		{`__range_pair[0-9]+`, "__range_pairN"},
	}
	for _, repl := range replacements {
		re := regexp.MustCompile(repl.pattern)
		s = re.ReplaceAllString(s, repl.name)
	}
	return s
}

func TestPrattGoForExpansion(t *testing.T) {
	env := prattForEnv(t)

	tests := []struct {
		src  string
		want string
	}{
		{
			src:  `for {}`,
			want: `(quote (for [nil true nil]))`,
		},
		{
			src:  `for i < 3 { i++ }`,
			want: `(quote (for [nil (< i 3) nil] (infix [i ++])))`,
		},
		{
			src:  `for i := 0; i < 3; i++ { sum += i }`,
			want: `(quote (for [(set i 0) (< i 3) (++ i)] (infix [sum += i])))`,
		},
		{
			src:  `for ; ; { break }`,
			want: `(quote (for [nil true nil] (infix [break])))`,
		},
		{
			src:  `for k := range n { sum += k }`,
			want: `(quote (letseq [__range_srcN n __range_lenN (__rangeLen __range_srcN)] (for [(def __range_iN 0) (< __range_iN __range_lenN) (set __range_iN (+ __range_iN 1))] (def k (__rangeKey __range_srcN __range_iN)) (infix [sum += k]))))`,
		},
		{
			src:  `for k, v = range h { sum += v }`,
			want: `(quote (letseq [__range_srcN h __range_lenN (__rangeLen __range_srcN)] (for [(def __range_iN 0) (< __range_iN __range_lenN) (set __range_iN (+ __range_iN 1))] (let [__range_pairN (__rangePair __range_srcN __range_iN)] (begin (set k (first __range_pairN)) (set v (second __range_pairN)) (infix [sum += v]))))))`,
		},
	}

	for _, tc := range tests {
		if got := infixExpandString(t, env, tc.src); got != tc.want {
			t.Fatalf("infixExpand {%s}\n got: %s\nwant: %s", tc.src, got, tc.want)
		}
	}
}

func TestPrattGoForEval(t *testing.T) {
	env := prattForEnv(t)

	cases := []struct {
		name string
		code string
		want int64
	}{
		{
			name: "forever with break",
			code: `(begin
				(def i 0)
				{for { i++; if i == 4 { break } }}
				i)`,
			want: 4,
		},
		{
			name: "while condition",
			code: `(begin
				(def i 0)
				{for i < 5 { i++ }}
				i)`,
			want: 5,
		},
		{
			name: "three clause fresh",
			code: `(begin
				(def sum 0)
				{for i := 0; i < 5; i++ { sum += i }}
				sum)`,
			want: 10,
		},
		{
			name: "three clause set",
			code: `(begin
				(def sum 0)
				(def i 0)
				{for i = 0; i < 5; i++ { sum += i }}
				sum)`,
			want: 10,
		},
		{
			name: "array one variable",
			code: `(begin
				(def sum 0)
				(def a [9 8 7])
				{for i := range a { sum += i }}
				sum)`,
			want: 3,
		},
		{
			name: "array two variables",
			code: `(begin
				(def sum 0)
				(def a [9 8 7])
				{for i, v := range a { sum += (+ i v) }}
				sum)`,
			want: 27,
		},
		{
			name: "hash one variable",
			code: `(begin
				(def s 0)
				(def h (hash a:10 b:20))
				{for k := range h { if k == a: { s += 1 } else { s += 10 } }}
				s)`,
			want: 11,
		},
		{
			name: "hash two variables",
			code: `(begin
				(def sum 0)
				(def h (hash a:10 b:20))
				{for k, v := range h { sum += v }}
				sum)`,
			want: 30,
		},
		{
			name: "integer range",
			code: `(begin
				(def sum 0)
				{for k := range 5 { sum += k }}
				sum)`,
			want: 10,
		},
		{
			name: "negative integer range",
			code: `(begin
				(def sum 0)
				{for k := range -3 { sum += k }}
				sum)`,
			want: 0,
		},
		{
			name: "continue and break",
			code: `(begin
				(def sum 0)
				{for i := 0; i < 8; i++ {
					if i == 2 { continue }
					if i == 6 { break }
					sum += i
				}}
				sum)`,
			want: 13,
		},
		{
			name: "range set",
			code: `(begin
				(def k 0)
				(def v 0)
				(def sum 0)
				(def a [4 5])
				{for k, v = range a { sum += (+ k v) }}
				sum)`,
			want: 10,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := recentInt(t, recentEval(t, env, tc.code)); got != tc.want {
				t.Fatalf("%s got %d, want %d", tc.name, got, tc.want)
			}
		})
	}
}

func TestPrattGoForErrors(t *testing.T) {
	env := prattForEnv(t)

	cases := []struct {
		name string
		code string
		want string
	}{
		{
			name: "missing body",
			code: `(infixExpand {for i < 3})`,
			want: "missing body",
		},
		{
			name: "bad semicolon count",
			code: `(infixExpand {for i := 0; i < 3; i++; i++ {}})`,
			want: "three-clause for header",
		},
		{
			name: "malformed range lhs",
			code: `(infixExpand {for k, v, z := range h {}})`,
			want: "range header",
		},
		{
			name: "integer two variable range",
			code: `(begin {for k, v := range 3 {}})`,
			want: "two-variable range over integer",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := env.EvalString(tc.code)
			if err == nil {
				t.Fatalf("%s succeeded, want error containing %q", tc.name, tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%s error = %q, want substring %q", tc.name, err.Error(), tc.want)
			}
		})
	}
}
