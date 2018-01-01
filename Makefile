.PHONY: build test all

all: build test

build:
	/bin/echo "package zygo" > zygo/gitcommit.go
	/bin/echo "func init() { GITLASTTAG = \"$(shell git describe --abbrev=0 --tags)\"; GITLASTCOMMIT = \"$(shell git rev-parse HEAD)\" }" >> zygo/gitcommit.go
	go install github.com/glycerine/zygomys/cmd/zygo

test:
	tests/testall.sh && echo "running 'go test'" && cd zygo && go test -v
