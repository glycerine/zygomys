![Image of Gopher flying](https://github.com/glycerine/zygomys/blob/master/media/high_altitude_gopher.png)

# zygomys - an embedded scripting language for Go

### quickly create a mini-language to drive your project

zygomys is an embeddable scripting language. It is a modernized Lisp with an object-oriented flavor, and
provides an interpreter and REPL (Read-Eval-Print-Loop;
that is, it comes with a command line interactive interface).

## why use zygomys?

zygomys allows you to create a Domain Specific Language to drive
your program with minimal fuss and maximum convenience.

It is written in Go and plays nicely with Go programs
and Go structs, using reflection to instantiate trees of Go structs
from the scripted configuration. These data structures are native
Go, and Go methods will run on them at compiled-Go speed.

Because it speaks JSON and Msgpack fluently, zygomys is ideally suited for driving
complex configurations and providing projects with a domain specific
language customized to your problem domain.

The example snippets in the tests/*.zy provide many examples.
The full [documentation can be found in the Wiki](https://github.com/glycerine/zygomys/wiki).
zygomys blends traditional and new. While the s-expression syntax
defines a Lisp, zygomys borrows some syntax from Clojure,
and some (notably the for-loop style) directly from the Go/C tradition.

The standalone REPL is called simply `zygo`.  `zygo` is also shorthand
for the whole project when speaking aloud. In writing, the full
`zygomys` is used to aid searchability.

### installation

~~~
$ go get github.com/glycerine/zygomys/cmd/zygo
~~~

### not your average parentheses... features in zygomys 4.7.0 include

 * [x] package mechanism that supports modularity and isolation of scripts/packages/libraries from each other. See tests/package.zy
 * [x] struct defintion and type checking. [See `tests/declare.zy` for examples.](https://github.com/glycerine/zygomys/blob/master/tests/declare.zy)
 * [x] Readable nested method calls: `(a.b.c.Fly)` calls method `Fly` on object `c` that lives within objects `a` and `b`.
 * [x] Use `zygo` to configure trees of Go structs, and then run methods on them at natively-compiled speed (since you are calling into Go code).
 * [x] sandbox-able environment; try `zygo -sandbox` and see the NewGlispSandbox() function.
 * [x] `emacs/zygo.el` emacs mode provides one-keypress stepping through code.
 * [x] Command-line editing, with tab-complete for keywords (courtesy of https://github.com/peterh/liner)
 * [x] JSON and Msgpack interop: serialization and deserialization
 * [x] `(range key value hash_or_array (body))` range loops act like Go for-range loops: iterate through hashes or arrays.
 * [x] `(for [(initializer) (test) (advance)] (body))` for-loops match those in C and Go. Both `(break)` and `(continue)` are available for additional loop control, and can be labeled to break out of nested loops.
 * [x] Raw bytes type `(raw string)` lets you do zero-copy `[]byte` manipulation.
 * [x] Record definitions `(defmap)` make configuration a breeze.
 * [x] Files can be recursively sourced with `(source "path-string")`.
 * [x] Go style raw string literals, using `` `backticks` ``, can contain newlines and `"` double quotes directly. Easy templating.
 * [x] Easy to extend. See the `repl/random.go`, `repl/regexp.go`, and `repl/time.go` files for examples.
 * [x] Clojure-like threading `(-> hash field1: field2:)` and `(:field hash)` selection. 
 * [x] Lisp-style macros for your DSL.
 * [x] optional infix notation within `{}` curly braces. Expressions typed at the REPL are assumed to be infix (wrapped in {} implicitly), enhancing the REPL experience for dealing with math.

### obligatory XKCD

![Obligatory XKCD: "elegant weapons... for a more civilized age"](http://imgs.xkcd.com/comics/lisp_cycles.png)


### additional features

 * [x] Go-style comments, both `/*block*/` and `//through end-of-line.`
 * [x] zygomys is a small Go library, easy to integrate and use/extend.
 * [x] Float (float64), Int (int64), Char, String, Symbol, List, Array, and Hash datatypes builtin.
 * [x] Arithmetic (`+`, `-`, `*`, `/`, `mod`, `**`)
 * [x] Shift Operators (`sll`, `srl`, `sra`)
 * [x] Bitwise operations (`bitAnd`, `bitOr`, `bitXor`)
 * [x] Comparison operations (`<`, `>`, `<=`, `>=`, `==`, `!=`)
 * [x] Short-circuit boolean operators (`and` and `or`)
 * [x] Conditionals (`cond`)
 * [x] Lambdas (`fn`)
 * [x] Bindings (`def`, `defn`, `let`, `letseq`)
 * [x] Standalone and embedable REPL.
 * [x] Tail-call optimization
 * [x] Go API
 * [x] Macro System with macexpand `(macexpand (yourMacro))` makes writing/debugging macros easier.
 * [x] Syntax quoting -- with caret `^()` instead of backtick.
 * [x] Backticks used for raw multiline strings, as in Go.
 * [x] Lisp-expression quoting uses `%` (not `'`; which delimits runes as in Go).
 * [x] Channel and goroutine support
 * [x] Full closures with lexical scope.

[See the wiki for lots of details and a full description of the zygomys language.](https://github.com/glycerine/zygomys/wiki).

### where did the name zygomys come from?

zygomys is a contraction of Zygogeomys, [a genus of pocket gophers. The Michoacan pocket gopher (Zygogeomys trichopus) finds its natural habitat in high-altitude forests.](https://en.wikipedia.org/wiki/Michoacan_pocket_gopher)

The term is also descriptive. The stem `zygo` comes from the Greek for yoke, indicating a pair or a union of two things, and `mys` comes from the Greek for mouse. The union of Go and Lisp in a small cute package, that is zygomys.


### license

Two-clause BSD, see LICENSE file.

### author

Jason E. Aten, Ph.D.

### credits

The ancestor dialect, [Glisp](https://github.com/zhemao/glisp), was designed and implemented by [Howard Mao](https://zhehaomao.com/).

The Go gopher was designed by Renee French. (http://reneefrench.blogspot.com/)
The design is licensed under the Creative Commons 3.0 Attributions license.
Read this article for more details: https://blog.golang.org/gopher

[XKCD https://xkcd.com/297/](https://xkcd.com/297/) licensed under a Creative Commons Attribution-NonCommercial 2.5 License(https://xkcd.com/license.html).
