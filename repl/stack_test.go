package zygo

import (
	"fmt"
	"testing"

	cv "github.com/glycerine/goconvey/convey"
)

func Test020StacksDontAlias(t *testing.T) {

	cv.Convey(`stack.Clone() should avoid all aliasing, as should Pop()`, t, func() {
		env := NewGlisp()
		defer env.parser.Stop()

		t := env.NewStack(5)
		a := env.NewScope()
		b := env.NewScope()
		c := env.NewScope()

		t.Push(a)
		t.Push(b)

		show := func(s *Stack, b string) {
			pr, _ := s.Show(env, nil, b)
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
