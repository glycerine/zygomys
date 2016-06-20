package zygo

import (
	"fmt"
	"reflect"
)

// begin supporting SexpString structs

// PrintState threads the state of display through SexpString() and Show() calls,
// to give pretty-printing indentation and to avoid infinite looping on
// cyclic data structures.
type PrintState struct {
	Indent int
	Seen   Seen
}

func (ps *PrintState) SetSeen(x interface{}, name string) {
	if ps == nil {
		panic("can't SetSeen on a nil PrintState")
	}
	//P("SetSeen doing intake of x=%p, under name '%s'", x, name)
	//ps.Seen[reflect.ValueOf(x).Pointer()] = struct{}{}
	ps.Seen[reflect.ValueOf(x).Pointer()] = name
}

func (ps *PrintState) GetSeen(x interface{}) bool {
	if ps == nil {
		return false
	}
	up := reflect.ValueOf(x).Pointer()
	_, ok := ps.Seen[up]
	// debug
	if ok {
		//P("GetSeen reporting we have seen up=%x before", up)
	}
	return ok
}

func (ps *PrintState) GetIndent() int {
	if ps == nil {
		return 0
	}
	return ps.Indent
}

func (ps *PrintState) AddIndent(addme int) *PrintState {
	if ps == nil {
		return &PrintState{
			Indent: addme,
			Seen:   NewSeen(),
		}
	}
	return &PrintState{
		Indent: ps.Indent + addme,
		Seen:   ps.Seen,
	}
}

func NewPrintState() *PrintState {
	return &PrintState{
		Seen: NewSeen(),
	}
}

func NewPrintStateWithIndent(indent int) *PrintState {
	return &PrintState{
		Indent: indent,
		Seen:   NewSeen(),
	}
}

func (ps *PrintState) Clear() {
	ps.Indent = 0
	ps.Seen = NewSeen()
}

func (ps *PrintState) Dump() {
	fmt.Printf("ps Dump: ")
	if ps == nil {
		fmt.Printf("nil\n")
		return
	}
	for k, v := range ps.Seen {
		fmt.Printf("ps Dump: %p   -- %s\n", k, v)
	}
	fmt.Printf("\n")
}

// Seen tracks if a value has already been displayed, to
// detect and avoid cycles
//type Seen map[uintptr]struct{}
type Seen map[uintptr]string

func NewSeen() Seen {
	//return Seen(make(map[uintptr]struct{}))
	return Seen(make(map[uintptr]string))
}

// end supporting SexpString structs
