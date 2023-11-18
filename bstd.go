package bstd

import (
	"encoding/binary"
	"errors"
	"math"
	"time"

	"github.com/deneonet/benc/bpre"
	"golang.org/x/exp/constraints"
)

var ErrBytesToSmall = errors.New("insufficient data, given bytes are too small")
var ErrNegativeLen = errors.New("un-marshaled length is negative")

func MFUnmarshal(b []byte) ([][]byte, error) {
	var dec [][]byte
	var n uint32

	for {
		if 4 > len(b[n:]) {
			return dec, nil
		}
		s := binary.LittleEndian.Uint32(b[n : n+4])
		n += 4
		if int(s) > len(b[n:]) {
			return nil, errors.New("expected length not met")
		}
		dec = append(dec, b[n:n+s])
		n += s
	}
}

func MFMarshal(s int) (int, []byte) {
	encoded := bpre.GetMarshal(s + 4)
	binary.LittleEndian.PutUint32(encoded[:], uint32(s))
	return 4, encoded
}

func MFFinish(n int) int {
	return n + 4
}

func Marshal(s int) (int, []byte) {
	return 0, bpre.GetMarshal(s)
}

type SizeFunc[T any] func(t T) int
type MarshalFunc[T any] func(n int, b []byte, t T) int
type UnmarshalFunc[T any] func(n int, b []byte) (int, T, error)

func SizeSlice[T any](slice []T, sizer interface{}) int {
	s := 2
	for _, t := range slice {
		switch sizer.(type) {
		case func(t T) int:
			s += sizer.(func(t T) int)(t)
		case func() int:
			s += sizer.(func() int)()
		}
	}
	return s
}

func MarshalSlice[T any](n int, b []byte, slice []T, marshal MarshalFunc[T]) int {
	size := len(slice)
	_ = b[n:][1]
	b[n:][0] = byte(uint16(size))
	b[n:][1] = byte(uint16(size) >> 8)
	n += 2
	if size == 0 {
		return n
	}
	for _, elem := range slice {
		n = marshal(n, b, elem)
	}
	return n
}

func UnmarshalSlice[T any](n int, b []byte, unmarshal UnmarshalFunc[T]) (int, []T, error) {
	if len(b)-n < 2 {
		return n, nil, ErrBytesToSmall
	}
	_ = b[n : n+2][1]
	ui := uint16(b[n : n+2][0]) | uint16(b[n : n+2][1])<<8
	if ui < 0 {
		return n, nil, ErrNegativeLen
	}
	n += 2
	if len(b)-n < int(ui) {
		return n, nil, ErrBytesToSmall
	}
	ts := make([]T, ui)
	var t T
	var err error
	for i := 0; i < int(ui); i++ {
		n, t, err = unmarshal(n, b)
		if err != nil {
			return n, nil, errors.New("unmarshal err: " + err.Error())
		}
		ts[i] = t
	}
	return n, ts, nil
}

func SizeMap[K comparable, V any](m map[K]V, sizer interface{}, vSizer interface{}) int {
	s := 2
	for key, val := range m {
		switch sizer.(type) {
		case func(k K) int:
			s += sizer.(func(k K) int)(key)
		case func() int:
			s += sizer.(func() int)()
		}
		switch vSizer.(type) {
		case func(v V) int:
			s += vSizer.(func(v V) int)(val)
		case func() int:
			s += vSizer.(func() int)()
		}
	}
	return s
}

func MarshalMap[K comparable, V any](n int, b []byte, m map[K]V, kMarshal MarshalFunc[K], vMarshal MarshalFunc[V]) int {
	size := len(m)
	binary.LittleEndian.PutUint16(b[n:], uint16(size))
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
	size := int(binary.LittleEndian.Uint16(b[n : n+2]))
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
	size := binary.LittleEndian.Uint16(b[n : n+2])
	if size < 0 {
		return n, "", ErrNegativeLen
	}
	n += 2
	bs := b[n : n+int(size)]
	return n + int(size), string(bs), nil
}

func UnmarshalByteSlice(n int, b []byte) (int, []byte, error) {
	if len(b)-n < 4 {
		return n, nil, ErrBytesToSmall
	}
	size := binary.LittleEndian.Uint32(b[n : n+4])
	if size < 0 {
		return n, nil, ErrNegativeLen
	}
	n += 4
	bs := b[n : n+int(size)]
	return n + int(size), bs, nil
}

func UnmarshalTime(n int, b []byte) (int, time.Time, error) {
	if len(b)-n < 8 {
		return n, time.Time{}, ErrBytesToSmall
	}
	ui := binary.LittleEndian.Uint64(b[n : n+8])
	if ui < 0 {
		return n + 8, time.Time{}, ErrNegativeLen
	}
	return n + 8, time.Unix(0, int64(ui)), nil
}

func UnmarshalUInt(n int, b []byte) (int, uint, error) {
	if len(b)-n < 8 {
		return n, 0, ErrBytesToSmall
	}
	ui := binary.LittleEndian.Uint64(b[n : n+8])
	if ui < 0 {
		return n + 8, 0, ErrNegativeLen
	}
	return n + 8, uint(ui), nil
}

func UnmarshalUInt64(n int, b []byte) (int, uint64, error) {
	if len(b)-n < 8 {
		return n, 0, ErrBytesToSmall
	}
	ui := binary.LittleEndian.Uint64(b[n : n+8])
	if ui < 0 {
		return n + 8, 0, ErrNegativeLen
	}
	return n + 8, ui, nil
}

func UnmarshalUInt32(n int, b []byte) (int, uint32, error) {
	if len(b)-n < 4 {
		return n, 0, ErrBytesToSmall
	}
	ui := binary.LittleEndian.Uint32(b[n : n+4])
	if ui < 0 {
		return n + 4, 0, ErrNegativeLen
	}
	return n + 4, ui, nil
}

