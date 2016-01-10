all:
	/bin/echo "package zygo" > interpreter/gitcommit.go
	/bin/echo "func init() { GITLASTTAG = \"$(shell git describe --abbrev=0 --tags)\"; GITLASTCOMMIT = \"$(shell git rev-parse HEAD)\" }" >> interpreter/gitcommit.go
	go install && go install github.com/glycerine/zygo/interpreter && go build -o zygo && cp -p ./zygo $(GOPATH)/bin/

test:
	tests/testall.sh
