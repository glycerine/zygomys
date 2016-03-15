package zygo

import (
	"fmt"
	"time"
)

//go:generate msgp

//msgp:ignore Plane Wings Snoopy Hornet Hellcat

type Event struct {
	Id     int      `json:"id" msg:"id"`
	User   Person   `json:"user" msg:"user"`
	Flight string   `json:"flight" msg:"flight"`
	Pilot  []string `json:"pilot" msg:"pilot"`
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

// the interface Flyer confounds the msgp msgpack code generator,
// so put the msgp:ignore Plane above
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