func UnmarshalUInt16(n int, b []byte) (int, uint16, error) {
	if len(b)-n < 2 {
		return n, 0, ErrBytesToSmall
	}
	ui := binary.LittleEndian.Uint16(b[n : n+2])
	if ui < 0 {
		return n + 2, 0, ErrNegativeLen
	}
	return n + 2, ui, nil
}

func UnmarshalInt(n int, b []byte) (int, int, error) {
	if len(b)-n < 8 {
		return n, 0, ErrBytesToSmall
	}
	ui := binary.LittleEndian.Uint64(b[n : n+8])
	if ui < 0 {
		return n + 8, 0, ErrNegativeLen
	}
	return n + 8, int(DecodeZigZag(ui)), nil
}

func UnmarshalInt64(n int, b []byte) (int, int64, error) {
	if len(b)-n < 8 {
		return n, 0, ErrBytesToSmall
	}
	ui := binary.LittleEndian.Uint64(b[n : n+8])
	if ui < 0 {
		return n + 8, 0, ErrNegativeLen
	}
	return n + 8, int64(DecodeZigZag(ui)), nil
}

func UnmarshalInt32(n int, b []byte) (int, int32, error) {
	if len(b)-n < 4 {
		return n, 0, ErrBytesToSmall
	}
	ui := binary.LittleEndian.Uint32(b[n : n+4])
	if ui < 0 {
		return n + 4, 0, ErrNegativeLen
	}
	return n + 4, int32(DecodeZigZag(ui)), nil
}

func UnmarshalInt16(n int, b []byte) (int, int16, error) {
	if len(b)-n < 2 {
		return n, 0, ErrBytesToSmall
	}
	ui := binary.LittleEndian.Uint16(b[n : n+2])
	if ui < 0 {
		return n + 2, 0, ErrNegativeLen
	}
	return n + 2, int16(DecodeZigZag(ui)), nil
}

func UnmarshalFloat64(n int, b []byte) (int, float64, error) {
	if len(b)-n < 8 {
		return n, 0, ErrBytesToSmall
	}
	ui := binary.LittleEndian.Uint64(b[n : n+8])
	if ui < 0 {
		return n + 8, 0, ErrNegativeLen
	}
	return n + 8, math.Float64frombits(ui), nil
}

func UnmarshalFloat32(n int, b []byte) (int, float32, error) {
	if len(b)-n < 4 {
		return n, 0, ErrBytesToSmall
	}
	ui := binary.LittleEndian.Uint32(b[n : n+4])
	if ui < 0 {
		return n + 4, 0, ErrNegativeLen
	}
	return n + 4, math.Float32frombits(ui), nil
}

func UnmarshalBool(n int, b []byte) (int, bool, error) {
	if len(b)-n < 1 {
		return n, false, ErrBytesToSmall
	}
	return n + 1, uint16(b[n]) == 1, nil
}

func SizeString(s string) int {
	return len(s) + 2
}

func MarshalString(n int, b []byte, str string) int {
	binary.LittleEndian.PutUint16(b[n:], uint16(len(str)))
	return n + 2 + copy(b[n+2:], str)
}

func SizeByteSlice(bs []byte) int {
	return len(bs) + 4
}

func MarshalByteSlice(n int, b []byte, bs []byte) int {
	binary.LittleEndian.PutUint32(b[n:], uint32(len(bs)))
	return n + 4 + copy(b[n:], bs)
}

func SizeTime() int {
	return 8
}

func MarshalTime(n int, b []byte, t time.Time) int {
	binary.LittleEndian.PutUint64(b[n:], uint64(t.UnixNano()))
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
	binary.LittleEndian.PutUint64(b[n:], uint64(v))
	return n + 8
}

func SizeUInt64() int {
	return 8
}

func MarshalUInt64(n int, b []byte, v uint64) int {
	binary.LittleEndian.PutUint64(b[n:], v)
	return n + 8
}

func SizeUInt32() int {
	return 4
}

func MarshalUInt32(n int, b []byte, v uint32) int {
	binary.LittleEndian.PutUint32(b[n:], v)
	return n + 4
}

func SizeUInt16() int {
	return 2
}

func MarshalUInt16(n int, b []byte, v uint16) int {
	binary.LittleEndian.PutUint16(b[n:], v)
	return n + 2
}

func SizeInt() int {
	return 8
}

func MarshalInt(n int, b []byte, v int) int {
	binary.LittleEndian.PutUint64(b[n:], uint64(EncodeZigZag(v)))
	return n + 8
}

func SizeInt64() int {
	return 8
}

func MarshalInt64(n int, b []byte, v int64) int {
	binary.LittleEndian.PutUint64(b[n:], uint64(EncodeZigZag(v)))
	return n + 8
}

func SizeInt32() int {
	return 4
}

func MarshalInt32(n int, b []byte, v int32) int {
	binary.LittleEndian.PutUint32(b[n:], uint32(EncodeZigZag(v)))
	return n + 4
}

func SizeInt16() int {
	return 2
}

func MarshalInt16(n int, b []byte, v int16) int {
	binary.LittleEndian.PutUint16(b[n:], uint16(EncodeZigZag(v)))
	return n + 2
}

func SizeFloat64() int {
	return 8
}

func MarshalFloat64(n int, b []byte, v float64) int {
	binary.LittleEndian.PutUint64(b[n:], math.Float64bits(v))
	return n + 8
}

func SizeFloat32() int {
	return 4
}

func MarshalFloat32(n int, b []byte, v float32) int {
	binary.LittleEndian.PutUint32(b[n:], math.Float32bits(v))
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
