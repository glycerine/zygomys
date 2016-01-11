package zygo

import (
	"fmt"
	"github.com/shurcooL/go-goon"
	"time"
)

// demostrate calling Go

type Plane struct {
	Name  string  `json:"name" msg:"name"`
	Speed int     `json:"speed" msg:"speed"`
	Chld  []Flyer `json:"chld" msg:"chld"`
}

type Snoopy struct {
	Plane `json:"plane" msg:"plane"`
	Cry   string `json:"cry" msg:"cry"`
}

type Hornet struct {
	Plane `json:"plane" msg:"plane"`
}

type Hellcat struct {
	Plane `json:"plane" msg:"plane"`
}

func (p *Snoopy) Fly(ev *Weather) (s string, err error) {
	s = fmt.Sprintf("Snoopy sees weather %#v, cries '%s'", ev, p.Cry)
	fmt.Println(s)
	return
}

func (p *Snoopy) Sideeffect() {
	fmt.Printf("Sideeffect() called! p = %p\n", p)
}

func (b *Hornet) Fly(ev *Weather) (s string, err error) {
	fmt.Printf("Hornet sees weather %v", ev)
	return
}

func (b *Hellcat) Fly(ev *Weather) (s string, err error) {
	fmt.Printf("Hellcat sees weather %v", ev)
	return
}

type Flyer interface {
	Fly(ev *Weather) error
}

type Weather struct {
	Time    time.Time `json:"time" msg:"time"`
	Size    int64     `json:"size" msg:"size"`
	Type    string    `json:"type" msg:"type"`
	Details []byte    `json:"details" msg:"details"`
}

// Using reflection, invoke a Go method on a struct or interface.
// args[0] is a hash with an an attached GoStruct
// args[1] is a hash representing a method call on that struct.
// The returned Sexp is a hash that represents the result of that call.
func CallGo(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) < 2 {
		return SexpNull, WrongNargs
	}
	object, isHash := args[0].(SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("CallGo() error: first argument must be a hash or defmap with an attached GoObject")
	}
	method, isHash := args[1].(SexpHash)
	if !isHash {
		return SexpNull, fmt.Errorf("TwoHashCallGo() error: second argument must be a hash/record representing the method to call")
	}

	goon.Dump(object)
	goon.Dump(method)
	//fmt.Printf("expected = '%#v'\n", expected)

	// construct the call by reflection

	// return the result in a hash record too.
	return SexpNull, nil
}
