package zygo

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"

	cv "github.com/glycerine/goconvey/convey"
)

func Test001LexerPositionRecordingWorks(t *testing.T) {

	cv.Convey(`Given a function definition in a stream, the token positions should reflect the span of the source code function definition, so we can retreive the functions definition easily`, t, func() {

		str := `(defn hello [] "greetings!")`
		env := NewGlisp()
		stream := bytes.NewBuffer([]byte(str))
		env.parser.ResetAddNewInput(stream)
		expressions, err := env.parser.ParseTokens()
		panicOn(err)
		//goon.Dump(expressions[0])
		cv.So(expressions[0].SexpString(), cv.ShouldEqual, `(defn hello [] "greetings!")`)
	})
}

func Test006LexerAndParsingOfDotInvocations(t *testing.T) {

	cv.Convey(`Given a dot invocation method such as "(. subject method)" or "(.. subject method)", the parser should identify these as tokens. Tokens that start with dot '.' are special and reserved for system functions.`, t, func() {

		str := `(. subject method)`
		env := NewGlisp()
		stream := bytes.NewBuffer([]byte(str))
		env.parser.ResetAddNewInput(stream)
		expressions, err := env.parser.ParseTokens()
		panicOn(err)
		//goon.Dump(expressions[0])
		cv.So(expressions[0].SexpString(), cv.ShouldEqual, `(. subject method)`)
	})
}

func Test025LexingOfStringAtomsAndSymbols(t *testing.T) {

	cv.Convey(`our symbol regex should accept/reject what we expect and define in the Language doc.`, t, func() {

		// must use "\-" to avoid creating a range with just plain `-`
		// yes b: reg := `^[^a\-c]+$`
		// no  b: reg := `^[^a-c]+$`

		// have to allow & because it is the ... vararg indicator
		// wanted to allow $ to be system command indicator,
		// and possibly later allow for shell style substitution,
		// so it is always its own token/symbol, and should be accepted.
		notokay := []string{`~`, `@`, `(`, `)`, `[`, `]`, `{`, `}`, `'`, `#`,
			`:`, `^`, `\`, `|`, `%`, `"`, `;`} // `.`, "`"}

		okay := []string{`..`, `a.b`, `-`, `a-b`, `*a-b*`, `$`, `&`, `.`, `.method`}

		// for experimentation, comment out the actual test below
		reg := `^[^'#:;\\~@\[\]{}\^|"()%&]+$`
		symbolRegex := regexp.MustCompile(reg)
		x := symbolRegex

		// actual test:
		x = SymbolRegex // from lexer.go

		fmt.Printf("\nscanning notokay list =================\n")
		for _, a := range notokay {
			ans := x.MatchString(a)
			cv.So(ans, cv.ShouldEqual, false)
			if ans {
				fmt.Printf("bad,  '%s' unwantedly matches '%s'\n", a, x)
			} else {
				fmt.Printf("good, '%s' does not match     '%s'\n", a, x)
			}
		}

		fmt.Printf("\nscanning okay list =================\n")
		for _, a := range okay {
			ans := x.MatchString(a)
			cv.So(ans, cv.ShouldEqual, true)
			if ans {
				fmt.Printf("good, '%s' matches as expected       '%s'\n", a, x)
			} else {
				fmt.Printf("bad,  '%s' does not match but should '%s'\n", a, x)
			}
		}

	})
}

func Test030LexingPauseAndResume(t *testing.T) {

	cv.Convey(`to enable the repl to properly detect the end of a multiline expression (or an expression containing quoted parentheses), the lexer should be able to pause and resume when more input is available.`, t, func() {

		str := `(defn hello [] "greetings!(((")`
		str1 := `(defn hel`
		str2 := `lo [] "greetings!(((")`
		env := NewGlisp()
		stream := bytes.NewBuffer([]byte(str1))
		env.parser.ResetAddNewInput(stream)
		ex, err := env.parser.ParseTokens()
		P("expressions = %#v\nerr = %v", ex, err)

		env.parser.NewInput(bytes.NewBuffer([]byte(str2)))
		ex, err = env.parser.Resume()
		P("expressions = %#v\nerr = %v", ex, err)

		P("str=%s\n", str)
	})
}
