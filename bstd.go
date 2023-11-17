package bstd

import (
	"encoding/binary"
	"errors"
	"math"
	"time"

	"golang.org/x/exp/constraints"
)

func MFUnmarshal(b []byte) ([][]byte, error) {
	var dec [][]byte
	var n uint32

	for {
		if 4 > uint32(len(b[n:])) {
			return dec, nil
		}
		s := binary.LittleEndian.Uint32(b[n : n+4])
		n += 4
		if s > uint32(len(b[n:])) {
			return nil, errors.New("expected length not met")
		}
		dec = append(dec, b[n:n+s])
		n += s
	}
}

func MFMarshal(s int) (int, []byte) {
	encoded := make([]byte, s+4)
	binary.LittleEndian.PutUint32(encoded[0:], uint32(s))
	return 4, encoded
}

func MFFinish(n int) int {
	return n + 4
}

func Marshal(s int) (int, []byte) {
	return 0, make([]byte, s)
}

type SizeFunc[T any] func(t T) int
type MarshalFunc[T any] func(n int, b []byte, t T) int
type UnmarshalFunc[T any] func(n int, b []byte) (int, T, error)

func SizeSlice[T any](slice []T, sizer SizeFunc[T]) int {
	s := 2
	for _, t := range slice {
		s += sizer(t)
	}
	return s
}

func MarshalSlice[T any](n int, b []byte, slice []T, marshal MarshalFunc[T]) int {
	binary.LittleEndian.PutUint16(b[n:], uint16(len(slice)))
	n += 2
	for _, elem := range slice {
		n = marshal(n, b, elem)
	}
	return n
}

func UnmarshalSlice[T any](n int, b []byte, unmarshal UnmarshalFunc[T]) (int, []T, error) {
	if len(b)-n < 2 {
		return n, nil, errors.New("insufficient data for decoding length of Slice")
	}
	ui := binary.LittleEndian.Uint16(b[n : n+2])
	n += 2
	if len(b)-n < int(ui) {
		return n, nil, errors.New("insufficient data for decoding Slice")
	}
	ts := make([]T, ui)
	var t T
	var err error
	for i := 0; i < int(ui); i++ {
		n, t, err = unmarshal(n, b)
		if err != nil {
			return n, nil, errors.New("unmarshal slice: " + err.Error())
		}
		ts[i] = t
	}
	return n, ts, nil
}

func SizeMap[K comparable, V any](m map[K]V, sizer SizeFunc[K], vSizer SizeFunc[V]) int {
	s := 2
	for key, val := range m {
		s += sizer(key) + vSizer(val)
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
		return n, nil, errors.New("insufficient data for decoding length of Map")
	}
	size := int(binary.LittleEndian.Uint16(b[n : n+2]))
	n += 2

	if len(b)-n < size {
		return n, nil, errors.New("insufficient data for decoding Map")
	}

	result := make(map[K]V, size)

	for i := 0; i < size; i++ {
		var k K
		var v V
		var err error

		n, k, err = kUnmarshal(n, b)
		if err != nil {
			return n, nil, errors.New("unmarshal key of map: " + err.Error())
		}
		n, v, err = vUnmarshal(n, b)
		if err != nil {
			return n, nil, errors.New("unmarshal val of map: " + err.Error())
		}

		result[k] = v
	}

	return n, result, nil
}

func UnmarshalByte(n int, b []byte) (int, byte, error) {
	if len(b)-n < 1 {
		return n, 0, errors.New("insufficient data for decoding Byte")
	}
	return n + 1, b[n], nil
}

func UnmarshalString(n int, b []byte) (int, string, error) {
	if len(b)-n < 2 {
		return n, "", errors.New("insufficient data for decoding String")
	}
	size := binary.LittleEndian.Uint16(b[n : n+2])
	n += 2
	bs := b[n : uint16(n)+size]
	n += int(size)
	return n, string(bs), nil
}

func UnmarshalByteSlice(n int, b []byte) (int, []byte, error) {
	if len(b)-n < 4 {
		return n, nil, errors.New("insufficient data for decoding ByteSlice")
	}
	size := binary.LittleEndian.Uint32(b[n : n+4])
	n += 4
	bs := b[n : uint32(n)+size]
	n += int(size)
	return n, bs, nil
}

func UnmarshalTime(n int, b []byte) (int, time.Time, error) {
	if len(b)-n < 8 {
		return n, time.Time{}, errors.New("insufficient data for decoding Time")
	}
	ui := binary.LittleEndian.Uint64(b[n : n+8])
	n += 8
	return n, time.Unix(0, int64(ui)), nil
}

func UnmarshalUInt(n int, b []byte) (int, uint, error) {
	if len(b)-n < 8 {
		return n, 0, errors.New("insufficient data for decoding UInt")
	}
	ui := binary.LittleEndian.Uint64(b[n : n+8])
	n += 8
	return n, uint(ui), nil
}

func UnmarshalUInt64(n int, b []byte) (int, uint64, error) {
	if len(b)-n < 8 {
		return n, 0, errors.New("insufficient data for decoding UInt64")
	}
	ui := binary.LittleEndian.Uint64(b[n : n+8])
	n += 8
	return n, ui, nil
}

func UnmarshalUInt32(n int, b []byte) (int, uint32, error) {
	if len(b)-n < 4 {
		return n, 0, errors.New("insufficient data for decoding UInt32")
	}
	ui := binary.LittleEndian.Uint32(b[n : n+4])
	n += 4
	return n, ui, nil
}

func UnmarshalUInt16(n int, b []byte) (int, uint16, error) {
	if len(b)-n < 2 {
		return n, 0, errors.New("insufficient data for decoding UInt16")
	}
	ui := binary.LittleEndian.Uint16(b[n : n+2])
	n += 2
	return n, ui, nil
}

