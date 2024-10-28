package bstd

import (
	"encoding/binary"
	"math"
	"strconv"
	"unsafe"

	"github.com/deneonet/benc"
	"golang.org/x/exp/constraints"
)

type MarshalFunc[T any] func(n int, b []byte, t T) int

// For unsafe string too
func SkipString(n int, b []byte) (int, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, err
	}
	s := int(us)

	if len(b)-n < s {
		return n, benc.ErrBufTooSmall
	}
	return n + s, nil
}

// For unsafe string too
func SizeString(str string) int {
	v := len(str)
	return v + SizeUint(uint(v))
}

func MarshalString(n int, b []byte, str string) int {
	n = MarshalUint(n, b, uint(len(str)))
	return n + copy(b[n:], str)
}

func UnmarshalString(n int, b []byte) (int, string, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, "", err
	}
	s := int(us)

	if len(b)-n < s {
		return n, "", benc.ErrBufTooSmall
	}
	return n + s, string(b[n : n+s]), nil
}

//

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

func MarshalUnsafeString(n int, b []byte, str string) int {
	n = MarshalUint(n, b, uint(len(str)))
	return n + copy(b[n:], s2b(str))
}

func UnmarshalUnsafeString(n int, b []byte) (int, string, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, "", err
	}
	s := int(us)
	if s == 0 {
		return n, "", nil
	}

	if len(b)-n < s {
		return n, "", benc.ErrBufTooSmall
	}
	return n + s, b2s(b[n : n+s]), nil
}

//

func SkipSlice(n int, b []byte) (int, error) {
	lb := len(b)

	for {
		if lb-n < 4 {
			return 0, benc.ErrBufTooSmall
		}

		if b[n] == 1 && b[n+1] == 1 && b[n+2] == 1 && b[n+3] == 1 {
			return n + 4, nil
		}
		n++
	}
}

func SizeSlice[T any](slice []T, sizer interface{}) (s int) {
	v := len(slice)
	s += 4 + SizeUint(uint(v))

	switch p := sizer.(type) {
	case func() int:
		for i := 0; i < v; i++ {
			s += p()
		}
	case func(T) int:
		for _, t := range slice {
			s += p(t)
		}
	default:
		panic("benc: invalid `sizer` provided in `SizeSlice`")
	}
	return
}

func MarshalSlice[T any](n int, b []byte, slice []T, marshaler MarshalFunc[T]) int {
	n = MarshalUint(n, b, uint(len(slice)))
	for _, t := range slice {
		n = marshaler(n, b, t)
	}

	u := b[n : n+4]
	_ = u[3]
	u[0] = byte(1)
	u[1] = byte(1)
	u[2] = byte(1)
	u[3] = byte(1)
	return n + 4
}

func UnmarshalSlice[T any](n int, b []byte, unmarshaler interface{}) (int, []T, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, nil, err
	}
	s := int(us)

	var t T
	ts := make([]T, s)

	switch p := unmarshaler.(type) {
	case func(n int, b []byte) (int, T, error):
		for i := 0; i < s; i++ {
			n, t, err = p(n, b)
			if err != nil {
				return 0, nil, err
			}

			ts[i] = t
		}
	case func(n int, b []byte, v *T) (int, error):
		for i := 0; i < s; i++ {
			n, err = p(n, b, &ts[i])
			if err != nil {
				return 0, nil, err
			}
		}
	default:
		panic("benc: invalid `unmarshaler` provided in `UnmarshalSlice`")
	}

	return n + 4, ts, nil
}

// SkipMap = SkipSlice
func SkipMap(n int, b []byte) (int, error) {
	lb := len(b)

	for {
		if lb-n < 4 {
			return 0, benc.ErrBufTooSmall
		}

		if b[n] == 1 && b[n+1] == 1 && b[n+2] == 1 && b[n+3] == 1 {
			return n + 4, nil
		}
		n++
	}
}

func SizeMap[K comparable, V any](m map[K]V, kSizer interface{}, vSizer interface{}) (s int) {
	s += 4 + SizeUint(uint(len(m)))

	for k, v := range m {
		switch p := kSizer.(type) {
		case func() int:
			s += p()
		case func(K) int:
			s += p(k)
		default:
			panic("benc: invalid `kSizer` provided in `SizeMap`")
		}

		switch p := vSizer.(type) {
		case func() int:
			s += p()
		case func(V) int:
			s += p(v)
		default:
			panic("benc: invalid `vSizer` provided in `SizeMap`")
		}
	}
	return
}

