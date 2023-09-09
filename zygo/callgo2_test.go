package zygo

import (
	"fmt"
	"testing"

	cv "github.com/glycerine/goconvey/convey"
)

// more demonstrating how to pass data between script and Go.

var _ = fmt.Printf

type Table struct {
	Headers []string   `json:"headers" msg:"headers"`
	Rows    [][]string `json:"rows" msg:"rows"`
}

func Test019_ScriptCreatesData_GoReadsIt(t *testing.T) {

	cv.Convey(`example zygo script created content being then read from Go`, t, func() {

		env := NewZlisp()
		defer env.Close()

		// Typically you want to call env.StandardSetup()
		// right after creating a new env.
		// It will setup alot of parts of the env,
		// like defining the base types, allowing imports, etc.
		//
		// A sandboxed env, however, may not want to do this.
		env.StandardSetup()

		// Register the above Table struct, so we can copy
		// from zygo (table) to Go Table{}
		GoStructRegistry.RegisterUserdef(
			&RegisteredType{
				GenDefMap: true, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
					return &Table{}, nil
				}}, true, "table")

		code := `
        // A defmap is needed to define the table struct inside env.
        // The registry doesn't know about env(s), so it 
        // can't do it for us automatically.
        (defmap table)

        // Create an instance of table, with some data in it.
        (def t 
           (table headers:  ["wood"  "metal"] 
                  rows:    [["oak"  "silver"]
                            ["pine" "tin"   ]]))`

		//env.debugExec = true
		x, err := env.EvalString(code)
		panicOn(err)

		//vv("x = '%#v'", x)
		cv.So(x.(*SexpHash).TypeName, cv.ShouldEqual, "table")

		// provide a top level struct to fill in. In this
		// case the tree is just a 1 node deep.
		table := &Table{}
		tmp, err := SexpToGoStructs(x, table, env, nil, 0, table)
		panicOn(err)

		// The table and tmp are equal pointers. They point to the same Table.
		cv.So(table == tmp, cv.ShouldBeTrue)

		// The script created content is accessible from Go via tmp/table now.
		// Note that this is a copy.
		cv.So(table.Headers, cv.ShouldResemble, []string{"wood", "metal"})

		// So if we write to the copy...
		table.Headers[0] += "en ships, on the water"

		// the script version is unchanged...
		//
		// (Note that the assert will panic if it is not true.)
		_, err = env.EvalString(`(assert {t.headers[0] == "wood"})`)
		panicOn(err)

		switch f := tmp.(type) {
		case *Table:
			_ = f
			//fmt.Printf("my f is indeed a *Table: '%#v'", f)
		default:
			panic("wrong type")
		}

	})
}
