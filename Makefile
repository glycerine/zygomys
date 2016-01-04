all:
	/bin/echo "package glisp" > interpreter/gitcommit.go
	/bin/echo "func init() { GITLASTTAG = \"$(shell git describe --abbrev=0 --tags)\"; GITLASTCOMMIT = \"$(shell git rev-parse HEAD)\" }" >> interpreter/gitcommit.go
	go install && go install github.com/glycerine/glisp/interpreter && go build -o gl && cp -p ./gl $(GOPATH)/bin/

test:
	tests/testall.sh
