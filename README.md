![Image of Gopher flying](https://github.com/glycerine/zygomys/blob/master/biplane.jpg)

# Zygomys - fast, high level control

Zygomys is an embeddable Lisp interpreter and REPL (Read-Eval-Print-Loop;
that is, it comes with a command line interactive interface).
Zygomys is focused on Domain Specific Language (DSL) creation for your
scripting and configuration needs. It is written in Go and plays easily with Go programs
and structs defined within them. It counts as its original ancestor
[Howard Mao's inspiring Glisp project](https://github.com/zhemao/glisp).
It borrows certain constructs from Clojure, and others from Go, and
aims to make scripting and configuration very easy with a minimal footprint.

Zygomys is ideally suited for driving complex configurations and providng
your project with a domain specific language customized to your challenges.
The example snippets in the tests/*.zy provide many examples.
The full [documentation can be found in the Wiki](https://github.com/glycerine/zygomys/wiki).

The standalone REPL is called simply `zygo`.

### Not your Grandfather's LISP... features in Zygomys 1.1.7 include

 * [x] Clojure like threading `(-> hash field1: field2:)` and `(:field hash)` selection
 * [x] Raw bytes type `(raw string)` lets you do zero-copy `[]byte` manipulation
 * [x] Record definitions `(defmap)`
 * [x] Read external files with `(req path-to-file)`
 * [x] Go style raw string literals, using `backticks`.
 * [x] JSON and Msgpack interop: serialization and deserialization
 * [x] `(range key value hash (body))` range loops mirror for-range over a hash in Go.
 * [x] `(for [(initializer) (test) (advance)] (body))` for-loops match those in C. Both `(break)` and `(continue)` are available for additional loop control.
 * [x] Files can be recursively sourced with `(req path)` or `(source "path-string")`.
 * [x] Syntax-quote macro templates work inside `[]` arrays and `{}` hashes. `(macexpand)` is available for macro debugging.
 * [x] Easy to extend. See the `repl/random.go`, `repl/regexp.go`, and `repl/time.go` files for examples.

### Additional features

 * [x] Small code base, easy to integrate and use.
 * [x] Float, Int, Char, String, Symbol, List, Array, and Hash datatypes
 * [x] Arithmetic (`+`, `-`, `*`, `/`, `mod`, `**`)
 * [x] Shift Operators (`sll`, `srl`, `sra`)
 * [x] Bitwise operations (`bit-and`, `bit-or`, `bit-xor`)
 * [x] Comparison operations (`<`, `>`, `<=`, `>=`, `==`, `!=`, and `not=`)
 * [x] Short-circuit boolean operators (`and` and `or`)
 * [x] Conditionals (`cond`)
 * [x] Lambdas (`fn`)
 * [x] Bindings (`def`, `defn`, and `let`)
 * [x] Standalone and embedable REPL.
 * [x] Tail-call optimization
 * [x] Go API
 * [x] Macro System
 * [x] Syntax quoting -- with caret `^()` instead of backtick.
 * [x] Channel and goroutine support
 * [x] Pre- and Post- function call hooks

[See the wiki for lots of details and a full description of the Zygomys language.](https://github.com/glycerine/zygomys/wiki).

### where did the name Zygomys come from?

Zygomys is a contraction of Zygogeomys, [a genus of pocket gophers. The Michoacan pocket gopher (Zygogeomys trichopus) finds its natural habitat in high-altitude forests.](https://en.wikipedia.org/wiki/Michoacan_pocket_gopher)

### License

Two-clause BSD, see LICENSE file.

### Author

Jason E. Aten, Ph.D.

### Credits

The ancestor dialect, [Glisp](https://github.com/zhemao/glisp), was designed and implemented by [Howard Mao](https://zhehaomao.com/).

The Go gopher was designed by Renee French. (http://reneefrench.blogspot.com/)
The design is licensed under the Creative Commons 3.0 Attributions license.
Read this article for more details: https://blog.golang.org/gopher
