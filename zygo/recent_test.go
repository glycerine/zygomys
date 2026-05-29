package zygo

import (
	"fmt"
	"strings"
	"testing"
)

func recentInt(t *testing.T, expr Sexp) int64 {
	t.Helper()
	x, ok := expr.(*SexpInt)
	if !ok {
		t.Fatalf("expected *SexpInt, got %T/%v", expr, expr.SexpString(nil))
	}
	return x.Val
}

func recentEval(t *testing.T, env *Zlisp, code string) Sexp {
	t.Helper()
	res, err := env.EvalString(code)
	if err != nil {
		t.Fatalf("EvalString(%q) failed: %v", code, err)
	}
	return res
}

func TestRecentUserFunctionErrorRestoresVM(t *testing.T) {
	env := NewZlisp()
	defer env.Close()

	env.AddFunction("boom", func(*Zlisp, string, []Sexp) (Sexp, error) {
		return SexpNull, fmt.Errorf("boom")
	})

	if _, err := env.EvalString("(boom)"); err == nil {
		t.Fatalf("expected boom to fail")
	}
	if env.addrstack.Size() != 0 {
		t.Fatalf("addrstack leaked after user-function error: got %d", env.addrstack.Size())
	}
	if env.curfunc != env.mainfunc {
		t.Fatalf("curfunc not restored after user-function error: got %s", env.curfunc.name)
	}

	res := recentEval(t, env, "(+ 1 2)")
	if got := recentInt(t, res); got != 3 {
		t.Fatalf("after user-function error, (+ 1 2) = %d, want 3", got)
	}
}

func TestRecentCompiledFunctionErrorRestoresVM(t *testing.T) {
	env := NewZlisp()
	defer env.Close()

	if _, err := env.EvalString("(defn bad [] (assert false)) (bad)"); err == nil {
		t.Fatalf("expected bad to fail")
	}
	if env.addrstack.Size() != 0 {
		t.Fatalf("addrstack leaked after compiled-function error: got %d", env.addrstack.Size())
	}
	if env.linearstack.Size() != 1 {
		t.Fatalf("linearstack leaked after compiled-function error: got %d", env.linearstack.Size())
	}

	res := recentEval(t, env, "(+ 1 2)")
	if got := recentInt(t, res); got != 3 {
		t.Fatalf("after compiled-function error, (+ 1 2) = %d, want 3", got)
	}
}

func TestRecentTailCallCleansNewScope(t *testing.T) {
	env := NewZlisp()
	defer env.Close()

	res := recentEval(t, env, `
(defn nsTail [n]
  (newScope (cond (== n 0) 0 (nsTail (- n 1)))))
(nsTail 4)`)
	if got := recentInt(t, res); got != 0 {
		t.Fatalf("nsTail returned %d, want 0", got)
	}
	if env.linearstack.Size() != 1 {
		t.Fatalf("tail call through newScope leaked scopes: got linearstack size %d", env.linearstack.Size())
	}
}

func TestRecentTailCallInsideForDoesNotLeakLoopScope(t *testing.T) {
	env := NewZlisp()
	defer env.Close()

	recentEval(t, env, `
(defn loopTail [n]
  (for [(def i 0) (< i 1) (set i (+ i 1))]
    (cond (== n 0) 0 (loopTail (- n 1)))))
(loopTail 4)`)
	if env.linearstack.Size() != 1 {
		t.Fatalf("tail call through for leaked scopes: got linearstack size %d", env.linearstack.Size())
	}
}

func TestRecentBreakContinueCleanNestedScopes(t *testing.T) {
	env := NewZlisp()
	defer env.Close()

	recentEval(t, env, `
(for [(def i 0) (< i 3) (set i (+ i 1))]
  (let [x 1] (break)))`)
	if env.linearstack.Size() != 1 {
		t.Fatalf("break from nested let leaked scopes: got linearstack size %d", env.linearstack.Size())
	}

	env.Clear()
	recentEval(t, env, `
(for [(def i 0) (< i 2) (set i (+ i 1))]
  (let [x i] (continue)))`)
	if env.linearstack.Size() != 1 {
		t.Fatalf("continue from nested let leaked scopes: got linearstack size %d", env.linearstack.Size())
	}
}

