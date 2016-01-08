# GoDiesel

GoDiesel is an embeddable Lisp REPL focused on Domain Specific Language
(DSL) creation. It is written in Go and plays easily with Go programs
and structs defined within them. It was originally derived from
[Howard Mao's terrific Glisp project](https://github.com/zhemao/glisp).

GoDiesel's features have evolved to the point where it is a distinct dialect.
This was done to support certain critical features like raw string literals,
to support for certain scheme and clojure idioms, and to make it easier
to use as a library within a larger program.

It is ideally suited for driving complex configurations and providng
your project with a domain specific language customized to your challenges.
The example snippets in the tests/*.dsl provide many examples.
The full documentation can be found in the [Wiki](https://github.com/glycerine/godiesel/wiki).

Brief list of implemented features.

 * [x] Small code base, easy to extend and integrate.
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

Features in GoDiesel v1.1.1:

 * [x] Clojure like threading `(-> hash field1: field2:)` and `(:field hash)` selection
 * [x] Raw bytes type `(raw string)` lets you do zero-copy []byte manipulation
 * [x] Record definitions `(defmap)`
 * [x] Read external files with `(req path-to-file)`
 * [x] Go style raw string literals, using `backticks`.
 * [x] JSON and Msgpack interop: serialization and deserialization
 * [x] `(range key value hash (body))` range loops mirror for-range over a hash in Go.
 * [x] `(for [(initializer) (test) (advance)] (body))` for-loops match those in C.
 