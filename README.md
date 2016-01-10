# Zygomys

Zygomys is an embeddable Lisp REPL focused on Domain Specific Language
(DSL) creation. It is written in Go and plays easily with Go programs
and structs defined within them. It counts as its original ancestor
[Howard Mao's inspiring Glisp project](https://github.com/zhemao/glisp).
It borrows certain constructs from Clojure, and aims to make scripting
and configuration very easy with a minimal footprint.

Zygomys is ideally suited for driving complex configurations and providng
your project with a domain specific language customized to your challenges.
The example snippets in the tests/*.zy provide many examples.
The full [documentation can be found in the Wiki](https://github.com/glycerine/zygomys/wiki).

The standalone REPL is called simply `zygo`.

### Features in Zygomys v1.1.3:

 * [x] Clojure like threading `(-> hash field1: field2:)` and `(:field hash)` selection
 * [x] Raw bytes type `(raw string)` lets you do zero-copy `[]byte` manipulation
 * [x] Record definitions `(defmap)`
 * [x] Read external files with `(req path-to-file)`
 * [x] Go style raw string literals, using `backticks`.
 * [x] JSON and Msgpack interop: serialization and deserialization
 * [x] `(range key value hash (body))` range loops mirror for-range over a hash in Go.
 * [x] `(for [(initializer) (test) (advance)] (body))` for-loops match those in C. Use `(break)` and `(continue)` for loop control.

### Additional features

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

[See the wiki for lots of details and a full description of the Zygomys language.](https://github.com/glycerine/zygomys/wiki).

### where did the name Zygomys come from?

Zygomys is a contraction of Zygogeomys, a [genus of pocket gophers whose natural habitat is high-altitude forests.](https://en.wikipedia.org/wiki/Michoacan_pocket_gopher)

### License

Two-clause BSD, see LICENSE file.

### Author

Jason E. Aten, Ph.D.

### Credits

The grandparent LISP, [Glisp](https://github.com/zhemao/glisp), was designed and implemented by Howard Mao.
