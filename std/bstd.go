package bstd

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"

	"go.kine.bz/benc"
	"golang.org/x/exp/constraints"
)

type SkipFunc func(n int, b []byte) (int, error)
type MarshalFunc[T any] func(n int, b []byte, t T) int
type UnmarshalFunc[T any] func(n int, b []byte) (int, T, error)

// For unsafe string too
func SkipString(n int, b []byte) (int, error) {
	n, us, err := UnmarshalUVarint(n, b)
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
	return v + SizeUVarint(uint64(v))
}

func MarshalString(n int, b []byte, str string) int {
	n = MarshalUVarint(n, b, uint64(len(str)))
	return n + copy(b[n:], str)
}

func UnmarshalString(n int, b []byte) (int, string, error) {
	n, us, err := UnmarshalUVarint(n, b)
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
	n = MarshalUVarint(n, b, uint64(len(str)))
	return n + copy(b[n:], s2b(str))
}

func UnmarshalUnsafeString(n int, b []byte) (int, string, error) {
	n, us, err := UnmarshalUVarint(n, b)
	if err != nil {
		return 0, "", err
	}
	s := int(us)

	if len(b)-n < s {
		return n, "", benc.ErrBufTooSmall
	}
	return n + s, b2s(b[n : n+s]), nil
}

//

func SkipSlice(n int, b []byte) (int, error) {
	if len(b)-n < 4 {
		return 0, benc.ErrBufTooSmall
	}

	u := b[n : n+4]
	_ = u[3]
	return n + int(uint32(u[0])|uint32(u[1])<<8|uint32(u[2])<<16|uint32(u[3])<<24) + 4, nil
}

func SizeSlice[T any](slice []T, sizer interface{}) (s int) {
	v := len(slice)
	s += 4 + SizeUVarint(uint64(v))

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
		panic("[benc " + benc.BencVersion + "]: invalid `sizer` provided in `SizeSlice`")
	}
	return
}

func MarshalSlice[T any](n int, b []byte, slice []T, marshaler MarshalFunc[T]) int {
	u := b[n : n+4]
	n += 4

	sn := n
	n = MarshalUVarint(n, b, uint64(len(slice)))
	for _, t := range slice {
		n = marshaler(n, b, t)
	}

	_ = u[3]
	v32 := uint32(sn - n)
	u[0] = byte(v32)
	u[1] = byte(v32 >> 8)
	u[2] = byte(v32 >> 16)
	u[3] = byte(v32 >> 24)
	return n
}

func UnmarshalSlice[T any](n int, b []byte, unmarshaler interface{}) (int, []T, error) {
	n, us, err := UnmarshalUVarint(n+4, b)
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
				return 0, nil, fmt.Errorf("at index %d: %s", i, err.Error())
			}
		}
	default:
		panic("[benc " + benc.BencVersion + "]: invalid `unmarshaler` provided in `UnmarshalSlice`")
	}

	return n, ts, nil
}

// SkipMap = SkipSlice
func SkipMap(n int, b []byte, kSkipper SkipFunc, vSkipper SkipFunc) (int, error) {
	if len(b)-n < 4 {
		return 0, benc.ErrBufTooSmall
	}

	u := b[n : n+4]
	_ = u[3]
	return n + int(uint32(u[0])|uint32(u[1])<<8|uint32(u[2])<<16|uint32(u[3])<<24) + 4, nil
}

func SizeMap[K comparable, V any](m map[K]V, kSizer interface{}, vSizer interface{}) (s int) {
	s += 4 + SizeUVarint(uint64(len(m)))

	for k, v := range m {
		switch p := kSizer.(type) {
		case func() int:
			s += p()
		case func(K) int:
			s += p(k)
		default:
			panic("[benc " + benc.BencVersion + "]: invalid `kSizer` provided in `SizeMap`")
		}

		switch p := vSizer.(type) {
		case func() int:
			s += p()
		case func(V) int:
			s += p(v)
		default:
			panic("[benc " + benc.BencVersion + "]: invalid `vSizer` provided in `SizeMap`")
		}
	}
	return s + SizeUVarint(uint64(len(m)))
}

