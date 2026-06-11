package zygo

import "testing"

func TestInfixSeparatedLabelForLoopMatchesSexpLabelLoop(t *testing.T) {
	env := NewZlisp()
	defer env.Close()
	env.StandardSetup()

	recentEval(t, env, `
{
	isum := 0
	jsum := 0
	outerLoop:

	for i := 1; i < 5; i++ {
		isum = isum + i

		innerLoop:
		for j := 1; j < 5; j++ {
			jsum = jsum + j
			if j > 2 {
				continue outerLoop
			}
			if i > 2 && j > 3 {
				break outerLoop
			}
			jsum = jsum + 1000
		}
	}
	(assert (== isum 10))
	(assert (== jsum 8024))
}
`)
}
