#!/bin/sh

for lispfile in tests/*.lisp
do
    ./glisp "${lispfile}" && echo "${lispfile}: All tests passed"
done
