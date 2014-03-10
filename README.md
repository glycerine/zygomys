# GLisp

This is a LISP dialect designed as an embedded extension language for the Go
program language. It is implemented in pure Go, so it can be easily ported
to all systems and architectures that Go targets.

It is currently very incomplete. Here is a list of what is implemented so far.

 * Floats, Ints, Chars, Strings, Symbols, Lists, and Array datatypes
 * <, >, <=, >=, =, and not= comparison operators
 * `and` and `or` short-circuit boolean operators
 * Conditionals
 * Lambdas (`fn`)
 * Bindings (`def` and `defn`)

## In Progress

 * Arithmetic
 * Bitwise operations
 * `let` statement
 * Tail-call optimization
 * Sane Go API
 * Better name than "glisp"
