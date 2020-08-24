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
		env := NewZlisp()
		defer env.Close()

		stream := bytes.NewBuffer([]byte(str))
		env.parser.ResetAddNewInput(stream)
		expressions, err := env.parser.ParseTokens()
		panicOn(err)
		//goon.Dump(expressions[0])
		cv.So(expressions[0].SexpString(nil), cv.ShouldEqual, `(defn hello [] "greetings!")`)
	})
}

func Test002LexingScientificNotationOfFloats(t *testing.T) {

	cv.Convey(`Given a number 8.06e-05 it should be parsed as a single atom, not broken up at the '-' minus sign`, t, func() {

		str := `(def a 8.06e-05)`
		env := NewZlisp()
		defer env.Close()

		stream := bytes.NewBuffer([]byte(str))
		env.parser.ResetAddNewInput(stream)
		expressions, err := env.parser.ParseTokens()
		panicOn(err)
		//goon.Dump(expressions[0])
		cv.So(expressions[0].SexpString(nil), cv.ShouldEqual, `(def a 8.06e-05)`)

		str = `(def a 8.06e+05)`

		stream = bytes.NewBuffer([]byte(str))
		env.parser.ResetAddNewInput(stream)
		expressions, err = env.parser.ParseTokens()
		panicOn(err)
		//goon.Dump(expressions[0])
		cv.So(expressions[0].SexpString(nil), cv.ShouldEqual, `(def a 8.06e+05)`)

		str = `(def a 8.06e5)`

		stream = bytes.NewBuffer([]byte(str))
		env.parser.ResetAddNewInput(stream)
		expressions, err = env.parser.ParseTokens()
		panicOn(err)
		//goon.Dump(expressions[0])
		cv.So(expressions[0].SexpString(nil), cv.ShouldEqual, `(def a 8.06e+05)`)

	})
}

func Test006LexerAndParsingOfDotInvocations(t *testing.T) {

	cv.Convey(`Given a dot invocation method such as "(. subject method)" or "(.. subject method)", the parser should identify these as tokens. Tokens that start with dot '.' are special and reserved for system functions.`, t, func() {

		str := `(. subject method)`
		env := NewZlisp()
		defer env.Close()

		stream := bytes.NewBuffer([]byte(str))
		env.parser.ResetAddNewInput(stream)
		expressions, err := env.parser.ParseTokens()
		panicOn(err)
		//goon.Dump(expressions[0])
		cv.So(expressions[0].SexpString(nil), cv.ShouldEqual, `(. subject method)`)
	})
}

