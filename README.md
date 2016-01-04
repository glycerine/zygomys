# GLisp

This is a LISP dialect designed as an embedded extension language for the Go
programming language. It is implemented in pure Go, so it can be easily ported
to all systems and architectures that Go targets.

Here is a list of implemented features.

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
 * [x] Clojure like threading `(-> hash field1: field2:)` and `(:field hash)` selection
 * [x] Raw bytes type `(raw string)` lets you do zero-copy []byte manipulation
 * [x] Record definitions `(defmap)`
 * [x] Read external files with `(source path-to-file)`
 * [x] Go style raw string literals
 * [x] JSON and Msgpack interop: serialization and deserialization
 
The full documentation can be found in the [Wiki](https://github.com/glycerine/glisp/wiki).
