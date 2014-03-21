#!/bin/sh

for lispfile in tests/*.glisp
do
    ./glisp "${lispfile}" && echo "${lispfile}: All tests passed"
done
