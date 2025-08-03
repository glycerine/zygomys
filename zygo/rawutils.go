package zygo

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// can only return true for SexpRaw that have Base64:true set.
func IsBase64Function(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpRaw:
		if t.Base64 {
			return &SexpBool{Val: true}, nil
		}
	}
	return &SexpBool{}, nil
}

func MakeRawUnbase64(name string) ZlispUserFunction {

	return func(env *Zlisp, _ string, args []Sexp) (Sexp, error) {

		if len(args) < 1 {
			return &SexpRaw{}, nil
		}

		switch t := args[0].(type) {
		case *SexpRaw:
			// set the display bit
			switch name {
			case "unbase64":
				t.Base64 = false
			case "base64":
				t.Base64 = true
			case "flipbase64":
				t.Base64 = !t.Base64
			}
			return t, nil
		case *SexpStr:
			// user wants to convert from base64 string to raw, ok.
			return Raw64Builder(env, name, args)
		}
		return SexpNull, errors.New("argument must be raw or base64 encoded string")
	}
}

// very slow when reading 200k strings, so
// prefer extractRawHelper() now. Here for
// test of equivalency in raw_test.go
func extractRawHelperSlow(s string) string {
	if s[0] == '`' {
		s = s[1:]
	}
	n := len(s)
	if s[n-1] == '`' {
		s = s[:n-1]
	}
	splt := strings.Split(s, "\n")
	return strings.Join(splt, "")
}

func extractRawHelper(s string) string {

	var buf strings.Builder
	by := UnsafeStringToByteSlice(s)
	// determine where to copy, begin and end
	beg := 0
	end := len(s)
	if s[0] == '`' {
		beg++
	}
	if s[end-1] == '`' {
		end--
	}

	// copy the bytes to buf.
	// We've gotta eliminate the embedded
	// newlines, since extractRawHelperSlow() does,
	// and we are substituting for it.

	slc := by[beg:end]
	for len(slc) > 0 {
		newlinepos := bytes.Index(slc, newlineByteSlc)
		if -1 == newlinepos {
			// no newlines (left) in slc
			buf.Write(slc)
			break
		} else {
			if newlinepos > 0 {
				buf.Write(slc[:newlinepos])
				slc = slc[newlinepos+1:]
			} else {
				slc = slc[1:]
			}
		}
	}
	return buf.String()
}

var newlineByteSlc = []byte("\n")

func Raw64Builder(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	n := len(args)
	if n < 1 {
		return &SexpRaw{}, nil
	}

	var joined string

	//vv("Raw64Builder has len args = %v", len(args)) // 220865 on big file.
	// optimization for case when all args are *SexpStr; try
	// to avoid the slowness of extractRawHelper()
	allStrings := true
	var allMyStrings []string
	for _, x := range args {
		t, ok := x.(*SexpStr)
		if !ok {
			allStrings = false
			break
		}
		allMyStrings = append(allMyStrings, t.S)
	}
	if allStrings {
		// inlined helper
		var buf strings.Builder
		for _, s := range allMyStrings {
			by := UnsafeStringToByteSlice(s)
			// determine where to copy, begin and end
			beg := 0
			end := len(s)
			if s[0] == '`' {
				beg++
			}
			if s[end-1] == '`' {
				end--
			}

			// copy the bytes to buf.
			// We've gotta eliminate the embedded
			// newlines, since extractRawHelper() does,
			// and we are substituting for it.

			slc := by[beg:end]
			for len(slc) > 0 {
				newlinepos := bytes.Index(slc, newlineByteSlc)
				if -1 == newlinepos {
					// no newlines (left) in slc
					buf.Write(slc)
					break
				} else {
					if newlinepos > 0 {
						buf.Write(slc[:newlinepos])
						slc = slc[newlinepos+1:]
					} else {
						slc = slc[1:]
					}
				}
			}
		}
		joined = buf.String()
	} else {
		allMyStrings = nil // allow gc

		for _, x := range args {
			switch t := x.(type) {
			case *SexpStr:
				//vv("Raw64Builder sees strings '%s'", t.S)
				joined += extractRawHelper(t.S)
			case *SexpSymbol:
				rhs, err := t.RHS(env)
				panicOn(err)
				switch tt := rhs.(type) {
				case *SexpStr:
					joined += extractRawHelper(tt.S)
				default:
					vv("don't know what to do with rhs of symbol %T/val=%v", rhs, rhs)
				}

			default:
				vv("don't know what to do with %T/val=%v", x, x)
			}
		}
	}
	//vv("base64 decoding this string: '%s'", joined)
	by, err := base64.URLEncoding.DecodeString(joined)
	panicOn(err)
	return &SexpRaw{
		Val:    by,
		Base64: true,
	}, nil
}

