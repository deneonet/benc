package bgenimpl

import (
	"errors"
	"slices"

	"github.com/deneonet/benc"
	bstd "github.com/deneonet/benc/std"
)

var ErrEof = errors.New("reached end of decoding")
var ErrInvalidType = errors.New("the type decoded is invalid")

const (
	Container byte = iota + 2
	Bytes
	Varint
	Fixed8
	Fixed16
	Fixed32
	Fixed64
	ArrayMap
)

func skipByType(tn int, b []byte, t byte) (n int, err error) {
	n = tn
	switch t {
	case Bytes:
		n, err = bstd.SkipBytes(n, b)
	case ArrayMap:
		n, err = bstd.SkipSlice(n, b)
	case Varint:
		n, err = bstd.SkipVarint(n, b)
	case Container:
		for b[n] != 1 || b[n+1] != 1 {
			n, _, t, err = UnmarshalTag(n, b)
			if err != nil {
				return
			}

			n, err = skipByType(n, b, t)
			if err != nil {
				return
			}
		}
		n += 2
	case Fixed8:
		n += 1
	case Fixed16:
		n += 2
	case Fixed32:
		n += 4
	case Fixed64:
		n += 8
	default:
		err = ErrInvalidType
	}
	return
}

func HandleCompatibility(n int, b []byte, r []uint16, id uint16) (int, bool, error) {
	n, tId, typ, err := UnmarshalTag(n, b)
	if err != nil {
		return 0, false, ErrEof
	}

	for tId != id {
		if slices.Contains(r, tId) {
			n, err = skipByType(n, b, typ)
			if err != nil {
				return 0, false, err
			}

			n, tId, typ, err = UnmarshalTag(n, b)
			if err != nil {
				return 0, false, ErrEof
			}

			continue
		}

		if tId > 255 {
			return n - 3, false, nil
		}

		return n - 2, false, nil
	}

	return n, true, nil
}

func SkipTag(n int, b []byte) (int, error) {
	lb := len(b)
	if lb-n < 2 {
		return 0, benc.ErrBufTooSmall
	}

	l := b[n]&0x80 != 0
	n += 2

	if l {
		if lb-n < 1 {
			return 0, benc.ErrBufTooSmall
		}

		return n + 1, nil
	}
	return n, nil
}

func MarshalTag(n int, b []byte, t byte, id uint16) int {
	var c uint8
	if id > 255 {
		c |= 0x80
	}

	b[n] = c | t&0x7F
	n++

	if id > 255 {
		b[n] = byte(id >> 8)
		b[n+1] = byte(id & 0xFF)
		return n + 2
	}

	b[n] = byte(id)
	return n + 1
}

func UnmarshalTag(n int, b []byte) (int, uint16, byte, error) {
	lb := len(b)
	if lb-n < 2 {
		return 0, 0, 0, benc.ErrBufTooSmall
	}

	l := b[n]&0x80 != 0
	typ := b[n] & 0x7F
	n += 2

	if l {
		if lb-n < 1 {
			return 0, 0, 0, benc.ErrBufTooSmall
		}
		return n + 1, uint16(b[n-1])<<8 | uint16(b[n]), typ, nil
	}
	return n, uint16(b[n-1]), typ, nil
}
