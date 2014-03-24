#!/bin/sh

for lispfile in tests/*.glisp
do
    ./glisp "${lispfile}" && echo "${lispfile}: All tests passed" || \
        echo "test failure in ${lispfile}"
done