func MarshalMap[K comparable, V any](n int, b []byte, m map[K]V, kMarshaler MarshalFunc[K], vMarshaler MarshalFunc[V]) int {
	n = MarshalUint(n, b, uint(len(m)))
	for k, v := range m {
		n = kMarshaler(n, b, k)
		n = vMarshaler(n, b, v)
	}

	u := b[n : n+4]
	_ = u[3]
	u[0] = byte(1)
	u[1] = byte(1)
	u[2] = byte(1)
	u[3] = byte(1)
	return n + 4
}

func UnmarshalMap[K comparable, V any](n int, b []byte, kUnmarshaler interface{}, vUnmarshaler interface{}) (int, map[K]V, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, nil, err
	}
	s := int(us)

	var k K
	var v V
	ts := make(map[K]V, s)

	for i := 0; i < s; i++ {
		switch p := kUnmarshaler.(type) {
		case func(n int, b []byte) (int, K, error):
			n, k, err = p(n, b)
			if err != nil {
				return 0, nil, err
			}
		case func(n int, b []byte, k *K) (int, error):
			n, err = p(n, b, &k)
			if err != nil {
				return 0, nil, err
			}
		default:
			panic("benc: invalid `kUnmarshaler` provided in `UnmarshalMap`")
		}

		switch p := vUnmarshaler.(type) {
		case func(n int, b []byte) (int, V, error):
			n, v, err = p(n, b)
			if err != nil {
				return 0, nil, err
			}
		case func(n int, b []byte, v *V) (int, error):
			n, err = p(n, b, &v)
			if err != nil {
				return 0, nil, err
			}
		default:
			panic("benc: invalid `kUnmarshaler` provided in `UnmarshalMap`")
		}

		ts[k] = v
	}

	return n + 4, ts, nil
}

//

func SkipByte(n int, b []byte) (int, error) {
	if len(b)-n < 1 {
		return n, benc.ErrBufTooSmall
	}
	return n + 1, nil
}

func SizeByte() int {
	return 1
}

func MarshalByte(n int, b []byte, byt byte) int {
	b[n] = byt
	return n + 1
}

func UnmarshalByte(n int, b []byte) (int, byte, error) {
	if len(b)-n < 1 {
		return n, 0, benc.ErrBufTooSmall
	}
	return n + 1, b[n], nil
}

// SkipBytes = SkipString
func SkipBytes(n int, b []byte) (int, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, err
	}
	s := int(us)
	if len(b)-n < s {
		return n, benc.ErrBufTooSmall
	}
	return n + s, nil
}

func SizeBytes(bs []byte) int {
	v := len(bs)
	return v + SizeUint(uint(v))
}

func MarshalBytes(n int, b []byte, bs []byte) int {
	n = MarshalUint(n, b, uint(len(bs)))
	return n + copy(b[n:], bs)
}

func UnmarshalBytes(n int, b []byte) (int, []byte, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, nil, err
	}
	s := int(us)
	if len(b)-n < s {
		return 0, nil, benc.ErrBufTooSmall
	}
	return n + s, b[n : n+s], nil
}

var maxVarintLenMap = map[int]int{
	64: binary.MaxVarintLen64,
	32: binary.MaxVarintLen32,
}

var maxVarintLen = maxVarintLenMap[strconv.IntSize]

// Returns the new offset 'n' after skipping the marshalled varint.
//
// Possible errors returned:
//   - benc.ErrOverflow          - varint overflowed a N-bit unsigned integer.
//   - benc.ErrBufTooSmall       - 'buf' was too small to skip the marshalled varint.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipVarint(n int, buf []byte) (int, error) {
	for i, b := range buf[n:] {
		if i == maxVarintLen {
			return 0, benc.ErrOverflow
		}
		if b < 0x80 {
			if i == maxVarintLen-1 && b > 1 {
				return 0, benc.ErrOverflow
			}
			return n + i + 1, nil
		}
	}
	return 0, benc.ErrBufTooSmall
}

