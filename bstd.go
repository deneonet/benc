package bstd

import (
	"errors"
	"math"
	"time"

	"github.com/deneonet/benc/bpre"
	"golang.org/x/exp/constraints"
)

var ErrBytesToSmall = errors.New("insufficient data, given bytes are too small")
var ErrNIsNotZero = errors.New("n has to be 0")

func UnmarshalMF(b []byte) ([][]byte, error) {
	dec := bpre.GetUnmarshalMF(len(b))
	var n uint16
	var i int

	if dec == nil {
		var dec [][]byte
		for {
			if 2 > len(b[n:]) {
				return dec, nil
			}
			u := b[n : n+2]
			_ = u[1]
			size := uint16(u[0]) | uint16(u[1])<<8
			n += 2
			if int(size) > len(b[n:]) {
				return nil, ErrBytesToSmall
			}
			dec = append(dec, b[n:n+size])
			n += size
		}
	}

	for i = 0; i < len(b); i++ {
		if 2 > len(b[n:]) {
			return dec[:i], nil
		}
		u := b[n : n+2]
		_ = u[1]
		size := uint16(u[0]) | uint16(u[1])<<8
		n += 2
		if int(size) > len(b[n:]) {
			return nil, ErrBytesToSmall
		}
		dec[i] = b[n : n+size]
		n += size
	}

	return dec[:i], nil
}

func MarshalMF(s int) (int, []byte) {
	b := bpre.GetMarshal(s + 2)
	v := uint16(s)
	_ = b[1]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	return 2, b
}

func Marshal(s int) (int, []byte) {
	return 0, bpre.GetMarshal(s)
}

type SizeFunc[T any] func(t T) int
type SkipFunc func(n int, b []byte) (int, error)
type MarshalFunc[T any] func(n int, b []byte, t T) int
type UnmarshalFunc[T any] func(n int, b []byte) (int, T, error)

/*
bencVersion is the version this code was written in
*/
const bencVersion = "v1.0.7"

type MaxSize int

const (
	MaxSizeUint16 MaxSize = iota
	MaxSizeUint32
	MaxSizeInt64
)

/*
maxSizeToByteSize returns the byte size of the max size (uint16, uint32, uint64)
*/
func maxSizeToByteSize(ms MaxSize) int {
	switch ms {
	case MaxSizeUint16:
		return 2
	case MaxSizeUint32:
		return 4
	case MaxSizeInt64:
		return 8
	}

	panic("[benc " + bencVersion + "]: MaxSizeToByteSize: ms is invalid!")
}

/*
verifySize verifies that l is smaller or equal to ms and returns the result
*/
func verifySize(l int, ms MaxSize) bool {
	switch ms {
	case MaxSizeUint16:
		return l <= math.MaxUint16
	case MaxSizeUint32:
		return l <= math.MaxUint32
	case MaxSizeInt64:
		return true
	}

	return false
}

/*
SizeSlice returns the size of a slice in the benc format.

sizer: the corresponding benc size function for the slice's type! ms: the maximum slice size!

# Panics:

..if slice is too big for ms, ...if sizer is incorrect

# Example:

	slice := make([]int, math.MaxUint16 + 1)
	slice[0] = 1
	slice[len(slice)-1] = 1

	size := SizeSlice(slice, SizeInt, 4)
*/
func SizeSlice[T any](slice []T, sizer interface{}, ms ...MaxSize) int {
	maxs := MaxSizeUint16
	if len(ms) != 0 {
		maxs = ms[0]
	}

	s := maxSizeToByteSize(maxs)
	if !verifySize(len(slice), maxs) {
		panic("[benc " + bencVersion + "]: SizeSlice: length of slice is too big!")
	}

	for _, t := range slice {
		if p, ok := sizer.(func(t T) int); ok {
			s += p(t)
		} else if p, ok := sizer.(func() int); ok {
			s += p()
		} else {
			panic("[benc " + bencVersion + "]: SizeSlice: sizer is invalid!")
		}
	}

	return s
}

