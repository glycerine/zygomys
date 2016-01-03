package glisp

import "fmt"

// version information. See Makefile and gitcommit.go for update/init.
var GITLASTTAG string
var GITLASTCOMMIT string

func Version() string {
	return fmt.Sprintf("%s/%s", GITLASTTAG, GITLASTCOMMIT)
}
