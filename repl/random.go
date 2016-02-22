package zygo

import (
	"math/rand"
	"time"
)

var defaultRand = rand.New(rand.NewSource(time.Now().Unix()))

func RandomFunction(env *Glisp, name string,
	args []Sexp) (Sexp, error) {
	return SexpFloat{Val: defaultRand.Float64()}, nil
}

func (env *Glisp) ImportRandom() {
	env.AddFunction("random", RandomFunction)
}