// Returns the bytes needed to marshal a integer.
func SizeInt(sv int) int {
	v := uint(encodeZigZag(sv))
	i := 0
	for v >= 0x80 {
		v >>= 7
		i++
	}
	return i + 1
}

// Returns the new offset 'n' after marshalling the integer.
//
// !- Panics, if 'b' is too small.
func MarshalInt(n int, b []byte, sv int) int {
	v := uint(encodeZigZag(sv))
	i := n
	for v >= 0x80 {
		b[i] = byte(v) | 0x80
		v >>= 7
		i++
	}
	b[i] = byte(v)
	return i + 1
}

// Returns the new offset 'n', as well as the integer, that got unmarshalled.
//
// Possible errors returned:
//   - benc.ErrOverflow          - varint overflowed a N-bit integer.
//   - benc.ErrBufTooSmall       - 'buf' was too small to skip the unmarshal the integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalInt(n int, buf []byte) (int, int, error) {
	var x uint
	var s uint
	for i, b := range buf[n:] {
		if i == maxVarintLen {
			return 0, 0, benc.ErrOverflow
		}
		if b < 0x80 {
			if i == maxVarintLen-1 && b > 1 {
				return 0, 0, benc.ErrOverflow
			}
			return n + i + 1, int(decodeZigZag(x | uint(b)<<s)), nil
		}
		x |= uint(b&0x7f) << s
		s += 7
	}
	return 0, 0, benc.ErrBufTooSmall
}

// Returns the bytes needed to marshal a unsigned integer.
func SizeUint(v uint) int {
	i := 0
	for v >= 0x80 {
		v >>= 7
		i++
	}
	return i + 1
}

// Returns the new offset 'n' after marshalling the unsigned integer.
//
// !- Panics, if 'b' is too small.
func MarshalUint(n int, b []byte, v uint) int {
	i := n
	for v >= 0x80 {
		b[i] = byte(v) | 0x80
		v >>= 7
		i++
	}
	b[i] = byte(v)
	return i + 1
}

// Returns the new offset 'n', as well as the unsigned integer, that got unmarshalled.
//
// Possible errors returned:
//   - benc.ErrOverflow          - varint overflowed a N-bit unsigned integer
//   - benc.ErrBufTooSmall       - 'buf' was too small to skip the unmarshal the unsigned integer//
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalUint(n int, buf []byte) (int, uint, error) {
	var x uint
	var s uint
	for i, b := range buf[n:] {
		if i == maxVarintLen {
			return 0, 0, benc.ErrOverflow
		}
		if b < 0x80 {
			if i == maxVarintLen-1 && b > 1 {
				return 0, 0, benc.ErrOverflow
			}
			return n + i + 1, x | uint(b)<<s, nil
		}
		x |= uint(b&0x7f) << s
		s += 7
	}
	return 0, 0, benc.ErrBufTooSmall
}

// Returns the new offset 'n' after skipping the marshalled 64-bit unsigned integer.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to skip the marshalled 64-bit unsigned integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipUint64(n int, b []byte) (int, error) {
	if len(b)-n < 8 {
		return n, benc.ErrBufTooSmall
	}
	return n + 8, nil
}

// Returns the bytes needed to marshal a 64-bit unsigned integer.
func SizeUint64() int {
	return 8
}

// Returns the new offset 'n' after marshalling the 64-bit unsigned integer.
func MarshalUint64(n int, b []byte, v uint64) int {
	u := b[n : n+8]
	_ = u[7]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	u[2] = byte(v >> 16)
	u[3] = byte(v >> 24)
	u[4] = byte(v >> 32)
	u[5] = byte(v >> 40)
	u[6] = byte(v >> 48)
	u[7] = byte(v >> 56)
	return n + 8
}

// Returns the new offset 'n', as well as the 64-bit unsigned integer, that got unmarshalled.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to unmarshal the 64-bit unsigned integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalUint64(n int, b []byte) (int, uint64, error) {
	if len(b)-n < 8 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, v, nil
}

// Returns the new offset 'n' after skipping the marshalled 32-bit unsigned integer.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to skip the marshalled 32-bit unsigned integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipUint32(n int, b []byte) (int, error) {
	if len(b)-n < 4 {
		return n, benc.ErrBufTooSmall
	}
	return n + 4, nil
}