func MakeRaw64(args []Sexp) (*SexpRaw, error) {
	r, err := MakeRaw(args)
	if r != nil {
		r.Base64 = true
	}
	return r, err
}

func MakeRaw(args []Sexp) (*SexpRaw, error) {
	raw := make([]byte, 0)
	for i := 0; i < len(args); i++ {
		switch e := args[i].(type) {
		case *SexpStr:
			a := []byte(e.S)
			raw = append(raw, a...)
		case *SexpRaw:
			// passthrough, possibly to assist MakeRaw64 in changing display.
			// only works for the 1st arg, of course.
			return e, nil
		default:
			return &SexpRaw{},
				fmt.Errorf("raw takes only string arguments. We see %T: '%v'", e, e)
		}
	}
	return &SexpRaw{Val: raw}, nil
}

func RawToStringFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}

	switch t := args[0].(type) {
	case *SexpRaw:
		// we *want* to re-interpret the raw bytes as a string here.
		// So don't apply base64 transform.
		return &SexpStr{S: string(t.Val)}, nil
	}
	return SexpNull, errors.New("argument must be raw")
}

// actually duplicate/copy the underlying bytes
func CopyRawFunction(env *Zlisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	switch t := args[0].(type) {
	case *SexpRaw:
		buf := make([]byte, len(t.Val))
		copy(buf, t.Val)
		return &SexpRaw{Val: buf, Base64: t.Base64, Typ: t.Typ}, nil
	}
	return SexpNull, errors.New("argument must be raw")
}

func trimHelper(s string) string {
	if strings.HasPrefix(s, "(raw64 ") {
		s = s[len("raw64 "):]
	}
	n := len(s)
	if n > 0 && s[0] == ' ' {
		s = s[1:]
		n = len(s)
	}
	if n > 0 && s[0] == '"' {
		s = s[1:]
		n = len(s)
	}
	if n > 0 && s[0] == '`' {
		s = s[1:]
		n = len(s)
	}
	if n >= 2 && s[0:2] == `\"` {
		s = s[2:]
		n = len(s)
	}
	if n > 0 && s[n-1] == ')' {
		s = s[:n-1]
		n = len(s)
	}
	if n > 0 && s[n-1] == '`' {
		s = s[:n-1]
		n = len(s)
	}
	if n > 0 && s[n-1] == '"' {
		s = s[:n-1]
		n = len(s)
	}
	if n >= 2 && s[n-2:] == `\"` {
		s = s[:n-2]
		n = len(s)
	}
	return s
}

func ChunkedBase64StringToRaw(s string) *SexpRaw {
	// trim off the outer coatings
	s = trimHelper(s)
	splt := strings.Split(s, "\n")
	for i := range splt {
		splt[i] = trimHelper(splt[i])
	}
	join := strings.Join(splt, "")
	//vv("base64 decoding this string: '%s'", join)
	by, err := base64.URLEncoding.DecodeString(join)
	panicOn(err)
	return &SexpRaw{
		Val:    by,
		Base64: true,
	}
}

func ByteSliceToChunkedBase64String(b []byte) string {
	slc := ByteSliceToChunkedBase64StringNotJoined(b)
	return strings.Join(slc, "\n")
}

func ByteSliceToChunkedBase64StringNotJoined(b []byte) []string {
	s := base64.URLEncoding.EncodeToString(b)
	// chunk out to 80 char per line.
	res := []string{}
	left := s
	n := len(left)
	for n > 0 {
		if n <= 80 {
			res = append(res, `"`+left+`"`)
			break
		}
		res = append(res, `"`+left[:80]+`"`)
		left = left[80:]
		n -= 80
	}
	return res
}

func (r *SexpRaw) SexpString(ps *PrintState) string {
	if r.Base64 {
		s := ByteSliceToChunkedBase64String(r.Val)
		return "(raw64 " + s + ")"
	}
	return fmt.Sprintf("%#v", []byte(r.Val))
}
