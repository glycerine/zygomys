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

func TestDollarSigilNoLongerSelfEvaluatesWhenUnbound(t *testing.T) {
	env := newLazyTestEnv(t)

	lazyExpectEvalError(t, env, "(begin $missingLazyFormal)", "symbol `$missingLazyFormal` not found")
}
