package zygo

import (
	"fmt"
	"io"
)

var _ = fmt.Printf

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

func (s *LazyParser) Stop() (err error) {
	//vv("LazyParser.Stop()")
	if s.psr != nil {
		//vv("LazyParser.Stop() calling psr.Stop()")
		err = s.psr.Stop()
		s.psr = nil
	}
	return
}

// Start is lazy, so this is a no-op.
func (s *LazyParser) Start() {
	//vv("LazyParser.Start()")
	s.psr = nil
}

func (s *LazyParser) refresh() {
	//vv("LazyParser.refresh()")
	if s.psr == nil {
		s.psr = s.env.NewParser()
		s.psr.EagerlyRetireParserGoro = true
		s.psr.Start()
	}
}

func (s *LazyParser) Reset() {
	//vv("LazyParser.Reset()")
	if s.psr != nil {
		s.psr.Stop()
		s.psr = nil
		//s.psr.Reset()
	}
}
func (s *LazyParser) NewInput(sc io.RuneScanner) {
	s.refresh()
	s.psr.NewInput(sc)
}
func (s *LazyParser) ResetAddNewInput(sc io.RuneScanner) {
	s.refresh()
	s.psr.ResetAddNewInput(sc)
}

func (s *LazyParser) ParseExpression(depth int) (res Sexp, err error) {
	s.refresh()
	res, err = s.psr.ParseExpression(depth)
	if err != nil {
		fmt.Printf("err in lazy ParseExpression is '%v'", err)
	}
	s.cleanupHelper()
	return
}
func (s *LazyParser) ParseTokens() (sx []Sexp, err error) {
	s.refresh()
	sx, err = s.psr.ParseTokens()
	if err != nil {
		//fmt.Printf("err in lazy ParseTokens is '%v'\n", err)
		if err == ErrParserTimeout {
			s.psr = nil
			s.refresh()
			sx, err = s.psr.ParseTokens()
		}
	}
	s.cleanupHelper()
	return
}

func (s *LazyParser) Linenum() int {
	if s.psr == nil {
		return 0
	}
	return s.psr.lexer.Linenum()
}