/*
MarshalSlice encodes a slice in the benc format.

marshal: the corresponding benc marshal function for the slice's type! ms: the maximum slice size!

# Panics:

....if slice is too big for ms

# Example:

	n = MarshalSlice(n, buf, myIntSlice, bstd.MarshalInt, bstd.MaxSizeUint16)
*/
func MarshalSlice[T any](n int, b []byte, slice []T, marshal MarshalFunc[T], ms ...MaxSize) int {
	maxs := MaxSizeUint16
	if len(ms) != 0 {
		maxs = ms[0]
	}

	s := maxSizeToByteSize(maxs)
	if !verifySize(len(slice), maxs) {
		panic("[benc " + bencVersion + "]: MarshalSlice: length of slice is too big!")
	}

	v := len(slice)
	if v == 0 {
		return n + s
	}

	u := b[n : n+s]

	switch maxs {
	case MaxSizeUint16:
		_ = u[1]
		v16 := uint16(v)
		u[0] = byte(v16)
		u[1] = byte(v16 >> 8)
	case MaxSizeUint32:
		_ = u[3]
		v32 := uint32(v)
		u[0] = byte(v32)
		u[1] = byte(v32 >> 8)
		u[2] = byte(v32 >> 16)
		u[3] = byte(v32 >> 24)
	case MaxSizeInt64:
		_ = u[7]
		v64 := uint64(v)
		u[0] = byte(v64)
		u[1] = byte(v64 >> 8)
		u[2] = byte(v64 >> 16)
		u[3] = byte(v64 >> 24)
		u[4] = byte(v64 >> 32)
		u[5] = byte(v64 >> 40)
		u[6] = byte(v64 >> 48)
		u[7] = byte(v64 >> 56)
	}

	n += s

	for _, elem := range slice {
		n = marshal(n, b, elem)
	}
	return n
}

/*
UnmarshalSlice decodes a slice in the benc format.

unmarshal: the corresponding benc unmarshal function for the slice's type! ms: the maximum slice size!

# Panics:

doesn't panic

# Example:

	n = MarshalSlice(n, buf, myIntSlice, bstd.MarshalInt, bstd.MaxSizeUint16)
*/
func UnmarshalSlice[T any](n int, b []byte, unmarshal UnmarshalFunc[T], ms ...MaxSize) (int, []T, error) {
	maxs := MaxSizeUint16
	if len(ms) != 0 {
		maxs = ms[0]
	}

	s := maxSizeToByteSize(maxs)

	if len(b)-n < s {
		return n, nil, ErrBytesToSmall
	}

	u := b[n : n+s]
	var v int = 0

	switch maxs {
	case MaxSizeUint16:
		v = int(uint16(u[0]) | uint16(u[1])<<8)
	case MaxSizeUint32:
		_ = u[3]
		v = int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	case MaxSizeInt64:
		_ = u[7]
		v = int(uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
			uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56)
	}

	n += s
	if len(b)-n < v {
		return n, nil, ErrBytesToSmall
	}

	var t T
	var err error

	ts := make([]T, v)

	for i := 0; i < v; i++ {
		n, t, err = unmarshal(n, b)
		if err != nil {
			return n, nil, errors.New("unmarshal err: " + err.Error())
		}

		ts[i] = t
	}

	return n, ts, nil
}

func SizeMap[K comparable, V any](m map[K]V, kSizer interface{}, vSizer interface{}) int {
	s := 2
	for key, val := range m {
		if p, ok := kSizer.(func(k K) int); ok {
			s += p(key)
		} else if p, ok := kSizer.(func() int); ok {
			s += p()
		}
		if p, ok := vSizer.(func(v V) int); ok {
			s += p(val)
		} else if p, ok := vSizer.(func() int); ok {
			s += p()
		}
	}
	return s
}

func MarshalMap[K comparable, V any](n int, b []byte, m map[K]V, kMarshal MarshalFunc[K], vMarshal MarshalFunc[V]) int {
	size := len(m)
	v := uint16(size)
	u := b[n:]
	_ = u[1]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	n += 2
	if size == 0 {
		return n
	}
	for k, v := range m {
		n = kMarshal(n, b, k)
		n = vMarshal(n, b, v)
	}
	return n
}

func UnmarshalMap[K comparable, V any](n int, b []byte, kUnmarshal UnmarshalFunc[K], vUnmarshal UnmarshalFunc[V]) (int, map[K]V, error) {
	if len(b)-n < 2 {
		return n, nil, ErrBytesToSmall
	}
	u := b[n : n+2]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	if len(b)-n < size {
		return n, nil, ErrBytesToSmall
	}

	result := make(map[K]V, size)
	for i := 0; i < size; i++ {
		var k K
		var v V
		var err error

		n, k, err = kUnmarshal(n, b)
		if err != nil {
			return n, nil, errors.New("unmarshal err (key of map): " + err.Error())
		}
		n, v, err = vUnmarshal(n, b)
		if err != nil {
			return n, nil, errors.New("unmarshal err (val of map): " + err.Error())
		}

		result[k] = v
	}
	return n, result, nil
}

