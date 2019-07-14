.PHONY: build test all

all: build test

build:
	/bin/echo "package zygo" > zygo/gitcommit.go
	/bin/echo "func init() { GITLASTTAG = \"$(shell git describe --abbrev=0 --tags)\"; GITLASTCOMMIT = \"$(shell git rev-parse HEAD)\" }" >> zygo/gitcommit.go
	go install github.com/glycerine/zygomys/cmd/zygo
	go install github.com/glycerine/goconvey/convey
	go install github.com/jtolds/gls
test:
	tests/testall.sh && echo "running 'go test'" && cd zygo && go test -v
