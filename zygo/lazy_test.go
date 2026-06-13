package zygo

import (
	"strings"
	"testing"
)

func newLazyTestEnv(t *testing.T) *Zlisp {
	t.Helper()
	env := NewZlisp()
	env.StandardSetup()
	t.Cleanup(func() {
		if err := env.Close(); err != nil {
			t.Fatalf("env.Close() failed: %v", err)
		}
	})
	return env
}

func lazyString(t *testing.T, expr Sexp) string {
	t.Helper()
	s, ok := expr.(*SexpStr)
	if !ok {
		t.Fatalf("expected *SexpStr, got %T/%v", expr, expr.SexpString(nil))
	}
	return s.S
}

func lazyExpectEvalError(t *testing.T, env *Zlisp, code, want string) {
	t.Helper()
	_, err := env.EvalString(code)
	if err == nil {
		t.Fatalf("EvalString(%q) unexpectedly succeeded", code)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("EvalString(%q) error %q did not contain %q", code, err.Error(), want)
	}
}

func TestLazyFormalDoesNotEvaluateUnusedArgument(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(defn keep [$x] 7)
(keep (stop "lazy argument should not be forced"))`)

	if got := recentInt(t, res); got != 7 {
		t.Fatalf("keep returned %d, want 7", got)
	}
}

func TestStrictCallDoesNotReceiveLazyArg(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(def n (int 7))
n`)

	if got := recentInt(t, res); got != 7 {
		t.Fatalf("strict int conversion returned %d, want 7", got)
	}
}

func TestLazyFormalOnlyBindsDollarName(t *testing.T) {
	env := newLazyTestEnv(t)

	lazyExpectEvalError(t, env, `
(defn onlyDollar [$x] x)
(onlyDollar 10)`, "symbol `x` not found")

	res := recentEval(t, env, `
(defn lazyType [$x] (type? $x))
(lazyType (+ 1 2))`)

	if got := lazyString(t, res); got != "lazyArg" {
		t.Fatalf("type? of lazy arg returned %q, want lazyArg", got)
	}
}

func TestLazyFormalSubstituteReturnsCallExpressionWithoutForcing(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(defn labelOf [$x] (str (substitute $x)))
(labelOf (stop "substitute should not force"))`)

	if got, want := lazyString(t, res), `(stop "substitute should not force")`; got != want {
		t.Fatalf("substitute string returned %q, want %q", got, want)
	}
}

func TestLazyFormalForceEvaluatesInCallerLexicalEnvironment(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(defn receiver [$x]
  (let [a 100]
    (force $x)))
(defn caller []
  (let [a 7]
    (receiver (+ a 1))))
(caller)`)

	if got := recentInt(t, res); got != 8 {
		t.Fatalf("force evaluated in wrong environment: got %d, want 8", got)
	}
}

func TestLazyFormalForceMemoizesValue(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(def n 0)
(defn bump [] (set n (+ n 1)) n)
(defn forceTwice [$x] (+ (force $x) (force $x)))
(forceTwice (bump))`)

	if got := recentInt(t, res); got != 2 {
		t.Fatalf("forceTwice returned %d, want 2", got)
	}
	n := recentEval(t, env, "(+ n 0)")
	if got := recentInt(t, n); got != 1 {
		t.Fatalf("lazy formal was forced %d times, want 1", got)
	}
}

func TestLazyFormalForcePropagatesErrorsOnlyWhenForced(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(defn captureOnly [$x] (str (substitute $x)))
(captureOnly (stop "boom only on force"))`)
	if got, want := lazyString(t, res), `(stop "boom only on force")`; got != want {
		t.Fatalf("captureOnly returned %q, want %q", got, want)
	}

	lazyExpectEvalError(t, env, `
(defn forceBoom [$x] (force $x))
(forceBoom (stop "boom on force"))`, "boom on force")
}

