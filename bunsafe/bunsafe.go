package bunsafe

import (
	"errors"
	"unsafe"
)

var ErrBytesToSmall = errors.New("insufficient data, given bytes are too small")

//
// From:
// https://gist.github.com/yakuter/c0df0f4253ea639529f3589e99dc940b
//
//

// b2s converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func b2s(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

// s2b converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func s2b(s string) []byte {
	str := *(*[]byte)(unsafe.Pointer(&s))
	return str
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
	if len(b)-n < 2 {
		return n, "", ErrBytesToSmall
	}
	u := b[n:]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	bs := b[n : n+size]
	return n + size, b2s(bs), nil
}
