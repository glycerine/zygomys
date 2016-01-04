all:
	/bin/echo "package glisp" > interpreter/gitcommit.go
	/bin/echo "func init() { GITLASTTAG = \"$(shell git describe --abbrev=0 --tags)\"; GITLASTCOMMIT = \"$(shell git rev-parse HEAD)\" }" >> interpreter/gitcommit.go
	go install && go build -o gl

test:
	tests/testall.sh