func UnmarshalByte(n int, b []byte) (int, byte, error) {
	if len(b)-n < 1 {
		return n, 0, ErrBytesToSmall
	}
	return n + 1, b[n], nil
}

func UnmarshalString(n int, b []byte) (int, string, error) {
	if len(b)-n < 2 {
		return n, "", ErrBytesToSmall
	}
	u := b[n : n+2]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	bs := b[n : n+size]
	return n + size, string(bs), nil
}

func UnmarshalByteSlice(n int, b []byte) (int, []byte, error) {
	if len(b)-n < 4 {
		return n, nil, ErrBytesToSmall
	}
	u := b[n : n+4]
	_ = u[3]
	size := int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	n += 4
	bs := b[n : n+size]
	return n + size, bs, nil
}

func UnmarshalTime(n int, b []byte) (int, time.Time, error) {
	if len(b)-n < 8 {
		return n, time.Time{}, ErrBytesToSmall
	}
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, time.Unix(0, int64(v)), nil
}

func UnmarshalUInt(n int, b []byte) (int, uint, error) {
	if len(b)-n < 8 {
		return n, 0, ErrBytesToSmall
	}
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, uint(v), nil
}

func UnmarshalUInt64(n int, b []byte) (int, uint64, error) {
	if len(b)-n < 8 {
		return n, 0, ErrBytesToSmall
	}
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, v, nil
}

func UnmarshalUInt32(n int, b []byte) (int, uint32, error) {
	if len(b)-n < 4 {
		return n, 0, ErrBytesToSmall
	}
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, v, nil
}

func UnmarshalUInt16(n int, b []byte) (int, uint16, error) {
	if len(b)-n < 2 {
		return n, 0, ErrBytesToSmall
	}
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, v, nil
}

func UnmarshalInt(n int, b []byte) (int, int, error) {
	if len(b)-n < 8 {
		return n, 0, ErrBytesToSmall
	}
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, int(DecodeZigZag(v)), nil
}

func UnmarshalInt64(n int, b []byte) (int, int64, error) {
	if len(b)-n < 8 {
		return n, 0, ErrBytesToSmall
	}
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, int64(DecodeZigZag(v)), nil
}

func UnmarshalInt32(n int, b []byte) (int, int32, error) {
	if len(b)-n < 4 {
		return n, 0, ErrBytesToSmall
	}
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, int32(DecodeZigZag(v)), nil
}

func UnmarshalInt16(n int, b []byte) (int, int16, error) {
	if len(b)-n < 2 {
		return n, 0, ErrBytesToSmall
	}
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, int16(DecodeZigZag(v)), nil
}

func UnmarshalFloat64(n int, b []byte) (int, float64, error) {
	if len(b)-n < 8 {
		return n, 0, ErrBytesToSmall
	}
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, math.Float64frombits(v), nil
}

func UnmarshalFloat32(n int, b []byte) (int, float32, error) {
	if len(b)-n < 4 {
		return n, 0, ErrBytesToSmall
	}
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, math.Float32frombits(v), nil
}

func UnmarshalBool(n int, b []byte) (int, bool, error) {
	if len(b)-n < 1 {
		return n, false, ErrBytesToSmall
	}
	return n + 1, uint8(b[n]) == 1, nil
}

func SkipSlice(n int, b []byte, skipper SkipFunc) (int, error) {
	if len(b)-n < 2 {
		return n, ErrBytesToSmall
	}
	u := b[n : n+2]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	if len(b)-n < size {
		return n, ErrBytesToSmall
	}
	var err error
	for i := 0; i < size; i++ {
		n, err = skipper(n, b)
		if err != nil {
			return n, errors.New("skipping err: " + err.Error())
		}
	}
	return n, nil
}

func SkipMap(n int, b []byte, kSkipper SkipFunc, vSkipper SkipFunc) (int, error) {
	if len(b)-n < 2 {
		return n, ErrBytesToSmall
	}
	u := b[n : n+2]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	if len(b)-n < size {
		return n, ErrBytesToSmall
	}

	for i := 0; i < size; i++ {
		var err error
		n, err = kSkipper(n, b)
		if err != nil {
			return n, errors.New("skipping err (key of map): " + err.Error())
		}
		n, err = vSkipper(n, b)
		if err != nil {
			return n, errors.New("skipping err (val of map): " + err.Error())
		}
	}
	return n, nil
}

func SkipStringTag(n int, b []byte) (int, error) {
	if n != 0 {
		return 0, ErrNIsNotZero
	}
	if len(b)-n < 2 {
		return n, ErrBytesToSmall
	}
	u := b[n : n+2]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	return n + size, nil
}