func TestRecentDuplicateSymbolNumbersDoNotCollide(t *testing.T) {
	env := NewZlisp()
	defer env.Close()

	dup := env.Duplicate()
	dupSym := dup.MakeSymbol("dupOnlyRecent")
	envSym := env.MakeSymbol("envOnlyRecent")

	if dupSym.Number() == envSym.Number() {
		t.Fatalf("duplicate and original interned different symbols with the same number %d", dupSym.Number())
	}
}

func TestRecentDuplicateHasLoopStack(t *testing.T) {
	env := NewZlisp()
	defer env.Close()

	dup := env.Duplicate()
	defer dup.Close()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("loading a for-loop in Duplicate panicked: %v", r)
		}
	}()
	if err := dup.LoadString(`(for [(def i 0) (< i 1) (set i (+ i 1))] i)`); err != nil {
		t.Fatalf("Duplicate failed to load for-loop: %v", err)
	}
}

func TestRecentRunReportsStackUnderflow(t *testing.T) {
	env := NewZlisp()
	defer env.Close()

	env.mainfunc.fun = []Instruction{ClearStackmarkInstr{sym: env.MakeSymbol("__missing_stackmark")}}
	env.curfunc = env.mainfunc
	env.pc = 0

	if _, err := env.Run(); err != StackUnderFlowErr {
		t.Fatalf("Run() error = %v, want %v", err, StackUnderFlowErr)
	}
}

func TestRecentRunTreatsPCPastEndAsReachedEnd(t *testing.T) {
	env := NewZlisp()
	defer env.Close()

	env.mainfunc.fun = []Instruction{PushInstr{expr: &SexpInt{Val: 1}}}
	env.curfunc = env.mainfunc
	env.pc = 2

	res, err := env.Run()
	if err != nil {
		t.Fatalf("Run with pc past end failed: %v", err)
	}
	if res != SexpNull {
		t.Fatalf("Run with pc past end returned %v, want nil", res.SexpString(nil))
	}
}

func TestRecentClosureSnapshotStopsAtFunctionBoundary(t *testing.T) {
	env := NewZlisp()
	defer env.Close()

	recentEval(t, env, `(defn maker [] (fn [] 1))`)
	res := recentEval(t, env, `(let [noise 99] (maker))`)
	fn, ok := res.(*SexpFunction)
	if !ok {
		t.Fatalf("maker returned %T, want *SexpFunction", res)
	}
	if fn.closingOverScopes == nil {
		t.Fatalf("returned function has no closure")
	}
	if got := fn.closingOverScopes.Stack.Size(); got != 1 {
		t.Fatalf("closure captured %d scopes, want only the maker function scope", got)
	}
}

func TestRecentStackPopKeepsBackingArray(t *testing.T) {
	env := NewZlisp()
	defer env.Close()

	stack := env.NewStack(0)
	stack.Push(env.NewScope())
	stack.Push(env.NewScope())
	stack.Push(env.NewScope())
	before := cap(stack.elements)
	if _, err := stack.Pop(); err != nil {
		t.Fatalf("Pop failed: %v", err)
	}
	if after := cap(stack.elements); after != before {
		t.Fatalf("Pop changed backing capacity from %d to %d", before, after)
	}
}

func TestRecentGetStackTraceDoesNotConsumeAddrStack(t *testing.T) {
	env := NewZlisp()
	defer env.Close()

	env.addrstack.PushAddr(env.mainfunc, 12)
	env.addrstack.PushAddr(env.mainfunc, 34)
	before := env.addrstack.Size()

	trace := env.GetStackTrace(fmt.Errorf("sample"))
	if !strings.Contains(trace, "sample") {
		t.Fatalf("trace did not include error: %q", trace)
	}
	if after := env.addrstack.Size(); after != before {
		t.Fatalf("GetStackTrace consumed addrstack: before %d after %d", before, after)
	}
}
