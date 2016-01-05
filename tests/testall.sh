#!/bin/sh

for lispfile in tests/*.dsl
do
    ./gdsl -exitonfail "${lispfile}" && \
        echo "${lispfile} passed" || \
        echo "${lispfile} failed"
done
