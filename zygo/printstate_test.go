package zygo

import (
	cv "github.com/glycerine/goconvey/convey"
	"testing"
)

func Test040SeenMapWorks(t *testing.T) {

	cv.Convey(`To allow cycle detection, given a set of pointers of various types, Seen should set and note when they have seen.`, t, func() {
		ps := NewPrintState()
		a := &SexpPair{}
		b := &SexpPointer{}
		d := &SexpStr{}
		cv.So(ps.GetSeen(a), cv.ShouldBeFalse)
		cv.So(ps.GetSeen(b), cv.ShouldBeFalse)
		cv.So(ps.GetSeen(d), cv.ShouldBeFalse)

		ps.SetSeen(a, "a")
		ps.SetSeen(b, "b")

		cv.So(ps.GetSeen(a), cv.ShouldBeTrue)
		cv.So(ps.GetSeen(b), cv.ShouldBeTrue)
		cv.So(ps.GetSeen(d), cv.ShouldBeFalse)
	})
}
