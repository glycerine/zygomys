#!/bin/sh

for lispfile in tests/*.zy
do
    ./zygo -exitonfail "${lispfile}" && \
        echo "${lispfile} passed" || \
        echo "${lispfile} failed"
done
