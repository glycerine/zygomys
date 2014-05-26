# GLisp

This is a LISP dialect designed as an embedded extension language for the Go
programming language. It is implemented in pure Go, so it can be easily ported
to all systems and architectures that Go targets.

It is currently very incomplete. Here is a list of what is implemented so far.

 * Float, Int, Char, String, Symbol, List, Array, and Hash datatypes
 * Arithmetic (`+`, `-`, `*`, `/`, `mod`)
 * Shift Operators (`sll`, `srl`, `sra`)
 * Bitwise operations (`bit-and`, `bit-or`, `bit-xor`)
 * Comparison operations (`<`, `>`, `<=`, `>=`, `=`, and `not=`)
 * Short-circuit boolean operators (`and` and `or`)
 * Conditionals (`cond`)
 * Lambdas (`fn`)
 * Bindings (`def`, `defn`, and `let`)
 * A Basic Repl
 * Tail-call optimization
 * Go API

The full documentation can be found in the [Wiki](https://github.com/zhemao/glisp/wiki).

## In Progress

 * Channel and goroutine support
 * Macro System
