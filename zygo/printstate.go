package zygo

import (
	"fmt"
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
	ps.Seen[x] = struct{}{}
}

func (ps *PrintState) GetSeen(x interface{}) bool {
	if ps == nil {
		return false
	}
	_, ok := ps.Seen[x]
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
		fmt.Printf("ps Dump: %p   -- %v\n", k, v)
	}
	fmt.Printf("\n")
}

// Seen tracks if a value has already been displayed, to
// detect and avoid cycles.
//
/* Q: How to do garbage-collection safe graph traversal in a graph of Go objects?

A: "Instead of converting the pointer to a uintptr, just store the pointer
itself in a map[interface{}]bool.  If you encounter the same pointer
again, you will get the same map entry.  The GC must guarantee that
using pointers as map keys will work even if the pointers move."

- Ian Lance Taylor on golang-nuts (2016 June 20).
*/
type Seen map[interface{}]struct{}

func NewSeen() Seen {
	return Seen(make(map[interface{}]struct{}))
}

// end supporting SexpString structs
