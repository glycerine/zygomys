package zygo

// a maker simply creates a go struct
// and returns a pointer to it.
type Maker interface {
	Make() interface{}
}

var MakerRegistry = map[string]Maker{}

func init() {
	MakerRegistry["event"] = &Event{}
	MakerRegistry["person"] = &Person{}
}

func (e *Event) Make() interface{} {
	return &Event{}
}

func (e *Person) Make() interface{} {
	return &Person{}
}
