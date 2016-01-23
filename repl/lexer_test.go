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
		defer env.parser.Stop()

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
		defer env.parser.Stop()

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

		fmt.Printf("\n\n ==== SymbolRegexp should function as expected\n")
		{
			// must use "\-" to avoid creating a range with just plain `-`
			// yes b: reg := `^[^a\-c]+$`
			// no  b: reg := `^[^a-c]+$`

			// have to allow & because it is the ... vararg indicator
			// wanted to allow $ to be system command indicator,
			// and possibly later allow for shell style substitution,
			// so it is always its own token/symbol, and should be accepted.
			symbolNotOkay := []string{`~`, `@`, `(`, `)`, `[`, `]`, `{`, `}`, `'`, `#`,
				`:`, `^`, `\`, `|`, `%`, `"`, `;`}
			// NB: have to allow  `a.b` and `a.b.` or else file paths,
			// used as arguments in macro (req) for example, won't
			// get lexed into symbols.

			//okay := []string{`..`, `a.b`, `-`, `a-b`, `*a-b*`, `$`, `&`, `.`, `.method`}
			symbolOkay := []string{`-`, `a-b`, `*a-b*`, `$`, `&`}

			// for experimentation, comment out the actual test below
			//reg := `^[^'#:;\\~@\[\]{}\^|"()%.]+$`
			//reg := `^[.]$|^[.][^'#:;\\~@\[\]{}\^|"()%.0-9][^'#:;\\~@\[\]{}\^|"()%]*$`
			//symbolRegex := regexp.MustCompile(reg)
			//x := symbolRegex

			CheckRegex(symbolNotOkay, symbolOkay, SymbolRegex) // SymbolRegex from lexer.go
		}

		fmt.Printf("\n\n ==== DotSymbolRegexp should function as expected\n")
		{
			//reg := `^[.]$|^[.][^'#:;\\~@\[\]{}\^|"()%.0-9][^'#:;\\~@\[\]{}\^|"()%]*$`
			//dotSymbolRegex := regexp.MustCompile(reg)

			dotSymbolNotOkay := []string{`~`, `@`, `(`, `)`, `[`, `]`, `{`, `}`, `'`, `#`,
				`:`, `^`, `\`, `|`, `%`, `"`, `;`, `.9`, `.a.`, `.a.b.`, `..`, `...`}

			//okay := []string{`..`, `a.b`, `-`, `a-b`, `*a-b*`, `$`, `&`, `.`, `.method`}
			dotSymbolOkay := []string{`.`, `.h`, `.method`, `.a.b`, `.a.b.c`}

			CheckRegex(dotSymbolNotOkay, dotSymbolOkay, DotSymbolRegex) // test DotSymbolRegex from lexer.go
		}
	})
}

func CheckRegex(notokay []string, okay []string, x *regexp.Regexp) {
	fmt.Printf("\nscanning notokay list =================\n")
	for _, a := range notokay {
		ans := x.MatchString(a)
		if ans {
			fmt.Printf("bad,  '%s' unwantedly matches '%s'\n", a, x)
		} else {
			fmt.Printf("good, '%s' does not match     '%s'\n", a, x)
		}
		cv.So(ans, cv.ShouldEqual, false)
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
}

func Test030LexingPauseAndResume(t *testing.T) {

	cv.Convey(`to enable the repl to properly detect the end of a multiline expression (or an expression containing quoted parentheses), the lexer should be able to pause and resume when more input is available.`, t, func() {

		str := `(defn hello [] "greetings!(((")`
		str1 := `(defn hel`
		str2 := `lo [] "greetings!(((")`
		env := NewGlisp()
		defer env.parser.Stop()
		stream := bytes.NewBuffer([]byte(str1))
		env.parser.ResetAddNewInput(stream)
		ex, err := env.parser.ParseTokens()

		P("\n In lexer_test, after parsing with incomplete input, we should get 0 expressions back.\n")
		cv.So(len(ex), cv.ShouldEqual, 0)
		P("\n In lexer_test, after ParseTokens on incomplete fragment, expressions = '%v' and err = '%v'\n", SexpArray(ex).SexpString(), err)

		P("\n In lexer_test: calling parser.NewInput() to provide str2='%s'\n", str2)
		env.parser.NewInput(bytes.NewBuffer([]byte(str2)))
		P("\n In lexer_test: done with parser.NewInput(), now calling parser.ParseTokens()\n")
		ex, err = env.parser.ParseTokens()
		P(`
 in lexer test: After providing the 2nd half of the input, we returned from env.parser.ParseTokens()
 with expressions = %v
 with err = %v
`, SexpArray(ex).SexpString(), err)

		cv.So(len(ex), cv.ShouldEqual, 1)
		panicOn(err)

		P("str=%s\n", str)
	})
}

func Test026RegexpSplittingOfDotSymbols(t *testing.T) {

	cv.Convey("our DotPartsRegex should split dot-symbol `.a.b.c` into `.a`, `.b`, and `.c`", t, func() {
		target := ".a.b.c"
		path := DotPartsRegex.FindAllString(target, -1)
		fmt.Printf("path = %#v\n", path)
		cv.So(len(path), cv.ShouldEqual, 3)
		cv.So(path[0], cv.ShouldEqual, ".a")
		cv.So(path[1], cv.ShouldEqual, ".b")
		cv.So(path[2], cv.ShouldEqual, ".c")
	})
}