// Returns the bytes needed to marshal a 32-bit unsigned integer.
func SizeUint32() int {
	return 4
}

// Returns the new offset 'n' after marshalling the 32-bit unsigned integer.
func MarshalUint32(n int, b []byte, v uint32) int {
	u := b[n : n+4]
	_ = u[3]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	u[2] = byte(v >> 16)
	u[3] = byte(v >> 24)
	return n + 4
}

// Returns the new offset 'n', as well as the 32-bit unsigned integer, that got unmarshalled.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to unmarshal the 32-bit unsigned integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalUint32(n int, b []byte) (int, uint32, error) {
	if len(b)-n < 4 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, v, nil
}

// Returns the new offset 'n' after skipping the marshalled 16-bit unsigned integer.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to skip the marshalled 16-bit unsigned integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipUint16(n int, b []byte) (int, error) {
	if len(b)-n < 2 {
		return n, benc.ErrBufTooSmall
	}
	return n + 2, nil
}

// Returns the bytes needed to marshal a 16-bit unsigned integer.
func SizeUint16() int {
	return 2
}

// Returns the new offset 'n' after marshalling the 16-bit unsigned integer.
func MarshalUint16(n int, b []byte, v uint16) int {
	u := b[n : n+2]
	_ = u[1]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	return n + 2
}

// Returns the new offset 'n', as well as the 16-bit unsigned integer, that got unmarshalled.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to unmarshal the 16-bit unsigned integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalUint16(n int, b []byte) (int, uint16, error) {
	if len(b)-n < 2 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, v, nil
}

// Returns the new offset 'n' after skipping the marshalled 64-bit integer.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to skip the marshalled 64-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipInt64(n int, b []byte) (int, error) {
	if len(b)-n < 8 {
		return n, benc.ErrBufTooSmall
	}
	return n + 8, nil
}

// Returns the bytes needed to marshal a 64-bit integer.
func SizeInt64() int {
	return 8
}

// Returns the new offset 'n' after marshalling the 64-bit integer.
func MarshalInt64(n int, b []byte, v int64) int {
	v64 := uint64(encodeZigZag(v))
	u := b[n : n+8]
	_ = u[7]
	u[0] = byte(v64)
	u[1] = byte(v64 >> 8)
	u[2] = byte(v64 >> 16)
	u[3] = byte(v64 >> 24)
	u[4] = byte(v64 >> 32)
	u[5] = byte(v64 >> 40)
	u[6] = byte(v64 >> 48)
	u[7] = byte(v64 >> 56)
	return n + 8
}

// Returns the new offset 'n', as well as the 64-bit integer, that got unmarshalled.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to unmarshal the 64-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalInt64(n int, b []byte) (int, int64, error) {
	if len(b)-n < 8 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, int64(decodeZigZag(v)), nil
}

// Returns the new offset 'n' after skipping the marshalled 32-bit integer.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to skip the marshalled 32-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipInt32(n int, b []byte) (int, error) {
	if len(b)-n < 4 {
		return n, benc.ErrBufTooSmall
	}
	return n + 4, nil
}

// Returns the bytes needed to marshal a 32-bit integer.
func SizeInt32() int {
	return 4
}

// Returns the new offset 'n' after marshalling the 32-bit integer.
func MarshalInt32(n int, b []byte, v int32) int {
	v32 := uint32(encodeZigZag(v))
	u := b[n : n+4]
	_ = u[3]
	u[0] = byte(v32)
	u[1] = byte(v32 >> 8)
	u[2] = byte(v32 >> 16)
	u[3] = byte(v32 >> 24)
	return n + 4
}

// Returns the new offset 'n', as well as the 32-bit integer, that got unmarshalled.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to unmarshal the 32-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalInt32(n int, b []byte) (int, int32, error) {
	if len(b)-n < 4 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, int32(decodeZigZag(v)), nil
}

// Returns the new offset 'n' after skipping the marshalled 16-bit integer.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to skip the marshalled 16-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipInt16(n int, b []byte) (int, error) {
	if len(b)-n < 2 {
		return n, benc.ErrBufTooSmall
	}
	return n + 2, nil
}

