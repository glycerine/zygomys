package zygo

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
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
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
			var zb0002 uint32
			zb0002, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zb0002 > 0 {
				zb0002--
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
			var zb0003 uint32
			zb0003, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Pilot) >= int(zb0003) {
				z.Pilot = (z.Pilot)[:zb0003]
			} else {
				z.Pilot = make([]string, zb0003)
			}
			for za0001 := range z.Pilot {
				z.Pilot[za0001], err = dc.ReadString()
				if err != nil {
					return
				}
			}
		case "cancelled":
			z.Cancelled, err = dc.ReadBool()
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
func (z *Event) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "id"
	err = en.Append(0x85, 0xa2, 0x69, 0x64)
	if err != nil {
		return
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
		return
	}
	err = en.WriteString(z.User.First)
	if err != nil {
		return
	}
	// write "last"
	err = en.Append(0xa4, 0x6c, 0x61, 0x73, 0x74)
	if err != nil {
		return
	}
	err = en.WriteString(z.User.Last)
	if err != nil {
		return
	}
	// write "flight"
	err = en.Append(0xa6, 0x66, 0x6c, 0x69, 0x67, 0x68, 0x74)
	if err != nil {
		return
	}
	err = en.WriteString(z.Flight)
	if err != nil {
		return
	}
	// write "pilot"
	err = en.Append(0xa5, 0x70, 0x69, 0x6c, 0x6f, 0x74)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Pilot)))
	if err != nil {
		return
	}
	for za0001 := range z.Pilot {
		err = en.WriteString(z.Pilot[za0001])
		if err != nil {
			return
		}
	}
	// write "cancelled"
	err = en.Append(0xa9, 0x63, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x6c, 0x65, 0x64)
	if err != nil {
		return
	}
	err = en.WriteBool(z.Cancelled)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Event) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "id"
	o = append(o, 0x85, 0xa2, 0x69, 0x64)
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
	for za0001 := range z.Pilot {
		o = msgp.AppendString(o, z.Pilot[za0001])
	}
	// string "cancelled"
	o = append(o, 0xa9, 0x63, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x6c, 0x65, 0x64)
	o = msgp.AppendBool(o, z.Cancelled)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Event) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
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
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zb0002 > 0 {
				zb0002--
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
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Pilot) >= int(zb0003) {
				z.Pilot = (z.Pilot)[:zb0003]
			} else {
				z.Pilot = make([]string, zb0003)
			}
			for za0001 := range z.Pilot {
				z.Pilot[za0001], bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
			}
		case "cancelled":
			z.Cancelled, bts, err = msgp.ReadBoolBytes(bts)
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

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Event) Msgsize() (s int) {
	s = 1 + 3 + msgp.IntSize + 5 + 1 + 6 + msgp.StringPrefixSize + len(z.User.First) + 5 + msgp.StringPrefixSize + len(z.User.Last) + 7 + msgp.StringPrefixSize + len(z.Flight) + 6 + msgp.ArrayHeaderSize
	for za0001 := range z.Pilot {
		s += msgp.StringPrefixSize + len(z.Pilot[za0001])
	}
	s += 10 + msgp.BoolSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *NestInner) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "hello":
			z.Hello, err = dc.ReadString()
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
func (z NestInner) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "hello"
	err = en.Append(0x81, 0xa5, 0x68, 0x65, 0x6c, 0x6c, 0x6f)
	if err != nil {
		return
	}
	err = en.WriteString(z.Hello)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z NestInner) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "hello"
	o = append(o, 0x81, 0xa5, 0x68, 0x65, 0x6c, 0x6c, 0x6f)
	o = msgp.AppendString(o, z.Hello)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *NestInner) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "hello":
			z.Hello, bts, err = msgp.ReadStringBytes(bts)
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

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z NestInner) Msgsize() (s int) {
	s = 1 + 6 + msgp.StringPrefixSize + len(z.Hello)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *NestOuter) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "inner":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.Inner = nil
			} else {
				if z.Inner == nil {
					z.Inner = new(NestInner)
				}
				var zb0002 uint32
				zb0002, err = dc.ReadMapHeader()
				if err != nil {
					return
				}
				for zb0002 > 0 {
					zb0002--
					field, err = dc.ReadMapKeyPtr()
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "hello":
						z.Inner.Hello, err = dc.ReadString()
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
func (z *NestOuter) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "inner"
	err = en.Append(0x81, 0xa5, 0x69, 0x6e, 0x6e, 0x65, 0x72)
	if err != nil {
		return
	}
	if z.Inner == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		// map header, size 1
		// write "hello"
		err = en.Append(0x81, 0xa5, 0x68, 0x65, 0x6c, 0x6c, 0x6f)
		if err != nil {
			return
		}
		err = en.WriteString(z.Inner.Hello)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *NestOuter) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "inner"
	o = append(o, 0x81, 0xa5, 0x69, 0x6e, 0x6e, 0x65, 0x72)
	if z.Inner == nil {
		o = msgp.AppendNil(o)
	} else {
		// map header, size 1
		// string "hello"
		o = append(o, 0x81, 0xa5, 0x68, 0x65, 0x6c, 0x6c, 0x6f)
		o = msgp.AppendString(o, z.Inner.Hello)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *NestOuter) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "inner":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Inner = nil
			} else {
				if z.Inner == nil {
					z.Inner = new(NestInner)
				}
				var zb0002 uint32
				zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				for zb0002 > 0 {
					zb0002--
					field, bts, err = msgp.ReadMapKeyZC(bts)
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "hello":
						z.Inner.Hello, bts, err = msgp.ReadStringBytes(bts)
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

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *NestOuter) Msgsize() (s int) {
	s = 1 + 6
	if z.Inner == nil {
		s += msgp.NilSize
	} else {
		s += 1 + 6 + msgp.StringPrefixSize + len(z.Inner.Hello)
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Person) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
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
		return
	}
	err = en.WriteString(z.First)
	if err != nil {
		return
	}
	// write "last"
	err = en.Append(0xa4, 0x6c, 0x61, 0x73, 0x74)
	if err != nil {
		return
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
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
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

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Person) Msgsize() (s int) {
	s = 1 + 6 + msgp.StringPrefixSize + len(z.First) + 5 + msgp.StringPrefixSize + len(z.Last)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Weather) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "time":
			z.Time, err = dc.ReadTime()
			if err != nil {
				return
			}
		case "size":
			z.Size, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "type":
			z.Type, err = dc.ReadString()
			if err != nil {
				return
			}
		case "details":
			z.Details, err = dc.ReadBytes(z.Details)
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
func (z *Weather) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "time"
	err = en.Append(0x84, 0xa4, 0x74, 0x69, 0x6d, 0x65)
	if err != nil {
		return
	}
	err = en.WriteTime(z.Time)
	if err != nil {
		return
	}
	// write "size"
	err = en.Append(0xa4, 0x73, 0x69, 0x7a, 0x65)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.Size)
	if err != nil {
		return
	}
	// write "type"
	err = en.Append(0xa4, 0x74, 0x79, 0x70, 0x65)
	if err != nil {
		return
	}
	err = en.WriteString(z.Type)
	if err != nil {
		return
	}
	// write "details"
	err = en.Append(0xa7, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Details)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Weather) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "time"
	o = append(o, 0x84, 0xa4, 0x74, 0x69, 0x6d, 0x65)
	o = msgp.AppendTime(o, z.Time)
	// string "size"
	o = append(o, 0xa4, 0x73, 0x69, 0x7a, 0x65)
	o = msgp.AppendInt64(o, z.Size)
	// string "type"
	o = append(o, 0xa4, 0x74, 0x79, 0x70, 0x65)
	o = msgp.AppendString(o, z.Type)
	// string "details"
	o = append(o, 0xa7, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73)
	o = msgp.AppendBytes(o, z.Details)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Weather) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "time":
			z.Time, bts, err = msgp.ReadTimeBytes(bts)
			if err != nil {
				return
			}
		case "size":
			z.Size, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "type":
			z.Type, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "details":
			z.Details, bts, err = msgp.ReadBytesBytes(bts, z.Details)
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

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Weather) Msgsize() (s int) {
	s = 1 + 5 + msgp.TimeSize + 5 + msgp.Int64Size + 5 + msgp.StringPrefixSize + len(z.Type) + 8 + msgp.BytesPrefixSize + len(z.Details)
	return
}
