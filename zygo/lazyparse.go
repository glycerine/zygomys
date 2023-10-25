package zygo

import (
	"fmt"
	"io"
	"runtime"
)

var _ = fmt.Printf
var _ = runtime.NumGoroutine

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
	if s.psr != nil {
		recur := s.psr.getRecur()
		if recur == 0 {
			//vv("LazyParser.cleanupHelper true and recur == 0; stopping Parser. num goro = %v", runtime.NumGoroutine())
			s.psr.Stop()
			s.psr = nil
			// verified that goroutine count decreases by 1: Yes. It does.
			//vv("LazyParser.cleanupHelper true and recur == 0; stopped Parser. num goro = %v", runtime.NumGoroutine())
		} else {
			//vv("cleanupHelper doing nothing b/c recur = %v", recur)
		}
	}
	//vv("at end of LazyParser.cleanupHelper, s.psr = %p", s.psr)
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
	//vv("LazyParser.Start() called. is no-op.")
}

func (s *LazyParser) refresh() {
	if s.psr == nil {
		//vv("LazyParser.refresh() making new parser.")
		s.psr = s.env.NewParser()
		s.psr.Start()
	} else {
		//vv("LazyParser.refresh(): s.psr is not nil, so not making NewParser")
	}
}

func (s *LazyParser) Reset() {
	if s.psr != nil {
		//vv("LazyParser.Reset() stopping old parser")
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
	// make a parsing goroutine on demand (i.e. lazily)
	s.refresh()
	res, err = s.psr.ParseExpression(depth)
	if err != nil {
		fmt.Printf("err in lazy ParseExpression is '%v'", err)
	}
	// shutdown the parsing goroutine eagerly
	s.cleanupHelper()
	return
}
func (s *LazyParser) ParseTokens() (sx []Sexp, err error) {
	s.refresh()
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
