#!/bin/sh

for lispfile in tests/*.glisp
do
    ./glisp -exitonfail "${lispfile}" || echo "${lispfile} failed"
done
