package zygo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shurcooL/go-goon"
	"io"
	"time"
)

// demostrate calling Go

type Wings struct {
	SpanCm int
}

type Plane struct {
	Wings

	//Name  string `json:"name" msg:"name"`
	Speed   int     `json:"speed" msg:"speed"`
	Chld    Flyer   `json:"chld" msg:"chld"`
	Friends []Flyer `json:"friends"`
}

type Snoopy struct {
	Plane    `json:"plane" msg:"plane"`
	Cry      string  `json:"cry" msg:"cry"`
	Pack     []int   `json:"pack"`
	Carrying []Flyer `json:"carrying"`
}

type Hornet struct {
	Plane    `json:"plane" msg:"plane"`
	Mass     float64
	Nickname string
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
	Fly(ev *Weather) (s string, err error)
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

// mirror from the sexp side:  pointers to hashes in arrays;
//  match them on the go side: pointers to GoStructs in arrays
// so that children hierarchies can be called.
func GoLinkFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	switch asHash := args[0].(type) {
	default:
		return SexpNull, fmt.Errorf("value must be a hash or defmap record")
	case SexpHash:
		tn := *(asHash.TypeName)
		factory, hasMaker := GostructRegistry[tn]
		if !hasMaker {
			return SexpNull, fmt.Errorf("type '%s' not registered in GostructRegistry", tn)
		}
		newStruct := factory(env)
		fmt.Printf("\n newStruct = %#v\n", newStruct)
		jsonBytes := []byte(SexpToJson(asHash))

		fmt.Printf("jsonBytes = '%s'\n", string(jsonBytes))

		jsonDecoder := json.NewDecoder(bytes.NewBuffer(jsonBytes))
		err := jsonDecoder.Decode(&newStruct)
		switch err {
		case io.EOF:
		case nil:
		default:
			return SexpNull, fmt.Errorf("error during jsonDecoder.Decode() on type '%s': '%s'", tn, err)
		}
	}

	return SexpNull, nil
}

//func (a Flyer) MarshalJSON() ([]byte, error) {
// return nil, nil
//}

/*
func (a Flyer) UnmarshalJSON(b []byte) (err error) {
	j, s, n := author{}, "", uint64(0)
	if err = json.Unmarshal(b, &j); err == nil {
		*a = Author(j)
		return
	}
	if err = json.Unmarshal(b, &s); err == nil {
		a.Email = s
		return
	}
	if err = json.Unmarshal(b, &n); err == nil {
		a.ID = n
	}
	return
}
*/
