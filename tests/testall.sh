#!/bin/sh

for lispfile in tests/*.lisp
do
    ./glisp "${lispfile}"
done