func MarshalMap[K comparable, V any](n int, b []byte, m map[K]V, kMarshaler MarshalFunc[K], vMarshaler MarshalFunc[V]) int {
	u := b[n : n+4]
	n += 4

	sn := n
	n = MarshalUVarint(n, b, uint64(len(m)))
	for k, v := range m {
		n = kMarshaler(n, b, k)
		n = vMarshaler(n, b, v)
	}

	_ = u[3]
	v32 := uint32(sn - n)
	u[0] = byte(v32)
	u[1] = byte(v32 >> 8)
	u[2] = byte(v32 >> 16)
	u[3] = byte(v32 >> 24)
	return n
}

func UnmarshalMap[K comparable, V any](n int, b []byte, kUnmarshaler interface{}, vUnmarshaler interface{}) (int, map[K]V, error) {
	n, us, err := UnmarshalUVarint(n+4, b)
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
				return 0, nil, fmt.Errorf("(key) at index %d: %s", i, err.Error())
			}
		case func(n int, b []byte, k *K) (int, error):
			n, err = p(n, b, &k)
			if err != nil {
				return 0, nil, fmt.Errorf("(key) at index %d: %s", i, err.Error())
			}
		default:
			panic("[benc " + benc.BencVersion + "]: invalid `kUnmarshaler` provided in `UnmarshalMap`")
		}

		switch p := vUnmarshaler.(type) {
		case func(n int, b []byte) (int, V, error):
			n, v, err = p(n, b)
			if err != nil {
				return 0, nil, fmt.Errorf("(value) at index %d: %s", i, err.Error())
			}
		case func(n int, b []byte, v *V) (int, error):
			n, err = p(n, b, &v)
			if err != nil {
				return 0, nil, fmt.Errorf("(value) at index %d: %s", i, err.Error())
			}
		default:
			panic("[benc " + benc.BencVersion + "]: invalid `kUnmarshaler` provided in `UnmarshalMap`")
		}

		ts[k] = v
	}

	return n, ts, nil
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
	n, us, err := UnmarshalUVarint(n, b)
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
	return v + SizeUVarint(uint64(v))
}

func MarshalBytes(n int, b []byte, bs []byte) int {
	n = MarshalUVarint(n, b, uint64(len(bs)))
	return n + copy(b[n:], bs)
}

func UnmarshalBytes(n int, b []byte) (int, []byte, error) {
	n, us, err := UnmarshalUVarint(n, b)
	if err != nil {
		return 0, nil, err
	}
	s := int(us)
	if len(b)-n < s {
		return 0, nil, benc.ErrBufTooSmall
	}
	return n + s, b[n : n+s], nil
}

//

func SkipUInt64(n int, b []byte) (int, error) {
	if len(b)-n < 8 {
		return n, benc.ErrBufTooSmall
	}
	return n + 8, nil
}

func SizeUInt64() int {
	return 8
}

