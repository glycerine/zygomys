package gdsl

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Event) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "id":
			z.Id, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "user":
			var isz uint32
			isz, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "first":
					z.User.First, err = dc.ReadString()
					if err != nil {
						return
					}
				case "last":
					z.User.Last, err = dc.ReadString()
					if err != nil {
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		case "flight":
			z.Flight, err = dc.ReadString()
			if err != nil {
				return
			}
		case "pilot":
			var xsz uint32
			xsz, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Pilot) >= int(xsz) {
				z.Pilot = z.Pilot[:xsz]
			} else {
				z.Pilot = make([]string, xsz)
			}
			for xvk := range z.Pilot {
				z.Pilot[xvk], err = dc.ReadString()
				if err != nil {
					return
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Event) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "id"
	err = en.Append(0x84, 0xa2, 0x69, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Id)
	if err != nil {
		return
	}
	// write "user"
	// map header, size 2
	// write "first"
	err = en.Append(0xa4, 0x75, 0x73, 0x65, 0x72, 0x82, 0xa5, 0x66, 0x69, 0x72, 0x73, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.User.First)
	if err != nil {
		return
	}
	// write "last"
	err = en.Append(0xa4, 0x6c, 0x61, 0x73, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.User.Last)
	if err != nil {
		return
	}
	// write "flight"
	err = en.Append(0xa6, 0x66, 0x6c, 0x69, 0x67, 0x68, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Flight)
	if err != nil {
		return
	}
	// write "pilot"
	err = en.Append(0xa5, 0x70, 0x69, 0x6c, 0x6f, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Pilot)))
	if err != nil {
		return
	}
	for xvk := range z.Pilot {
		err = en.WriteString(z.Pilot[xvk])
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Event) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "id"
	o = append(o, 0x84, 0xa2, 0x69, 0x64)
	o = msgp.AppendInt(o, z.Id)
	// string "user"
	// map header, size 2
	// string "first"
	o = append(o, 0xa4, 0x75, 0x73, 0x65, 0x72, 0x82, 0xa5, 0x66, 0x69, 0x72, 0x73, 0x74)
	o = msgp.AppendString(o, z.User.First)
	// string "last"
	o = append(o, 0xa4, 0x6c, 0x61, 0x73, 0x74)
	o = msgp.AppendString(o, z.User.Last)
	// string "flight"
	o = append(o, 0xa6, 0x66, 0x6c, 0x69, 0x67, 0x68, 0x74)
	o = msgp.AppendString(o, z.Flight)
	// string "pilot"
	o = append(o, 0xa5, 0x70, 0x69, 0x6c, 0x6f, 0x74)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Pilot)))
	for xvk := range z.Pilot {
		o = msgp.AppendString(o, z.Pilot[xvk])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Event) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "id":
			z.Id, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "user":
			var isz uint32
			isz, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for isz > 0 {
				isz--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				case "first":
					z.User.First, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				case "last":
					z.User.Last, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		case "flight":
			z.Flight, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "pilot":
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Pilot) >= int(xsz) {
				z.Pilot = z.Pilot[:xsz]
			} else {
				z.Pilot = make([]string, xsz)
			}
			for xvk := range z.Pilot {
				z.Pilot[xvk], bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *Event) Msgsize() (s int) {
	s = 1 + 3 + msgp.IntSize + 5 + 1 + 6 + msgp.StringPrefixSize + len(z.User.First) + 5 + msgp.StringPrefixSize + len(z.User.Last) + 7 + msgp.StringPrefixSize + len(z.Flight) + 6 + msgp.ArrayHeaderSize
	for xvk := range z.Pilot {
		s += msgp.StringPrefixSize + len(z.Pilot[xvk])
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Person) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "first":
			z.First, err = dc.ReadString()
			if err != nil {
				return
			}
		case "last":
			z.Last, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Person) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "first"
	err = en.Append(0x82, 0xa5, 0x66, 0x69, 0x72, 0x73, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.First)
	if err != nil {
		return
	}
	// write "last"
	err = en.Append(0xa4, 0x6c, 0x61, 0x73, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Last)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Person) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "first"
	o = append(o, 0x82, 0xa5, 0x66, 0x69, 0x72, 0x73, 0x74)
	o = msgp.AppendString(o, z.First)
	// string "last"
	o = append(o, 0xa4, 0x6c, 0x61, 0x73, 0x74)
	o = msgp.AppendString(o, z.Last)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Person) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "first":
			z.First, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "last":
			z.Last, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z Person) Msgsize() (s int) {
	s = 1 + 6 + msgp.StringPrefixSize + len(z.First) + 5 + msgp.StringPrefixSize + len(z.Last)
	return
}
