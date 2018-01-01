package zygo

import (
	"fmt"
	"time"
)

//go:generate msgp

//msgp:ignore Plane Wings Snoopy Hornet Hellcat SetOfPlanes

// the pointer wasn't getting followed.
type NestOuter struct {
	Inner *NestInner `msg:"inner" json:"inner" zid:"0"`
}

type NestInner struct {
	Hello string `msg:"hello" json:"hello" zid:"0"`
}

type Event struct {
	Id        int      `json:"id" msg:"id"`
	User      Person   `json:"user" msg:"user"`
	Flight    string   `json:"flight" msg:"flight"`
	Pilot     []string `json:"pilot" msg:"pilot"`
	Cancelled bool     `json:"cancelled" msg:"cancelled"`
}

type Person struct {
	First string `json:"first" msg:"first"`
	Last  string `json:"last" msg:"last"`
}

func (ev *Event) DisplayEvent(from string) {
	fmt.Printf("%s %#v", from, ev)
}

type Wings struct {
	SpanCm int
}

type SetOfPlanes struct {
	Flyers []Flyer `json:"flyers" msg:"flyers"`
}

// the interface Flyer confounds the msgp msgpack code generator,
// so put the msgp:ignore Plane above
type Plane struct {
	Wings

	ID      int     `json:"id" msg:"id"`
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

func (p *Snoopy) Fly(w *Weather) (s string, err error) {
	w.Type = "VERY " + w.Type // side-effect, for demo purposes
	s = fmt.Sprintf("Snoopy sees weather '%s', cries '%s'", w.Type, p.Cry)
	fmt.Println(s)
	for _, flyer := range p.Friends {
		flyer.Fly(w)
	}
	return
}

func (p *Snoopy) GetCry() string {
	return p.Cry
}

func (p *Snoopy) EchoWeather(w *Weather) *Weather {
	return w
}

func (p *Snoopy) Sideeffect() {
	fmt.Printf("Sideeffect() called! p = %p\n", p)
}

func (b *Hornet) Fly(w *Weather) (s string, err error) {
	fmt.Printf("Hornet.Fly() called. I see weather %v\n", w.Type)
	return
}

func (b *Hellcat) Fly(w *Weather) (s string, err error) {
	fmt.Printf("Hellcat.Fly() called. I see weather %v\n", w.Type)
	return
}

type Flyer interface {
	Fly(w *Weather) (s string, err error)
}

type Weather struct {
	Time    time.Time `json:"time" msg:"time"`
	Size    int64     `json:"size" msg:"size"`
	Type    string    `json:"type" msg:"type"`
	Details []byte    `json:"details" msg:"details"`
}

func (w *Weather) IsSunny() bool {
	return w.Type == "sunny"
}

func (env *Zlisp) ImportDemoData() {

	env.AddFunction("nestouter", DemoNestInnerOuterFunction)
	env.AddFunction("nestinner", DemoNestInnerOuterFunction)

	rt := &RegisteredType{GenDefMap: true, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &NestOuter{}, nil
	}}
	GoStructRegistry.RegisterUserdef(rt, true, "nestouter", "NestOuter")

	rt = &RegisteredType{GenDefMap: true, Factory: func(env *Zlisp, h *SexpHash) (interface{}, error) {
		return &NestInner{}, nil
	}}
	GoStructRegistry.RegisterUserdef(rt, true, "nestinner", "NestInner")

}

// constructor
func DemoNestInnerOuterFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {

	n := len(args)
	switch n {
	case 0:
		return SexpNull, WrongNargs
	default:
		// many parameters, treat as key:value pairs in the hash/record.
		return ConstructorFunction(env, "msgmap", append([]Sexp{&SexpStr{S: name}}, MakeList(args)))
	}
}
