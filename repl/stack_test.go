package zygo

import (
	"fmt"
	cv "github.com/glycerine/goconvey/convey"
	"testing"
)

func Test020StacksDontAlias(t *testing.T) {

	cv.Convey(`stack.Clone() should avoid all aliasing, as should Pop()`, t, func() {
		env := NewGlisp()
		t := env.NewStack(5)
		a := env.NewScope()
		b := env.NewScope()
		c := env.NewScope()

		t.Push(a)
		t.Push(b)

		show := func(s *Stack, b string) {
			pr, _ := s.Show(env, 0, b)
			fmt.Println(pr)
		}
		show(t, "t")

		r := t.Clone()
		show(r, "r")

		t.Push(c)

		show(t, "t after t.push(c)")
		show(r, "r after t.push(c)")
	})
}
