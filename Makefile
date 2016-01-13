all:
	/bin/echo "package zygo" > repl/gitcommit.go
	/bin/echo "func init() { GITLASTTAG = \"$(shell git describe --abbrev=0 --tags)\"; GITLASTCOMMIT = \"$(shell git rev-parse HEAD)\" }" >> repl/gitcommit.go
	go install github.com/glycerine/zygomys/cmd/zygo

test:
	tests/testall.sh && cd repl && go test -v
