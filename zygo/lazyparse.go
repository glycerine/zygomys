package zygo

import (
	"io"
)

// ParserI is implemented by Parser and LazyParser
type ParserI interface {
	Stop() error
	Start()
	Reset()
	NewInput(s io.RuneScanner)
	ResetAddNewInput(s io.RuneScanner)
	ParseExpression(depth int) (res Sexp, err error)
	ParseTokens() (sx []Sexp, err error)
	Linenum() int
}

// LazyParser wraps Parser and creates a parser only
// when needed. It shuts down the background Parser
// goroutine once parsing of the current expression is
// complete.
type LazyParser struct {
	env *Zlisp
	psr *Parser
}

func (env *Zlisp) NewLazyParser() *LazyParser {
	return &LazyParser{env: env}
}

// cleanupHelper checks s.psr state.
// If we are done with a complete parse, and have no state
// left on the Parser goroutine, then we shut it down eagerly
// to avoid leaving extra goroutines around.
func (s *LazyParser) cleanupHelper() {

}

func (s *LazyParser) Stop() error {
	if s.psr != nil {
		return s.psr.Stop()
	}
	return nil
}

// Start is lazy, so this is a no-op.
func (s *LazyParser) Start() {}

func (s *LazyParser) Reset() {
	if s.psr != nil {
		//s.psr.Stop()
		s.psr.Reset()
	}
}
func (s *LazyParser) NewInput(sc io.RuneScanner) {
	if s.psr == nil {
		s.psr = s.env.NewParser()
		s.psr.Start()
	}
	s.psr.NewInput(sc)
}
func (s *LazyParser) ResetAddNewInput(sc io.RuneScanner) {
	if s.psr == nil {
		s.psr = s.env.NewParser()
		s.psr.Start()
	}
	s.psr.ResetAddNewInput(sc)
}

func (s *LazyParser) ParseExpression(depth int) (res Sexp, err error) {
	if s.psr == nil {
		s.psr = s.env.NewParser()
		s.psr.Start()
	}
	res, err = s.psr.ParseExpression(depth)
	s.cleanupHelper()
	return
}
func (s *LazyParser) ParseTokens() (sx []Sexp, err error) {
	if s.psr == nil {
		s.psr = s.env.NewParser()
		s.psr.Start()
	}
	sx, err = s.psr.ParseTokens()
	s.cleanupHelper()
	return
}

func (s *LazyParser) Linenum() int {
	if s.psr == nil {
		return 0
	}
	return s.psr.lexer.Linenum()
}
