package gdslext

import (
	"github.com/glycerine/godiesel/interpreter"
	"math/rand"
	"time"
)

var defaultRand = rand.New(rand.NewSource(time.Now().Unix()))

func RandomFunction(env *gdsl.Glisp, name string,
	args []gdsl.Sexp) (gdsl.Sexp, error) {
	return gdsl.SexpFloat(defaultRand.Float64()), nil
}

func ImportRandom(env *gdsl.Glisp) {
	env.AddFunction("random", RandomFunction)
}
