package bgenimpl

import (
	"errors"
	"slices"

	"go.kine.bz/benc"
	bstd "go.kine.bz/benc/std"
)

var ErrEof = errors.New("eof: reached end of decoding")
var ErrInvalidType = errors.New("the type decoded is invalid, and can't be used")

const (
	Container byte = iota + 2
	Bytes
	Fixed8
	Fixed16
	Fixed32
	Fixed64
)

func skipByType(tn int, b []byte, t byte) (n int, err error) {
	n = tn
	switch t {
	case Bytes:
		n, err = bstd.SkipBytes(n, b)
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
	if lb-n < 1 {
		return n, benc.ErrBufTooSmall
	}

	l := b[n]&0x80 != 0
	n++

	nToSkip := 1
	if l {
		nToSkip = 2
	}

	if lb-n < nToSkip {
		return n, benc.ErrBufTooSmall
	}
	return n + nToSkip, nil
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
	if lb-n < 1 {
		return n, 0, 0, benc.ErrBufTooSmall
	}

	l := b[n]&0x80 != 0
	typ := b[n] & 0x7F
	n++

	nToSkip := 1
	if l {
		nToSkip = 2
	}
	if lb-n < nToSkip {
		return n, 0, 0, benc.ErrBufTooSmall
	}

	if l {
		return n + 2, uint16(b[n])<<8 | uint16(b[n+1]), typ, nil
	}
	return n + 1, uint16(b[n]), typ, nil
}
