package zygo

import (
	"fmt"
	"testing"
)

func TestSandboxFunctions(t *testing.T) {

	// given
	s := NewSandboxSafeGlisp()

	// when
	sysFuncs := SystemFunctions()
	sandSafeFuncs := SandboxSafeFunctions()

	// then
	// no system functions should pass
	for name := range sysFuncs {
		_, err := s.EvalString(fmt.Sprintf("(println %s)", name))
		if err == nil {
			t.Error(err)
		}
	}

	// all sandSafeFuncs should be fine
	for name := range sandSafeFuncs {
		_, err := s.EvalString(fmt.Sprintf("(println %s)", name))
		if err == nil {
			t.Error(err)
		}
	}
}
