package zygo

import (
	"encoding/binary"
	"github.com/glycerine/blake2b"
)

// Blake2bUint64 returns an 8 byte BLAKE2b cryptographic
// hash of the raw.
//
// we're using the pure go: https://github.com/dchest/blake2b
//
// but the C-wrapped refence may be helpful as well --
//
// reference: https://godoc.org/github.com/codahale/blake2
// reference: https://blake2.net/
// reference: https://tools.ietf.org/html/rfc7693
//
func Blake2bUint64(raw []byte) uint64 {
	cfg := &blake2b.Config{Size: 8}
	h, err := blake2b.New(cfg)
	panicOn(err)
	h.Write(raw)
	by := h.Sum(nil)
	return binary.LittleEndian.Uint64(by[:8])
}