func Test025LexingOfStringAtomsAndSymbols(t *testing.T) {

	cv.Convey(`our symbol regex should accept/reject what we expect and define in the Language doc.`, t, func() {

		fmt.Printf("\n\n ==== SymbolRegexp should function as expected\n")
		{
			// must use "\-" to avoid creating a range with just plain `-`
			// yes b: reg := `^[^a\-c]+$`
			// no  b: reg := `^[^a-c]+$`

			symbolNotOkay := []string{`(`, `)`, `[`, `]`, `{`, `}`, `'`,
				`:`, `^`, `\`, `|`, `%`, `"`, `;`, `,`, `&`, `#`, `?`, `$`,
				`a-b`, `*a-b*`}

			symbolOkay := []string{`$hello`,
				`#a`, `?a`, `a:`, `$a:`, `#a:`, `?a:`}

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
				`:`, `^`, `\`, `|`, `%`, `"`, `;`, `.9`, `.a.`, `.a.b.`, `..`, `...`, `,`}

			//okay := []string{`..`, `a.b`, `-`, `a-b`, `*a-b*`, `$`, `&`, `.`, `.method`}
			dotSymbolOkay := []string{`.`, `.h`, `.method`, `.a.b`, `.a.b.c`, `a.b`}

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
		if ans {
			fmt.Printf("good, '%s' matches as expected       '%s'\n", a, x)
		} else {
			fmt.Printf("bad,  '%s' does not match but should '%s'\n", a, x)
		}
		cv.So(ans, cv.ShouldEqual, true)
	}
}

func Test030LexingPauseAndResume(t *testing.T) {

	cv.Convey(`to enable the repl to properly detect the end of a multiline expression (or an expression containing quoted parentheses), the lexer should be able to pause and resume when more input is available.`, t, func() {

		str := `(defn hello [] "greetings!(((")`
		str1 := `(defn hel`
		str2 := `lo [] "greetings!(((")`
		env := NewZlisp()
		defer env.Close()
		stream := bytes.NewBuffer([]byte(str1))
		env.parser.ResetAddNewInput(stream)
		ex, err := env.parser.ParseTokens()

		P("\n In lexer_test, after parsing with incomplete input, we should get 0 expressions back.\n")
		cv.So(len(ex), cv.ShouldEqual, 0)
		P("\n In lexer_test, after ParseTokens on incomplete fragment, expressions = '%v' and err = '%v'\n", (&SexpArray{Val: ex, Env: env}).SexpString(nil), err)

		P("\n In lexer_test: calling parser.NewInput() to provide str2='%s'\n", str2)
		env.parser.NewInput(bytes.NewBuffer([]byte(str2)))
		P("\n In lexer_test: done with parser.NewInput(), now calling parser.ParseTokens()\n")
		ex, err = env.parser.ParseTokens()
		P(`
 in lexer test: After providing the 2nd half of the input, we returned from env.parser.ParseTokens()
 with expressions = %v
 with err = %v
`, (&SexpArray{Val: ex, Env: env}).SexpString(nil), err)

		cv.So(len(ex), cv.ShouldEqual, 1)
		panicOn(err)

		P("str=%s\n", str)
	})
}

func Test031LexingPauseAndResumeAroundBacktickString(t *testing.T) {

	cv.Convey(`to enable the repl to properly detect the end of a multiline backtick string, the lexer should be able to pause and resume when more input is available.`, t, func() {

		str := `{a=` + "`\n\n`}"
		str1 := "{a=`"
		str2 := "\n\n`}"
		env := NewZlisp()
		defer env.Close()
		stream := bytes.NewBuffer([]byte(str1))
		env.parser.ResetAddNewInput(stream)
		ex, err := env.parser.ParseTokens()

		P("\n In lexer_test, after parsing with incomplete input, we should get 0 expressions back.\n")
		cv.So(len(ex), cv.ShouldEqual, 0)
		P("\n In lexer_test, after ParseTokens on incomplete fragment, expressions = '%v' and err = '%v'\n", (&SexpArray{Val: ex, Env: env}).SexpString(nil), err)

		P("\n In lexer_test: calling parser.NewInput() to provide str2='%s'\n", str2)
		env.parser.NewInput(bytes.NewBuffer([]byte(str2)))
		P("\n In lexer_test: done with parser.NewInput(), now calling parser.ParseTokens()\n")
		ex, err = env.parser.ParseTokens()
		P(`
 in lexer test: After providing the 2nd half of the input, we returned from env.parser.ParseTokens()
 with expressions = %v
 with err = %v
`, (&SexpArray{Val: ex, Env: env}).SexpString(nil), err)

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

func Test027BuiltinOperators(t *testing.T) {

	cv.Convey("our lexer should lex without needing space between builtin operators like `-` and `+`, so `a+b` should parse as three tokens", t, func() {
		// +, -, ++, --, :=, =, ==, <=, >=, <, >, <-, ->, *, **, `.`, /
		//  but first fwd slash takes care of /
		// recognizing 1st char: +, -, =, <, >, *, `.`  means shift to operator mode
		// where we recognized 1 and 2 character builtin operators.
		// once in operator: allowed 2nd tokens: +, -, =, -, *
		ans := BuiltinOpRegex.MatchString(`* `)
		cv.So(ans, cv.ShouldEqual, false)
		ans = BuiltinOpRegex.MatchString(`-1`)
		cv.So(ans, cv.ShouldEqual, false)
	})
}

func Test028FloatingPointRegex(t *testing.T) {

	cv.Convey("our lexer should recognize negative floating point and negative integers", t, func() {
		ans := DecimalRegex.MatchString(`-1`)
		cv.So(ans, cv.ShouldEqual, true)
		ans = FloatRegex.MatchString(`-1e-10`)
		cv.So(ans, cv.ShouldEqual, true)
	})
}

func Test042ImaginaryFloatingPointRegex(t *testing.T) {

	cv.Convey("our lexer should recognize complex/imaginary floating point numbers and these should not be confused with reals/floating point real-only numbers", t, func() {
		ans := FloatRegex.MatchString(`-1e-10i`)
		cv.So(ans, cv.ShouldEqual, false)
		ans = FloatRegex.MatchString(`-1.i`)
		cv.So(ans, cv.ShouldEqual, false)
		ans = FloatRegex.MatchString(`1.2i`)
		cv.So(ans, cv.ShouldEqual, false)
		ans = FloatRegex.MatchString(`.2i`)
		cv.So(ans, cv.ShouldEqual, false)

		ans = ComplexRegex.MatchString(`-1e-10i`)
		cv.So(ans, cv.ShouldEqual, true)
		ans = ComplexRegex.MatchString(`-1.i`)
		cv.So(ans, cv.ShouldEqual, true)
		ans = ComplexRegex.MatchString(`1.2i`)
		cv.So(ans, cv.ShouldEqual, true)
		ans = ComplexRegex.MatchString(`.2i`)
		cv.So(ans, cv.ShouldEqual, true)
	})
}

func Test042Uint64Regex(t *testing.T) {

	cv.Convey("our lexer should recognize uint64 by their ULL suffix, and allow the 0x prefix hex and the 0o prefix for octal", t, func() {
		ans := Uint64Regex.MatchString(`0xffULL`)
		cv.So(ans, cv.ShouldEqual, true)
		ans = Uint64Regex.MatchString(`0o777ULL`)
		cv.So(ans, cv.ShouldEqual, true)
		ans = Uint64Regex.MatchString(`ULL`)
		cv.So(ans, cv.ShouldEqual, false)
		ans = Uint64Regex.MatchString(`-1ULL`)
		cv.So(ans, cv.ShouldEqual, false)
		ans = Uint64Regex.MatchString(`0ULL`)
		cv.So(ans, cv.ShouldEqual, true)
	})
}