// Returns the bytes needed to marshal a 16-bit integer.
func SizeInt16() int {
	return 2
}

// Returns the new offset 'n' after marshalling the 16-bit integer.
func MarshalInt16(n int, b []byte, v int16) int {
	v16 := uint16(encodeZigZag(v))
	u := b[n : n+2]
	_ = u[1]
	u[0] = byte(v16)
	u[1] = byte(v16 >> 8)
	return n + 2
}

// Returns the new offset 'n', as well as the 16-bit integer, that got unmarshalled.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to unmarshal the 16-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalInt16(n int, b []byte) (int, int16, error) {
	if len(b)-n < 2 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, int16(decodeZigZag(v)), nil
}

// TODO: Int8

// Returns the new offset 'n' after skipping the marshalled 64-bit float.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to skip the marshalled 64-bit float.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipFloat64(n int, b []byte) (int, error) {
	if len(b)-n < 8 {
		return n, benc.ErrBufTooSmall
	}
	return n + 8, nil
}

// Returns the bytes needed to marshal a 64-bit float.
func SizeFloat64() int {
	return 8
}

// Returns the new offset 'n' after marshalling the 64-bit float.
func MarshalFloat64(n int, b []byte, v float64) int {
	v64 := math.Float64bits(v)
	u := b[n : n+8]
	_ = u[7]
	u[0] = byte(v64)
	u[1] = byte(v64 >> 8)
	u[2] = byte(v64 >> 16)
	u[3] = byte(v64 >> 24)
	u[4] = byte(v64 >> 32)
	u[5] = byte(v64 >> 40)
	u[6] = byte(v64 >> 48)
	u[7] = byte(v64 >> 56)
	return n + 8
}

// Returns the new offset 'n', as well as the 64-bit float, that got unmarshalled.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to unmarshal the 64-bit float.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalFloat64(n int, b []byte) (int, float64, error) {
	if len(b)-n < 8 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, math.Float64frombits(v), nil
}

// Returns the new offset 'n' after skipping the marshalled 32-bit float.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to skip the marshalled 32-bit float.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipFloat32(n int, b []byte) (int, error) {
	if len(b)-n < 4 {
		return n, benc.ErrBufTooSmall
	}
	return n + 4, nil
}

// Returns the bytes needed to marshal a 32-bit float.
func SizeFloat32() int {
	return 4
}

// Returns the new offset 'n' after marshalling the 32-bit float.
func MarshalFloat32(n int, b []byte, v float32) int {
	v32 := math.Float32bits(v)
	u := b[n : n+4]
	_ = u[3]
	u[0] = byte(v32)
	u[1] = byte(v32 >> 8)
	u[2] = byte(v32 >> 16)
	u[3] = byte(v32 >> 24)
	return n + 4
}

// Returns the new offset 'n', as well as the 32-bit float, that got unmarshalled.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to unmarshal the 32-bit float.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalFloat32(n int, b []byte) (int, float32, error) {
	if len(b)-n < 4 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, math.Float32frombits(v), nil
}

// Returns the new offset 'n' after skipping the marshalled bool.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to skip the marshalled bool.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipBool(n int, b []byte) (int, error) {
	if len(b)-n < 1 {
		return 0, benc.ErrBufTooSmall
	}
	return n + 1, nil
}

// Returns the bytes needed to marshal a bool.
func SizeBool() int {
	return 1
}

// Returns the new offset 'n' after marshalling the bool.
func MarshalBool(n int, b []byte, v bool) int {
	var i byte
	if v {
		i = 1
	}
	b[n] = i
	return n + 1
}

// Returns the new offset 'n', as well as the bool, that got unmarshalled.
//
// Possible errors returned:
//   - benc.ErrBufTooSmall       - 'b' was too small to unmarshal the bool.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalBool(n int, b []byte) (int, bool, error) {
	if len(b)-n < 1 {
		return 0, false, benc.ErrBufTooSmall
	}
	return n + 1, uint8(b[n]) == 1, nil
}

func encodeZigZag[T constraints.Signed](t T) T {
	if t < 0 {
		return ^(t << 1)
	}
	return t << 1
}

func decodeZigZag[T constraints.Unsigned](t T) T {
	if t&1 == 1 {
		return ^(t >> 1)
	}
	return t >> 1
}
