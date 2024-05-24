package bunsafe

import (
	"errors"
	"github.com/deneonet/benc/bmd"
	"unsafe"
)

var ErrBytesToSmall = errors.New("insufficient data, given bytes are too small")

type StringHeader struct {
	Data *byte
	Len  int
}

// b2s converts byte slice to a string without memory allocation.
//
// Previously used: -- return *(*string)(unsafe.Pointer(&b))
//
// Removed because reflect.SliceHeader is deprecated, so I use unsafe.String
// see https://github.com/golang/go/issues/53003
func b2s(b []byte) string {
	return unsafe.String(&b[0], len(b))
}

// s2b converts string to a byte slice without memory allocation.
//
// Previously used: -- *(*[]byte)(unsafe.Pointer(&s))
//
// Removed because of: https://github.com/golang/go/issues/53003
// +
// because reflect.StringHeader is deprecated, so I use a new StringHeader type
// see https://github.com/golang/go/issues/53003#issuecomment-1145241692
func s2b(s string) []byte {
	header := (*StringHeader)(unsafe.Pointer(&s))
	bytes := *(*[]byte)(unsafe.Pointer(header))
	return bytes
}

func MarshalString(n int, b []byte, str string) int {
	v := uint16(len(str))
	u := b[n:]
	_ = u[1]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	return n + 2 + copy(b[n+2:], s2b(str))
}

func UnmarshalString(n int, b []byte) (int, string, error) {
	if len(b)-n == 2 {
		return n + 2, "", nil
	} else if len(b)-n < 2 {
		return n, "", ErrBytesToSmall
	}
	u := b[n:]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	bs := b[n : n+size]
	return n + size, b2s(bs), nil
}

func MarshalStringMD(n int, b []byte, str string) int {
	b[n] = bmd.String
	n++
	v := uint16(len(str))
	u := b[n:]
	_ = u[1]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	return n + 2 + copy(b[n+2:], s2b(str))
}

func UnmarshalStringMD(n int, b []byte) (int, string, error) {
	if len(b)-n == 3 {
		return n + 3, "", nil
	} else if len(b)-n < 3 {
		return n, "", ErrBytesToSmall
	}
	if b[n] != bmd.String {
		return n, "", errors.New("expected a bunsafe string, found something else. check your marshal process")
	}
	n++
	u := b[n:]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	bs := b[n : n+size]
	return n + size, b2s(bs), nil
}
