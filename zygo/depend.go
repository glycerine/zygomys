package zygo

// test dependencies don't always get noticed
// go get (without the -t flag)
// make sure the go-convey gets installed
// when we are

import (
	cv "github.com/glycerine/goconvey/convey"
	"testing"
)

// don't need to call, here for the import dependency only
func UseConveyDummy() {
	var t testing.T
	cv.Convey("only here to force get get to fetch the goconvey dependency", t, func() {})
}
