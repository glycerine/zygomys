# GLisp

This is a LISP dialect designed as an embedded extension language for the Go
programming language. It is implemented in pure Go, so it can be easily ported
to all systems and architectures that Go targets.

It is currently very incomplete. Here is a list of what is implemented so far.

 * Floats, Ints, Chars, Strings, Symbols, Lists, and Array datatypes
 * Arithmetic (`+`, `-`, `*`, `/`, `mod`)
 * Shift Operators (`sll`, `srl`, `sra`)
 * Bitwise operations (`bit-and`, `bit-or`, `bit-xor`)
 * Comparison operations (`<`, `>`, `<=`, `>=`, `=`, and `not=`)
 * Short-circuit boolean operators (`and` and `or`)
 * Conditionals (`cond`)
 * Lambdas (`fn`)
 * Bindings (`def`, `defn`, and `let`)
 * A Basic Repl

## In Progress

 * Tail-call optimization
 * Channel and goroutine support
 * Sane Go API
 * Better name than "glisp"
