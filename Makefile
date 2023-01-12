.PHONY: build test all

all: build test

build:
	/bin/echo "package zygo" > zygo/gitcommit.go
	/bin/echo "func init() { GITLASTTAG = \"$(shell git describe --abbrev=0 --tags)\"; GITLASTCOMMIT = \"$(shell git rev-parse HEAD)\" }" >> zygo/gitcommit.go
	cd cmd/zygo; go install .

test:
	tests/testall.sh && echo "running 'go test'" && cd zygo && go test -v