func TestLazyFormalCanBeMixedWithStrictFormals(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(def n 0)
(defn bump [] (set n (+ n 1)) n)
(defn mixed [a $b c] (+ a c))
(mixed (bump) (stop "middle lazy argument should not be forced") (bump))`)

	if got := recentInt(t, res); got != 3 {
		t.Fatalf("mixed returned %d, want 3", got)
	}
	n := recentEval(t, env, "(+ n 0)")
	if got := recentInt(t, n); got != 2 {
		t.Fatalf("strict argument evaluation count was %d, want 2", got)
	}
}

func TestLazyFormalWorksInAnonymousFunction(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `((fn [$x] (force $x)) (+ 10 5))`)

	if got := recentInt(t, res); got != 15 {
		t.Fatalf("anonymous lazy function returned %d, want 15", got)
	}
}

func TestLazyFormalWorksInTypedFuncDeclaration(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(func forceTyped [$x:int64] [n:int64] (force $x))
(forceTyped (+ 20 22))`)

	if got := recentInt(t, res); got != 42 {
		t.Fatalf("forceTyped returned %d, want 42", got)
	}
}

func TestLazyFormalWorksThroughSymbolAlias(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(defn keepAlias [$x] (type? $x))
(def g keepAlias)
(g (stop "aliased lazy argument should not be forced"))`)

	if got := lazyString(t, res); got != "lazyArg" {
		t.Fatalf("aliased lazy call returned type %q, want lazyArg", got)
	}
}

func TestLazyFormalWorksThroughFunctionParameter(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(defn keepParam [$x] (str (substitute $x)))
(defn caller [f] (f (stop "parameter lazy argument should not be forced")))
(caller keepParam)`)

	if got, want := lazyString(t, res), `(stop "parameter lazy argument should not be forced")`; got != want {
		t.Fatalf("parameter lazy substitute returned %q, want %q", got, want)
	}
}

func TestLazyFormalWorksThroughComputedCallee(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(defn keepComputed [$x] (str (substitute $x)))
(((fn [] keepComputed)) (stop "computed lazy argument should not be forced"))`)

	if got, want := lazyString(t, res), `(stop "computed lazy argument should not be forced")`; got != want {
		t.Fatalf("computed lazy substitute returned %q, want %q", got, want)
	}
}

func TestDynamicStrictCallDoesNotReceiveLazyArg(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(defn strictType [x] (type? x))
(def g strictType)
(g (+ 1 2))`)

	if got := lazyString(t, res); got != "int64" {
		t.Fatalf("dynamic strict call returned type %q, want int64", got)
	}
}

func TestApplyWrapsValueOriginLazyArgument(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(defn inspectApply [$x] (list (type? $x) (str (substitute $x)) (force $x)))
(apply inspectApply [42])`)

	got, err := ListToArray(res)
	if err != nil {
		t.Fatalf("apply result was not a list: %T/%v", res, res.SexpString(nil))
	}
	if len(got) != 3 {
		t.Fatalf("apply result length %d, want 3", len(got))
	}
	if typ := lazyString(t, got[0]); typ != "lazyArg" {
		t.Fatalf("apply type result %q, want lazyArg", typ)
	}
	if sub := lazyString(t, got[1]); sub != "42" {
		t.Fatalf("apply substitute result %q, want 42", sub)
	}
	if forced := recentInt(t, got[2]); forced != 42 {
		t.Fatalf("apply force result %d, want 42", forced)
	}
}

func TestMapWrapsValueOriginLazyArgument(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(defn inspectMap [$x] (str (substitute $x)))
(map inspectMap [4 5])`)

	arr, ok := res.(*SexpArray)
	if !ok {
		t.Fatalf("map result was %T/%v, want array", res, res.SexpString(nil))
	}
	if len(arr.Val) != 2 {
		t.Fatalf("map result length %d, want 2", len(arr.Val))
	}
	if got := lazyString(t, arr.Val[0]); got != "4" {
		t.Fatalf("map first substitute %q, want 4", got)
	}
	if got := lazyString(t, arr.Val[1]); got != "5" {
		t.Fatalf("map second substitute %q, want 5", got)
	}
}

func TestComputedCallEvaluatesCalleeBeforeStrictArguments(t *testing.T) {
	env := newLazyTestEnv(t)

	res := recentEval(t, env, `
(def order "")
(defn chooseStrict [] (set order (concat order "c")) (fn [x] order))
((chooseStrict) (set order (concat order "a")))`)

	if got := lazyString(t, res); got != "ca" {
		t.Fatalf("computed call order %q, want ca", got)
	}
}

func TestDollarSigilNoLongerSelfEvaluatesWhenUnbound(t *testing.T) {
	env := newLazyTestEnv(t)

	lazyExpectEvalError(t, env, "(begin $missingLazyFormal)", "symbol `$missingLazyFormal` not found")
}
