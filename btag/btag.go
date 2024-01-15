package btag

import (
	"errors"
	"github.com/deneonet/benc/bmd"
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

func SMarshalMD(s int, tag string) (int, []byte) {
	v := uint16(len(tag))
	b := bpre.GetMarshal(s + 3 + int(v))
	b[0] = bmd.StringTag
	_ = b[2]
	b[1] = byte(v)
	b[2] = byte(v >> 8)
	return 3 + copy(b[3:], tag), b
}

func UMarshalMD(s int, tag uint16) (int, []byte) {
	b := bpre.GetMarshal(s + 3)
	b[0] = bmd.UIntTag
	_ = b[2]
	b[1] = byte(tag)
	b[2] = byte(tag >> 8)
	return 3, b
}

func SUnmarshalMD(n int, b []byte) (int, string, error) {
	if n != 0 {
		return 0, "", bmd.ErrNIsNotZero
	}
	if len(b)-n < 3 {
		return n, "", bmd.ErrBytesToSmall
	}
	if b[n] != bmd.StringTag {
		return n, "", errors.New("expected a string tag, found something else. check your marshal process")
	}
	n += 1
	u := b[n : n+2]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	bs := b[n : n+size]
	return n + size, string(bs), nil
}

func UUnmarshalMD(n int, b []byte) (int, uint16, error) {
	if n != 0 {
		return 0, 0, bmd.ErrNIsNotZero
	}
	if len(b)-n < 3 {
		return n, 0, bmd.ErrBytesToSmall
	}
	if b[n] != bmd.UIntTag {
		return n, 0, errors.New("expected a uint tag, found something else. check your marshal process")
	}
	n += 1
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, v, nil
}

func SUnmarshal(n int, b []byte) (int, string, error) {
	if n != 0 {
		return 0, "", bmd.ErrNIsNotZero
	}
	if len(b)-n < 2 {
		return n, "", bmd.ErrBytesToSmall
	}
	u := b[n : n+2]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	bs := b[n : n+size]
	return n + size, string(bs), nil
}

func UUnmarshal(n int, b []byte) (int, uint16, error) {
	if n != 0 {
		return 0, 0, bmd.ErrNIsNotZero
	}
	if len(b)-n < 2 {
		return n, 0, bmd.ErrBytesToSmall
	}
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, v, nil
}
