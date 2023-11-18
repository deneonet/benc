package btag

import (
	"github.com/deneonet/benc/bpre"
)

func SMarshalMF(s int, tag string) (int, []byte) {
	l := len(tag)
	b := bpre.GetMarshal(s + 4 + l)
	v := uint16(s)
	_ = b[1]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	v = uint16(l)
	u := b[2:]
	_ = u[1]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	return 4 + copy(b[4:], tag), b
}

func SMarshal(s int, tag string) (int, []byte) {
	v := uint16(len(tag))
	b := bpre.GetMarshal(s + 2 + int(v))
	_ = b[1]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	return 2 + copy(b[2:], tag), b
}

func UMarshalMF(s int, tag uint16) (int, []byte) {
	b := bpre.GetMarshal(s + 4)
	v := uint16(s)
	_ = b[1]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	u := b[2:]
	_ = u[1]
	u[0] = byte(tag)
	u[1] = byte(tag >> 8)
	return 4, b
}

func UMarshal(s int, tag uint16) (int, []byte) {
	b := bpre.GetMarshal(s + 2)
	_ = b[1]
	b[0] = byte(tag)
	b[1] = byte(tag >> 8)
	return 2, b
}
