all:
	/bin/echo "package gdsl" > interpreter/gitcommit.go
	/bin/echo "func init() { GITLASTTAG = \"$(shell git describe --abbrev=0 --tags)\"; GITLASTCOMMIT = \"$(shell git rev-parse HEAD)\" }" >> interpreter/gitcommit.go
	go install && go install github.com/glycerine/godiesel/interpreter && go build -o gdsl && cp -p ./gdsl $(GOPATH)/bin/

test:
	tests/testall.sh
