package zygo

import (
	"bytes"
	"testing"

	cv "github.com/glycerine/goconvey/convey"
)

func Test001LexerPositionRecordingWorks(t *testing.T) {

	cv.Convey(`Given a function definition in a stream, the token positions should reflect the span of the source code function definition, so we can retreive the functions definition easily`, t, func() {

		str := `(defn hello [] "greetings!")`
		env := NewGlisp()
		stream := bytes.NewBuffer([]byte(str))
		lexer := NewLexerFromStream(stream)
		expressions, err := ParseTokens(env, lexer)
		panicOn(err)
		//goon.Dump(expressions[0])
		cv.So(expressions[0].SexpString(), cv.ShouldEqual, `(defn hello [] "greetings!")`)
	})
}

func Test006LexerAndParsingOfDotInvocations(t *testing.T) {

	cv.Convey(`Given a dot invocation method such as "(. subject method)" or "(.. subject method)", the parser should identify these as tokens, and should reject other tokens that *start with* dots. Tokens that start with dot '.' are special and reserved`, t, func() {

		str := `(. subject method)`
		env := NewGlisp()
		stream := bytes.NewBuffer([]byte(str))
		lexer := NewLexerFromStream(stream)
		expressions, err := ParseTokens(env, lexer)
		panicOn(err)
		//goon.Dump(expressions[0])
		cv.So(expressions[0].SexpString(), cv.ShouldEqual, `(. subject method)`)
	})
}