func UnmarshalInt(n int, b []byte) (int, int, error) {
	if len(b)-n < 8 {
		return n, 0, errors.New("insufficient data for decoding Int")
	}
	ui := binary.LittleEndian.Uint64(b[n : n+8])
	n += 8
	return n, int(DecodeZigZag(ui)), nil
}

func UnmarshalInt64(n int, b []byte) (int, int64, error) {
	if len(b)-n < 8 {
		return n, 0, errors.New("insufficient data for decoding Int64")
	}
	ui := binary.LittleEndian.Uint64(b[n : n+8])
	n += 8
	return n, int64(DecodeZigZag(ui)), nil
}

func UnmarshalInt32(n int, b []byte) (int, int32, error) {
	if len(b)-n < 4 {
		return n, 0, errors.New("insufficient data for decoding Int32")
	}
	ui := binary.LittleEndian.Uint32(b[n : n+4])
	n += 4
	return n, int32(DecodeZigZag(ui)), nil
}

func UnmarshalInt16(n int, b []byte) (int, int16, error) {
	if len(b)-n < 2 {
		return n, 0, errors.New("insufficient data for decoding Int16")
	}
	ui := binary.LittleEndian.Uint16(b[n : n+2])
	n += 2
	return n, int16(DecodeZigZag(ui)), nil
}

func UnmarshalFloat64(n int, b []byte) (int, float64, error) {
	if len(b)-n < 8 {
		return n, 0, errors.New("insufficient data for decoding Float64")
	}
	ui := binary.LittleEndian.Uint64(b[n : n+8])
	n += 8
	return n, math.Float64frombits(ui), nil
}

func UnmarshalFloat32(n int, b []byte) (int, float32, error) {
	if len(b)-n < 4 {
		return n, 0, errors.New("insufficient data for decoding Float32")
	}
	ui := binary.LittleEndian.Uint32(b[n : n+4])
	n += 4
	return n, math.Float32frombits(ui), nil
}

func UnmarshalBool(n int, b []byte) (int, bool, error) {
	if len(b)-n < 1 {
		return n, false, errors.New("insufficient data for decoding Bool")
	}
	return n + 1, uint16(b[n]) == 1, nil
}

func SizeString(s string) int {
	return len(s) + 2
}

func MarshalString(n int, b []byte, str string) int {
	s := []byte(str)
	binary.LittleEndian.PutUint16(b[n:], uint16(len(s)))
	n += 2
	copy(b[n:], s)
	n += len(s)
	return n
}

func SizeByteSlice(bs []byte) int {
	return len(bs) + 4
}

func MarshalByteSlice(n int, b []byte, bs []byte) int {
	binary.LittleEndian.PutUint32(b[n:], uint32(len(bs)))
	n += 4
	copy(b[n:], bs)
	n += len(bs)
	return n
}

func SizeTime(_ time.Time) int {
	return 8
}

func MarshalTime(n int, b []byte, t time.Time) int {
	binary.LittleEndian.PutUint64(b[n:], uint64(t.UnixNano()))
	n += 8
	return n
}

func SizeByte(_ byte) int {
	return 1
}

func MarshalByte(n int, b []byte, byt byte) int {
	b[n] = byt
	return n + 1
}

func SizeUInt(_ uint) int {
	return 8
}

func MarshalUInt(n int, b []byte, v uint) int {
	binary.LittleEndian.PutUint64(b[n:], uint64(v))
	n += 8
	return n
}

func SizeUInt64(_ uint64) int {
	return 8
}

func MarshalUInt64(n int, b []byte, v uint64) int {
	binary.LittleEndian.PutUint64(b[n:], v)
	n += 8
	return n
}

func SizeUInt32(_ uint32) int {
	return 4
}

func MarshalUInt32(n int, b []byte, v uint32) int {
	binary.LittleEndian.PutUint32(b[n:], v)
	n += 4
	return n
}

func SizeUInt16(_ uint16) int {
	return 2
}

func MarshalUInt16(n int, b []byte, v uint16) int {
	binary.LittleEndian.PutUint16(b[n:], v)
	n += 2
	return n
}

func SizeInt(_ int) int {
	return 8
}

func MarshalInt(n int, b []byte, v int) int {
	binary.LittleEndian.PutUint64(b[n:], uint64(EncodeZigZag(v)))
	n += 8
	return n
}

func SizeInt64(_ int64) int {
	return 8
}

func MarshalInt64(n int, b []byte, v int64) int {
	binary.LittleEndian.PutUint64(b[n:], uint64(EncodeZigZag(v)))
	n += 8
	return n
}

func SizeInt32(_ int32) int {
	return 4
}

func MarshalInt32(n int, b []byte, v int32) int {
	binary.LittleEndian.PutUint32(b[n:], uint32(EncodeZigZag(v)))
	n += 4
	return n
}

func SizeInt16(_ int16) int {
	return 2
}

func MarshalInt16(n int, b []byte, v int16) int {
	binary.LittleEndian.PutUint16(b[n:], uint16(EncodeZigZag(v)))
	n += 2
	return n
}

func SizeFloat64(_ float64) int {
	return 8
}

func MarshalFloat64(n int, b []byte, v float64) int {
	binary.LittleEndian.PutUint64(b[n:], math.Float64bits(v))
	n += 8
	return n
}

func SizeFloat32(_ float32) int {
	return 4
}

func MarshalFloat32(n int, b []byte, v float32) int {
	binary.LittleEndian.PutUint32(b[n:], math.Float32bits(v))
	n += 4
	return n
}

func SizeBool(_ bool) int {
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
