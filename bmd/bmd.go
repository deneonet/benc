package bmd

import (
	"errors"
	"golang.org/x/exp/constraints"
	"math"
	"time"
)

var ErrBytesToSmall = errors.New("insufficient data, given bytes are too small")
var ErrNIsNotZero = errors.New("n has to be 0")

const (
	Int byte = iota
	Int16
	Int32
	Int64
	UInt
	UInt16
	UInt32
	UInt64
	Float32
	Float64
	Bool
	Byte
	String
	Slice
	Map
	ByteSlice
	StringTag
	UIntTag
	Time
)

func getDataTypeName(dataType byte) string {
	switch dataType {
	case Int:
		return "int"
	case Int16:
		return "int16"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case UInt:
		return "uint"
	case UInt16:
		return "uint16"
	case UInt32:
		return "uint32"
	case UInt64:
		return "uint64"
	case Float32:
		return "float32"
	case Float64:
		return "float64"
	case Bool:
		return "bool"
	case Byte:
		return "byte"
	case String:
		return "string"
	case Slice:
		return "slice"
	case Map:
		return "map"
	case ByteSlice:
		return "byte slice"
	case StringTag:
		return "string tag"
	case UIntTag:
		return "uint tag"
	case Time:
		return "time"
	default:
		return "unknown"
	}
}

type SizeFunc[T any] func(t T) int
type SkipFunc func(n int, b []byte) (int, error)
type MarshalFunc[T any] func(n int, b []byte, t T) int
type UnmarshalFunc[T any] func(n int, b []byte) (int, T, error)