func MarshalUInt64(n int, b []byte, v uint64) int {
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

func UnmarshalUInt64(n int, b []byte) (int, uint64, error) {
	if len(b)-n < 8 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, v, nil
}

//

func SkipUInt32(n int, b []byte) (int, error) {
	if len(b)-n < 4 {
		return n, benc.ErrBufTooSmall
	}
	return n + 4, nil
}

func SizeUInt32() int {
	return 4
}

func MarshalUInt32(n int, b []byte, v uint32) int {
	u := b[n : n+4]
	_ = u[3]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	u[2] = byte(v >> 16)
	u[3] = byte(v >> 24)
	return n + 4
}

func UnmarshalUInt32(n int, b []byte) (int, uint32, error) {
	if len(b)-n < 4 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, v, nil
}

//

func SkipUInt16(n int, b []byte) (int, error) {
	if len(b)-n < 2 {
		return n, benc.ErrBufTooSmall
	}
	return n + 2, nil
}

func SizeUInt16() int {
	return 2
}

func MarshalUInt16(n int, b []byte, v uint16) int {
	u := b[n : n+2]
	_ = u[1]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	return n + 2
}

func UnmarshalUInt16(n int, b []byte) (int, uint16, error) {
	if len(b)-n < 2 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, v, nil
}

//

func SkipInt64(n int, b []byte) (int, error) {
	if len(b)-n < 8 {
		return n, benc.ErrBufTooSmall
	}
	return n + 8, nil
}

func SizeInt64() int {
	return 8
}

func MarshalInt64(n int, b []byte, v int64) int {
	v64 := uint64(EncodeZigZag(v))
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

func UnmarshalInt64(n int, b []byte) (int, int64, error) {
	if len(b)-n < 8 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, int64(DecodeZigZag(v)), nil
}

//

func SkipInt32(n int, b []byte) (int, error) {
	if len(b)-n < 4 {
		return n, benc.ErrBufTooSmall
	}
	return n + 4, nil
}

func SizeInt32() int {
	return 4
}

func MarshalInt32(n int, b []byte, v int32) int {
	v32 := uint32(EncodeZigZag(v))
	u := b[n : n+4]
	_ = u[3]
	u[0] = byte(v32)
	u[1] = byte(v32 >> 8)
	u[2] = byte(v32 >> 16)
	u[3] = byte(v32 >> 24)
	return n + 4
}

func UnmarshalInt32(n int, b []byte) (int, int32, error) {
	if len(b)-n < 4 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, int32(DecodeZigZag(v)), nil
}

//

func SkipInt16(n int, b []byte) (int, error) {
	if len(b)-n < 2 {
		return n, benc.ErrBufTooSmall
	}
	return n + 2, nil
}

func SizeInt16() int {
	return 2
}

func MarshalInt16(n int, b []byte, v int16) int {
	v16 := uint16(EncodeZigZag(v))
	u := b[n : n+2]
	_ = u[1]
	u[0] = byte(v16)
	u[1] = byte(v16 >> 8)
	return n + 2
}

func UnmarshalInt16(n int, b []byte) (int, int16, error) {
	if len(b)-n < 2 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, int16(DecodeZigZag(v)), nil
}

//

func SkipFloat64(n int, b []byte) (int, error) {
	if len(b)-n < 8 {
		return n, benc.ErrBufTooSmall
	}
	return n + 8, nil
}

func SizeFloat64() int {
	return 8
}

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

//

func SkipFloat32(n int, b []byte) (int, error) {
	if len(b)-n < 4 {
		return n, benc.ErrBufTooSmall
	}
	return n + 4, nil
}

func SizeFloat32() int {
	return 4
}

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

func UnmarshalFloat32(n int, b []byte) (int, float32, error) {
	if len(b)-n < 4 {
		return n, 0, benc.ErrBufTooSmall
	}
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, math.Float32frombits(v), nil
}

//

func SkipBool(n int, b []byte) (int, error) {
	if len(b)-n < 1 {
		return n, benc.ErrBufTooSmall
	}
	return n + 1, nil
}

func SizeBool() int {
	return 1
}

func MarshalBool(n int, b []byte, v bool) int {
	var i byte
	if v {
		i = 1
	}
	b[n] = i
	return n + 1
}

func UnmarshalBool(n int, b []byte) (int, bool, error) {
	if len(b)-n < 1 {
		return n, false, benc.ErrBufTooSmall
	}
	return n + 1, uint8(b[n]) == 1, nil
}

//

func SkipUVarint(n int, buf []byte) (int, error) {
	for i, b := range buf[n:] {
		if i == binary.MaxVarintLen64 {
			return n, benc.ErrOverflow
		}
		if b < 0x80 {
			if i == binary.MaxVarintLen64-1 && b > 1 {
				return n, benc.ErrOverflow
			}
			return i + 1, nil
		}
	}
	return n, benc.ErrBufTooSmall
}

func SizeUVarint(v uint64) int {
	i := 0
	for v >= 0x80 {
		v >>= 7
		i++
	}
	return i + 1
}

func MarshalUVarint(n int, b []byte, v uint64) int {
	i := n
	for v >= 0x80 {
		b[i] = byte(v) | 0x80
		v >>= 7
		i++
	}
	b[i] = byte(v)
	return i + 1
}

func UnmarshalUVarint(n int, buf []byte) (int, uint64, error) {
	if len(buf)-n < 1 {
		return 0, 0, benc.ErrBufTooSmall
	}

	var x uint64
	var s uint
	for i, b := range buf[n:] {
		if i == binary.MaxVarintLen64 {
			return n, 0, benc.ErrOverflow
		}
		if b < 0x80 {
			if i == binary.MaxVarintLen64-1 && b > 1 {
				return n, 0, benc.ErrOverflow
			}
			return n + i + 1, x | uint64(b)<<s, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return n, 0, benc.ErrBufTooSmall
}

//

func EncodeZigZag[T constraints.Signed](t T) T {
	if t < 0 {
		return ^(t << 1)
	}
	return t << 1
}

func DecodeZigZag[T constraints.Unsigned](t T) T {
	if t&1 == 1 {
		return ^(t >> 1)
	}
	return t >> 1
}
