#!/bin/sh

for lispfile in tests/*.glisp
do
    ./gl -exitonfail "${lispfile}" && \
        echo "${lispfile} passed" || \
        echo "${lispfile} failed"
done
