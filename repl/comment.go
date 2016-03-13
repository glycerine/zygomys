package zygo

type SexpComment struct {
	Comment string
	Block   bool
}

func (p *SexpComment) SexpString() string {
	return p.Comment
}

func (p *SexpComment) Type() *RegisteredType {
	return GoStructRegistry.Registry["comment"]
}