func SkipUIntTag(n int, b []byte) (int, error) {
	if n != 0 {
		return 0, ErrNIsNotZero
	}
	if len(b)-n < 2 {
		return n, ErrBytesToSmall
	}
	return n + 2, nil
}

func SkipByte(n int, b []byte) (int, error) {
	if len(b)-n < 1 {
		return n, ErrBytesToSmall
	}
	return n + 1, nil
}

func SkipString(n int, b []byte) (int, error) {
	if len(b)-n < 2 {
		return n, ErrBytesToSmall
	}
	u := b[n : n+2]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	return n + size, nil
}

func SkipByteSlice(n int, b []byte) (int, error) {
	if len(b)-n < 4 {
		return n, ErrBytesToSmall
	}
	u := b[n : n+4]
	_ = u[3]
	size := int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	n += 4
	return n + size, nil
}

func SkipTime(n int, b []byte) (int, error) {
	if len(b)-n < 8 {
		return n, ErrBytesToSmall
	}
	return n + 8, nil
}

func SkipUInt(n int, b []byte) (int, error) {
	if len(b)-n < 8 {
		return n, ErrBytesToSmall
	}
	return n + 8, nil
}

func SkipUInt64(n int, b []byte) (int, error) {
	if len(b)-n < 8 {
		return n, ErrBytesToSmall
	}
	return n + 8, nil
}

func SkipUInt32(n int, b []byte) (int, error) {
	if len(b)-n < 4 {
		return n, ErrBytesToSmall
	}
	return n + 4, nil
}

func SkipUInt16(n int, b []byte) (int, error) {
	if len(b)-n < 2 {
		return n, ErrBytesToSmall
	}
	return n + 2, nil
}

func SkipInt(n int, b []byte) (int, error) {
	if len(b)-n < 8 {
		return n, ErrBytesToSmall
	}
	return n + 8, nil
}

func SkipInt64(n int, b []byte) (int, error) {
	if len(b)-n < 8 {
		return n, ErrBytesToSmall
	}
	return n + 8, nil
}

func SkipInt32(n int, b []byte) (int, error) {
	if len(b)-n < 4 {
		return n, ErrBytesToSmall
	}
	return n + 4, nil
}

func SkipInt16(n int, b []byte) (int, error) {
	if len(b)-n < 2 {
		return n, ErrBytesToSmall
	}
	return n + 2, nil
}

func SkipFloat64(n int, b []byte) (int, error) {
	if len(b)-n < 8 {
		return n, ErrBytesToSmall
	}
	return n + 8, nil
}

func SkipFloat32(n int, b []byte) (int, error) {
	if len(b)-n < 4 {
		return n, ErrBytesToSmall
	}
	return n + 4, nil
}

func SkipBool(n int, b []byte) (int, error) {
	if len(b)-n < 1 {
		return n, ErrBytesToSmall
	}
	return n + 1, nil
}

func SizeString(s string) int {
	return len(s) + 2
}

func MarshalString(n int, b []byte, str string) int {
	v := uint16(len(str))
	u := b[n : n+2]
	_ = u[1]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	return n + 2 + copy(b[n+2:], str)
}

func SizeByteSlice(bs []byte) int {
	return len(bs) + 4
}

func MarshalByteSlice(n int, b []byte, bs []byte) int {
	v := uint32(len(bs))
	u := b[n : n+4]
	_ = u[3]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	u[2] = byte(v >> 16)
	u[3] = byte(v >> 24)
	n += 4
	return n + copy(b[n:], bs)
}

func SizeTime() int {
	return 8
}

func MarshalTime(n int, b []byte, t time.Time) int {
	v := uint64(t.UnixNano())
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

func SizeByte() int {
	return 1
}

func MarshalByte(n int, b []byte, byt byte) int {
	b[n] = byt
	return n + 1
}

func SizeUInt() int {
	return 8
}

func MarshalUInt(n int, b []byte, v uint) int {
	u := b[n : n+8]
	v64 := uint64(v)
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

func SizeInt() int {
	return 8
}

func MarshalInt(n int, b []byte, v int) int {
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

func VerifyMarshal(n int, b []byte) error {
	if n != len(b) {
		return errors.New("check for a mistake in calculating the size or in the marshal process")
	}
	return nil
}

func VerifyUnmarshal(n int, b []byte) error {
	if n != len(b) {
		return errors.New("check for a mistake in the unmarshal process")
	}
	return nil
}

func VerifyUnmarshalMF(n int, b []byte) error {
	if n+2 != len(b) {
		return errors.New("check for a mistake in the unmarshal (message framing) process")
	}
	return nil
}

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