func MarshalSlice[T any](n int, b []byte, slice []T, marshal MarshalFunc[T]) int {
	b[n] = Slice
	n += 1

	size := len(slice)
	u := b[n:]
	_ = u[1]
	u[0] = byte(uint16(size))
	u[1] = byte(uint16(size) >> 8)
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
	if len(b)-n < 3 {
		return n, nil, ErrBytesToSmall
	}
	if b[n] != Slice {
		return n, nil, errors.New("expected a slice, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+2]
	_ = u[1]
	size := uint16(u[0]) | uint16(u[1])<<8
	n += 2
	if len(b)-n < int(size) {
		return n, nil, ErrBytesToSmall
	}
	ts := make([]T, size)
	var t T
	var err error
	for i := 0; i < int(size); i++ {
		n, t, err = unmarshal(n, b)
		if err != nil {
			return n, nil, errors.New("unmarshal err: " + err.Error())
		}
		ts[i] = t
	}
	return n, ts, nil
}

func SizeMap[K comparable, V any](m map[K]V, kSizer interface{}, vSizer interface{}) int {
	s := 3
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
	b[n] = Map
	n += 1

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
	if len(b)-n < 3 {
		return n, nil, ErrBytesToSmall
	}
	if b[n] != Map {
		return n, nil, errors.New("expected a map, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
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
	if len(b)-n < 2 {
		return n, 0, ErrBytesToSmall
	}
	if b[n] != Byte {
		return n, 0, errors.New("expected a byte, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 1, b[n], nil
}

func UnmarshalString(n int, b []byte) (int, string, error) {
	if len(b)-n < 3 {
		return n, "", ErrBytesToSmall
	}
	if b[n] != String {
		return n, "", errors.New("expected a string, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+2]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	bs := b[n : n+size]
	return n + size, string(bs), nil
}

func UnmarshalByteSlice(n int, b []byte) (int, []byte, error) {
	if len(b)-n < 5 {
		return n, nil, ErrBytesToSmall
	}
	if b[n] != ByteSlice {
		return n, nil, errors.New("expected a byte slice, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+4]
	_ = u[3]
	size := int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	println(n + 4)
	println(n + size)
	println(len(b))
	bs := b[n+4 : n+size]
	return n + 4 + size, bs, nil
}

func UnmarshalTime(n int, b []byte) (int, time.Time, error) {
	if len(b)-n < 9 {
		return n, time.Time{}, ErrBytesToSmall
	}
	if b[n] != Time {
		return n, time.Time{}, errors.New("expected a time, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, time.Unix(0, int64(v)), nil
}

func UnmarshalUInt(n int, b []byte) (int, uint, error) {
	if len(b)-n < 9 {
		return n, 0, ErrBytesToSmall
	}
	if b[n] != UInt {
		return n, 0, errors.New("expected a uint, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, uint(v), nil
}

func UnmarshalUInt64(n int, b []byte) (int, uint64, error) {
	if len(b)-n < 9 {
		return n, 0, ErrBytesToSmall
	}
	if b[n] != UInt64 {
		return n, 0, errors.New("expected a uint64, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, v, nil
}

func UnmarshalUInt32(n int, b []byte) (int, uint32, error) {
	if len(b)-n < 5 {
		return n, 0, ErrBytesToSmall
	}
	if b[n] != UInt32 {
		return n, 0, errors.New("expected a uint32, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, v, nil
}

func UnmarshalUInt16(n int, b []byte) (int, uint16, error) {
	if len(b)-n < 3 {
		return n, 0, ErrBytesToSmall
	}
	if b[n] != UInt16 {
		return n, 0, errors.New("expected a uint16, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, v, nil
}

func UnmarshalInt(n int, b []byte) (int, int, error) {
	if len(b)-n < 9 {
		return n, 0, ErrBytesToSmall
	}
	if b[n] != Int {
		return n, 0, errors.New("expected a int, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, int(DecodeZigZag(v)), nil
}

func UnmarshalInt64(n int, b []byte) (int, int64, error) {
	if len(b)-n < 9 {
		return n, 0, ErrBytesToSmall
	}
	if b[n] != Int64 {
		return n, 0, errors.New("expected a int64, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, int64(DecodeZigZag(v)), nil
}

func UnmarshalInt32(n int, b []byte) (int, int32, error) {
	if len(b)-n < 5 {
		return n, 0, ErrBytesToSmall
	}
	if b[n] != Int32 {
		return n, 0, errors.New("expected a int32, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, int32(DecodeZigZag(v)), nil
}

func UnmarshalInt16(n int, b []byte) (int, int16, error) {
	if len(b)-n < 3 {
		return n, 0, ErrBytesToSmall
	}
	if b[n] != Int16 {
		return n, 0, errors.New("expected a int16, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, int16(DecodeZigZag(v)), nil
}

func UnmarshalFloat64(n int, b []byte) (int, float64, error) {
	if len(b)-n < 9 {
		return n, 0, ErrBytesToSmall
	}
	if b[n] != Float64 {
		return n, 0, errors.New("expected a float64, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, math.Float64frombits(v), nil
}

func UnmarshalFloat32(n int, b []byte) (int, float32, error) {
	if len(b)-n < 5 {
		return n, 0, ErrBytesToSmall
	}
	if b[n] != Float32 {
		return n, 0, errors.New("expected a float32, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, math.Float32frombits(v), nil
}

func UnmarshalBool(n int, b []byte) (int, bool, error) {
	if len(b)-n < 2 {
		return n, false, ErrBytesToSmall
	}
	if b[n] != Bool {
		return n, false, errors.New("expected a bool, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 1, uint8(b[n]) == 1, nil
}

func SkipSlice(n int, b []byte, skipper SkipFunc) (int, error) {
	if len(b)-n < 3 {
		return n, ErrBytesToSmall
	}
	if b[n] != Slice {
		return n, errors.New("expected a slice, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
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
	if len(b)-n < 3 {
		return n, ErrBytesToSmall
	}
	if b[n] != Map {
		return n, errors.New("expected a map, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
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
	if len(b)-n < 3 {
		return n, ErrBytesToSmall
	}

	if b[n] != StringTag {
		return n, errors.New("expected a slice, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1

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
	if len(b)-n < 3 {
		return n, ErrBytesToSmall
	}
	if b[n] != UIntTag {
		return n, errors.New("expected a uint tag, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 2, nil
}

func SkipByte(n int, b []byte) (int, error) {
	if len(b)-n < 2 {
		return n, ErrBytesToSmall
	}
	if b[n] != Byte {
		return n, errors.New("expected a byte, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 1, nil
}

func SkipString(n int, b []byte) (int, error) {
	if len(b)-n < 3 {
		return n, ErrBytesToSmall
	}
	if b[n] != String {
		return n, errors.New("expected a string, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+2]
	_ = u[1]
	size := int(uint16(u[0]) | uint16(u[1])<<8)
	n += 2
	return n + size, nil
}

func SkipByteSlice(n int, b []byte) (int, error) {
	if len(b)-n < 5 {
		return n, ErrBytesToSmall
	}
	if b[n] != ByteSlice {
		return n, errors.New("expected a byte slice, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	u := b[n : n+4]
	_ = u[3]
	size := int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	n += 4
	return n + size, nil
}

func SkipTime(n int, b []byte) (int, error) {
	if len(b)-n < 9 {
		return n, ErrBytesToSmall
	}
	if b[n] != Time {
		return n, errors.New("expected a time, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 8, nil
}

func SkipUInt(n int, b []byte) (int, error) {
	if len(b)-n < 9 {
		return n, ErrBytesToSmall
	}
	if b[n] != UInt {
		return n, errors.New("expected a uint, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 8, nil
}

func SkipUInt64(n int, b []byte) (int, error) {
	if len(b)-n < 9 {
		return n, ErrBytesToSmall
	}
	if b[n] != UInt64 {
		return n, errors.New("expected a uint64, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 8, nil
}

func SkipUInt32(n int, b []byte) (int, error) {
	if len(b)-n < 5 {
		return n, ErrBytesToSmall
	}
	if b[n] != UInt32 {
		return n, errors.New("expected a uint32, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 4, nil
}

func SkipUInt16(n int, b []byte) (int, error) {
	if len(b)-n < 3 {
		return n, ErrBytesToSmall
	}
	if b[n] != UInt16 {
		return n, errors.New("expected a uint16, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 2, nil
}

func SkipInt(n int, b []byte) (int, error) {
	if len(b)-n < 9 {
		return n, ErrBytesToSmall
	}
	if b[n] != Int {
		return n, errors.New("expected a int, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 8, nil
}

func SkipInt64(n int, b []byte) (int, error) {
	if len(b)-n < 9 {
		return n, ErrBytesToSmall
	}
	if b[n] != Int64 {
		return n, errors.New("expected a int64, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 8, nil
}

func SkipInt32(n int, b []byte) (int, error) {
	if len(b)-n < 5 {
		return n, ErrBytesToSmall
	}
	if b[n] != Int32 {
		return n, errors.New("expected a int32, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 4, nil
}

func SkipInt16(n int, b []byte) (int, error) {
	if len(b)-n < 3 {
		return n, ErrBytesToSmall
	}
	if b[n] != Int16 {
		return n, errors.New("expected a int16, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 2, nil
}

func SkipFloat64(n int, b []byte) (int, error) {
	if len(b)-n < 9 {
		return n, ErrBytesToSmall
	}
	if b[n] != Float64 {
		return n, errors.New("expected a float64, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 8, nil
}

func SkipFloat32(n int, b []byte) (int, error) {
	if len(b)-n < 5 {
		return n, ErrBytesToSmall
	}
	if b[n] != Float32 {
		return n, errors.New("expected a float32, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 4, nil
}

func SkipBool(n int, b []byte) (int, error) {
	if len(b)-n < 2 {
		return n, ErrBytesToSmall
	}
	if b[n] != Bool {
		return n, errors.New("expected a bool, found: " + getDataTypeName(b[n]) + ". check your marshal process")
	}
	n += 1
	return n + 1, nil
}

func SizeString(s string) int {
	return len(s) + 3
}

func MarshalString(n int, b []byte, str string) int {
	b[n] = String
	n += 1

	v := uint16(len(str))
	u := b[n : n+2]
	_ = u[1]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	return n + 2 + copy(b[n+2:], str)
}

func SizeByteSlice(bs []byte) int {
	return len(bs) + 5
}

func MarshalByteSlice(n int, b []byte, bs []byte) int {
	b[n] = ByteSlice
	n += 1

	v := uint32(len(bs))
	u := b[n : n+4]
	_ = u[3]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	u[2] = byte(v >> 16)
	u[3] = byte(v >> 24)
	return n + 4 + copy(b[n+4:], bs)
}

func SizeTime() int {
	return 9
}

func MarshalTime(n int, b []byte, t time.Time) int {
	b[n] = Time
	n += 1

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
	return 2
}

func MarshalByte(n int, b []byte, byt byte) int {
	b[n] = Byte
	n += 1
	b[n] = byt
	return n + 1
}

func SizeUInt() int {
	return 9
}

func MarshalUInt(n int, b []byte, v uint) int {
	b[n] = UInt
	n += 1
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
	return 9
}

func MarshalUInt64(n int, b []byte, v uint64) int {
	b[n] = UInt64
	n += 1
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
	return 5
}

func MarshalUInt32(n int, b []byte, v uint32) int {
	b[n] = UInt32
	n += 1
	u := b[n : n+4]
	_ = u[3]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	u[2] = byte(v >> 16)
	u[3] = byte(v >> 24)
	return n + 4
}

func SizeUInt16() int {
	return 3
}

func MarshalUInt16(n int, b []byte, v uint16) int {
	b[n] = UInt16
	n += 1
	u := b[n : n+2]
	_ = u[1]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	return n + 2
}

func SizeInt() int {
	return 9
}

func MarshalInt(n int, b []byte, v int) int {
	b[n] = Int
	n += 1
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
	return 9
}

func MarshalInt64(n int, b []byte, v int64) int {
	b[n] = Int64
	n += 1
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
	return 5
}

func MarshalInt32(n int, b []byte, v int32) int {
	b[n] = Int32
	n += 1
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
	return 3
}

func MarshalInt16(n int, b []byte, v int16) int {
	b[n] = Int16
	n += 1
	v16 := uint16(EncodeZigZag(v))
	u := b[n : n+2]
	_ = u[1]
	u[0] = byte(v16)
	u[1] = byte(v16 >> 8)
	return n + 2
}

func SizeFloat64() int {
	return 9
}

func MarshalFloat64(n int, b []byte, v float64) int {
	b[n] = Float64
	n += 1
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
	return 5
}

func MarshalFloat32(n int, b []byte, v float32) int {
	b[n] = Float32
	n += 1
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
	return 2
}

func MarshalBool(n int, b []byte, v bool) int {
	b[n] = Bool
	n += 1
	var i byte
	if v {
		i = 1
	}
	b[n] = i
	return n + 1
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
