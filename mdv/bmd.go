package bmd

import (
	"fmt"
	"math"
	"unsafe"

	"go.kine.bz/benc"
	"golang.org/x/exp/constraints"
)

const (
	Int16 byte = iota
	Int32
	Int64
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
)

const AllowedDataTypeStartID = 14

//nolint:funlen
func GetDataTypeName(dataType byte) string {
	switch dataType {
	case Int16:
		return "Int16"
	case Int32:
		return "Int32"
	case Int64:
		return "Int64"
	case UInt16:
		return "Uint16"
	case UInt32:
		return "Uint32"
	case UInt64:
		return "Uint64"
	case Float32:
		return "Float32"
	case Float64:
		return "Float64"
	case Bool:
		return "Bool"
	case Byte:
		return "Byte"
	case String:
		return "String"
	case Slice:
		return "Slice"
	case Map:
		return "Map"
	case ByteSlice:
		return "Byte slice"
	default:
		return "Invalid"
	}
}

type SkipFunc func(n int, b []byte) (int, error)
type UnmarshalFunc[T any] func(n int, b []byte) (int, T, error)

// For unsafe string too
func SkipString(n int, b []byte) (int, error) {
	lb := len(b) - n
	if lb < 2 {
		return n, benc.ErrBufTooSmall
	}

	dt := b[n]
	if dt != String {
		return n, fmt.Errorf("type mismatch: expected String, got %s", GetDataTypeName(dt))
	}
	n++

	s := int(b[n])
	n++

	if s != 2 && s != 4 && s != 8 {
		return n, benc.ErrInvalidSize
	}
	if lb-2 < s {
		return n, benc.ErrBufTooSmall
	}

	u := b[n : n+s]
	v := 0

	switch s {
	case 2:
		v = int(uint16(u[0]) | uint16(u[1])<<8)
	case 4:
		_ = u[3]
		v = int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	case 8:
		_ = u[7]
		v = int(uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
			uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56)
	}

	if lb-s-2 < v {
		return n, benc.ErrInvalidData
	}

	n += s
	return n + v, nil
}

// For unsafe string too
func SizeString(str string, ms ...int) (int, error) {
	s := 2
	v := len(str)
	if len(ms) == 1 {
		s = ms[0]
	}

	switch s {
	case 2:
		if v > math.MaxUint16 {
			return 0, benc.ErrDataTooBig
		}
	case 4:
		if v > math.MaxUint32 {
			return 0, benc.ErrDataTooBig
		}
	case 8:
		break
	default:
		panic("[benc " + benc.BencVersion + "]: invalid `ms` provided in `SizeString`: allowed values, are: 2, 4 and 8")
	}

	return v + s + 2, nil
}

func MarshalString(n int, b []byte, str string, ms ...int) (int, error) {
	b[n] = String
	n++

	s := 2
	if len(ms) == 1 {
		s = ms[0]
	}

	b[n] = byte(s)
	n++

	v := len(str)
	u := b[n : n+s]
	switch s {
	case 2:
		if v > math.MaxUint16 {
			return n - 1, benc.ErrDataTooBig
		}

		_ = u[1]
		u[0] = byte(v)
		u[1] = byte(v >> 8)
	case 4:
		if v > math.MaxUint32 {
			return n - 1, benc.ErrDataTooBig
		}

		_ = u[3]
		v32 := uint32(v)
		u[0] = byte(v32)
		u[1] = byte(v32 >> 8)
		u[2] = byte(v32 >> 16)
		u[3] = byte(v32 >> 24)
	case 8:
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
	default:
		panic("[benc " + benc.BencVersion + "]: invalid `ms` provided in `MarshalString`: allowed values, are: 2, 4 and 8")
	}

	n += s
	return n + copy(b[n:], str), nil
}

func UnmarshalString(n int, b []byte) (int, string, error) {
	lb := len(b) - n
	if lb < 2 {
		return n, "", benc.ErrBufTooSmall
	}

	dt := b[n]
	if dt != String {
		return n, "", fmt.Errorf("type mismatch: expected String, got %s", GetDataTypeName(dt))
	}
	n++

	s := int(b[n])
	n++

	if s != 2 && s != 4 && s != 8 {
		return n, "", benc.ErrInvalidSize
	}
	if lb-2 < s {
		return n, "", benc.ErrBufTooSmall
	}

	u := b[n : n+s]
	v := 0

	switch s {
	case 2:
		v = int(uint16(u[0]) | uint16(u[1])<<8)
	case 4:
		_ = u[3]
		v = int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	case 8:
		_ = u[7]
		v = int(uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
			uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56)
	}

	if lb-s-2 < v {
		return n, "", benc.ErrInvalidData
	}

	n += s
	bs := b[n : n+v]
	return n + v, string(bs), nil
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

func MarshalUnsafeString(n int, b []byte, str string, ms ...int) (int, error) {
	b[n] = String
	n++

	s := 2
	if len(ms) == 1 {
		s = ms[0]
	}

	b[n] = byte(s)
	n++

	v := len(str)
	u := b[n : n+s]

	switch s {
	case 2:
		if v > math.MaxUint16 {
			return n - 1, benc.ErrDataTooBig
		}

		_ = u[1]
		u[0] = byte(v)
		u[1] = byte(v >> 8)
	case 4:
		if v > math.MaxUint32 {
			return n - 1, benc.ErrDataTooBig
		}

		_ = u[3]
		v32 := uint32(v)
		u[0] = byte(v32)
		u[1] = byte(v32 >> 8)
		u[2] = byte(v32 >> 16)
		u[3] = byte(v32 >> 24)
	case 8:
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
	default:
		panic("[benc " + benc.BencVersion + "]: invalid `ms` provided in `MarshalString`: allowed values, are: 2, 4 and 8")
	}

	n += s
	return n + copy(b[n:], s2b(str)), nil
}

func UnmarshalUnsafeString(n int, b []byte) (int, string, error) {
	lb := len(b) - n
	if lb < 2 {
		return n, "", benc.ErrBufTooSmall
	}

	dt := b[n]
	if dt != String {
		return n, "", fmt.Errorf("type mismatch: expected String, got %s", GetDataTypeName(dt))
	}
	n++

	s := int(b[n])
	n++

	if s != 2 && s != 4 && s != 8 {
		return n, "", benc.ErrInvalidSize
	}
	if lb-2 < s {
		return n, "", benc.ErrBufTooSmall
	}

	u := b[n : n+s]
	v := 0

	switch s {
	case 2:
		v = int(uint16(u[0]) | uint16(u[1])<<8)
	case 4:
		_ = u[3]
		v = int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	case 8:
		_ = u[7]
		v = int(uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
			uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56)
	}

	if v == 0 {
		return n + s, "", nil
	}
	if lb-s-2 < v {
		return n, "", benc.ErrInvalidData
	}

	n += s
	bs := b[n : n+v]
	return n + v, b2s(bs), nil
}

//

func SkipSlice(n int, b []byte, skipper SkipFunc) (int, error) {
	lb := len(b) - n
	if lb < 2 {
		return n, benc.ErrBufTooSmall
	}

	dt := b[n]
	if dt != Slice {
		return n, fmt.Errorf("type mismatch: expected Slice, got %s", GetDataTypeName(dt))
	}
	n++

	s := int(b[n])
	n++

	if s != 2 && s != 4 && s != 8 {
		return n, benc.ErrInvalidSize
	}
	if lb-2 < s {
		return n, benc.ErrBufTooSmall
	}

	u := b[n : n+s]
	v := 0

	switch s {
	case 2:
		v = int(uint16(u[0]) | uint16(u[1])<<8)
	case 4:
		_ = u[3]
		v = int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	case 8:
		_ = u[7]
		v = int(uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
			uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56)
	}

	if lb-s-2 < v {
		return n, benc.ErrInvalidData
	}

	n += s
	var err error
	for i := 0; i < v; i++ {
		n, err = skipper(n, b)
		if err != nil {
			return n, fmt.Errorf("at index %d: %s", i, err.Error())
		}
	}
	return n, nil
}

func SizeSlice[T any](slice []T, sizer interface{}, ms ...int) (int, error) {
	s := 2
	v := len(slice)
	if len(ms) == 1 {
		s = ms[0]
	}

	switch s {
	case 2:
		if v > math.MaxUint16 {
			return 0, benc.ErrDataTooBig
		}
	case 4:
		if v > math.MaxUint32 {
			return 0, benc.ErrDataTooBig
		}
	case 8:
		break
	default:
		panic("[benc " + benc.BencVersion + "]: invalid `ms` provided in `SizeSlice`: allowed values, are: 2, 4 and 8")
	}

	var ts int
	var err error

	for i, t := range slice {
		switch p := sizer.(type) {
		case func() int:
			s += p()
		case func(T) (int, error):
			ts, err = p(t)
			if err != nil {
				return 0, fmt.Errorf("at index %d: %s", i, err.Error())
			}
			s += ts
		case func(T, ...int) (int, error):
			ts, err = p(t)
			if err != nil {
				return 0, fmt.Errorf("at index %d: %s", i, err.Error())
			}
			s += ts
		default:
			panic("[benc " + benc.BencVersion + "]: invalid `sizer` provided in `SizeSlice`")
		}
	}

	return s + 2, nil
}

func MarshalSlice[T any](n int, b []byte, slice []T, marshaler interface{}, ms ...int) (int, error) {
	b[n] = Slice
	n++

	s := 2
	if len(ms) == 1 {
		s = ms[0]
	}

	b[n] = byte(s)
	n++

	v := len(slice)
	u := b[n : n+s]

	switch s {
	case 2:
		if v > math.MaxUint16 {
			return n - 1, benc.ErrDataTooBig
		}

		_ = u[1]
		u[0] = byte(v)
		u[1] = byte(v >> 8)
	case 4:
		if v > math.MaxUint32 {
			return n - 1, benc.ErrDataTooBig
		}

		_ = u[3]
		v32 := uint32(v)
		u[0] = byte(v32)
		u[1] = byte(v32 >> 8)
		u[2] = byte(v32 >> 16)
		u[3] = byte(v32 >> 24)
	case 8:
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
	default:
		panic("[benc " + benc.BencVersion + "]: invalid `ms` provided in `MarshalSlice`: allowed values, are: 2, 4 and 8")
	}

	n += s
	var err error
	for i, t := range slice {
		switch p := marshaler.(type) {
		case func(n int, b []byte, t T) int:
			n = p(n, b, t)
		case func(n int, b []byte, t T) (int, error):
			n, err = p(n, b, t)
			if err != nil {
				return n, fmt.Errorf("at index %d: %s", i, err.Error())
			}
		case func(n int, b []byte, t T, ms ...int) (int, error):
			n, err = p(n, b, t)
			if err != nil {
				return n, fmt.Errorf("at index %d: %s", i, err.Error())
			}
		default:
			panic("[benc " + benc.BencVersion + "]: invalid `marshaler` provided in `MarshalSlice`")
		}
	}
	return n, nil
}

func UnmarshalSlice[T any](n int, b []byte, unmarshaler UnmarshalFunc[T]) (int, []T, error) {
	lb := len(b) - n
	if lb < 2 {
		return n, nil, benc.ErrBufTooSmall
	}

	dt := b[n]
	if dt != Slice {
		return n, nil, fmt.Errorf("type mismatch: expected Slice, got %s", GetDataTypeName(dt))
	}
	n++

	s := int(b[n])
	n++

	if s != 2 && s != 4 && s != 8 {
		return n, nil, benc.ErrInvalidSize
	}
	if lb-2 < s {
		return n, nil, benc.ErrBufTooSmall
	}

	u := b[n : n+s]
	v := 0

	switch s {
	case 2:
		v = int(uint16(u[0]) | uint16(u[1])<<8)
	case 4:
		_ = u[3]
		v = int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	case 8:
		_ = u[7]
		v = int(uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
			uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56)
	}

	if lb-s-2 < v {
		return n, nil, benc.ErrInvalidData
	}

	n += s

	var t T
	var err error

	ts := make([]T, v)

	for i := 0; i < v; i++ {
		n, t, err = unmarshaler(n, b)
		if err != nil {
			return n, nil, fmt.Errorf("at index %d: %s", i, err.Error())
		}

		ts[i] = t
	}

	return n, ts, nil
}

// TODO: Do the max size thing wingy

func SkipMap(n int, b []byte, kSkipper SkipFunc, vSkipper SkipFunc) (int, error) {
	lb := len(b) - n
	if lb < 2 {
		return n, benc.ErrBufTooSmall
	}

	dt := b[n]
	if dt != Map {
		return n, fmt.Errorf("type mismatch: expected Map, got %s", GetDataTypeName(dt))
	}
	n++

	s := int(b[n])
	n++

	if s != 2 && s != 4 && s != 8 {
		return n, benc.ErrInvalidSize
	}
	if lb-2 < s {
		return n, benc.ErrBufTooSmall
	}

	u := b[n : n+s]
	v := 0

	switch s {
	case 2:
		v = int(uint16(u[0]) | uint16(u[1])<<8)
	case 4:
		_ = u[3]
		v = int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	case 8:
		_ = u[7]
		v = int(uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
			uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56)
	}

	if lb-s-2 < v {
		return n, benc.ErrInvalidData
	}

	n += s
	var err error
	for i := 0; i < v; i++ {
		n, err = kSkipper(n, b)
		if err != nil {
			return n, fmt.Errorf("(key) at index %d: %s", i, err.Error())
		}

		n, err = vSkipper(n, b)
		if err != nil {
			return n, fmt.Errorf("(value) at index %d: %s", i, err.Error())
		}
	}
	return n, nil
}

func SizeMap[K comparable, V any](m map[K]V, kSizer interface{}, vSizer interface{}, ms ...int) (int, error) {
	s := 2
	v := len(m)
	if len(ms) == 1 {
		s = ms[0]
	}

	switch s {
	case 2:
		if v > math.MaxUint16 {
			return 0, benc.ErrDataTooBig
		}
	case 4:
		if v > math.MaxUint32 {
			return 0, benc.ErrDataTooBig
		}
	case 8:
		break
	default:
		panic("[benc " + benc.BencVersion + "]: invalid `ms` provided in `SizeMap`: allowed values, are: 2, 4 and 8")
	}

	var ts int
	var err error

	var i int
	for k, v := range m {
		switch p := kSizer.(type) {
		case func() int:
			s += p()
		case func(K) (int, error):
			ts, err = p(k)
			if err != nil {
				return 0, fmt.Errorf("(key) at index %d: %s", i, err.Error())
			}
			s += ts
		case func(K, ...int) (int, error):
			ts, err = p(k)
			if err != nil {
				return 0, fmt.Errorf("(key) at index %d: %s", i, err.Error())
			}
			s += ts
		default:
			panic("[benc " + benc.BencVersion + "]: invalid `kSizer` provided in `SizeMap`")
		}

		switch p := vSizer.(type) {
		case func() int:
			s += p()
		case func(V) (int, error):
			ts, err = p(v)
			if err != nil {
				return 0, fmt.Errorf("(value) at index %d: %s", i, err.Error())
			}
			s += ts
		case func(V, ...int) (int, error):
			ts, err = p(v)
			if err != nil {
				return 0, fmt.Errorf("(value) at index %d: %s", i, err.Error())
			}
			s += ts
		default:
			panic("[benc " + benc.BencVersion + "]: invalid `vSizer` provided in `SizeMap`")
		}
		i++
	}

	return s + 2, nil
}

func MarshalMap[K comparable, V any](n int, b []byte, m map[K]V, kMarshaler interface{}, vMarshaler interface{}, ms ...int) (int, error) {
	b[n] = Map
	n++

	s := 2
	if len(ms) == 1 {
		s = ms[0]
	}

	b[n] = byte(s)
	n++

	v := len(m)
	u := b[n : n+s]

	switch s {
	case 2:
		if v > math.MaxUint16 {
			return n - 1, benc.ErrDataTooBig
		}

		_ = u[1]
		u[0] = byte(v)
		u[1] = byte(v >> 8)
	case 4:
		if v > math.MaxUint32 {
			return n - 1, benc.ErrDataTooBig
		}

		_ = u[3]
		v32 := uint32(v)
		u[0] = byte(v32)
		u[1] = byte(v32 >> 8)
		u[2] = byte(v32 >> 16)
		u[3] = byte(v32 >> 24)
	case 8:
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
	default:
		panic("[benc " + benc.BencVersion + "]: invalid `ms` provided in `MarshalMap`: allowed values, are: 2, 4 and 8")
	}

	n += s
	var err error
	var i int
	for k, v := range m {
		switch p := kMarshaler.(type) {
		case func(n int, b []byte, k K) int:
			n = p(n, b, k)
		case func(n int, b []byte, k K) (int, error):
			n, err = p(n, b, k)
			if err != nil {
				return n, fmt.Errorf("(key) at index %d: %s", i, err.Error())
			}
		case func(n int, b []byte, k K, ms ...int) (int, error):
			n, err = p(n, b, k)
			if err != nil {
				return n, fmt.Errorf("(key) at index %d: %s", i, err.Error())
			}
		default:
			panic("[benc " + benc.BencVersion + "]: invalid `kMarshaler` provided in `MarshalMap`")
		}

		switch p := vMarshaler.(type) {
		case func(n int, b []byte, v V) int:
			n = p(n, b, v)
		case func(n int, b []byte, v V) (int, error):
			n, err = p(n, b, v)
			if err != nil {
				return n, fmt.Errorf("(value) at index %d: %s", i, err.Error())
			}
		case func(n int, b []byte, v V, ms ...int) (int, error):
			n, err = p(n, b, v)
			if err != nil {
				return n, fmt.Errorf("(value) at index %d: %s", i, err.Error())
			}
		default:
			panic("[benc " + benc.BencVersion + "]: invalid `vMarshaler` provided in `MarshalMap`")
		}

		i++
	}
	return n, nil
}

func UnmarshalMap[K comparable, V any](n int, b []byte, kUnmarshaler UnmarshalFunc[K], vUnmarshaler UnmarshalFunc[V]) (int, map[K]V, error) {
	lb := len(b) - n
	if lb < 2 {
		return n, nil, benc.ErrBufTooSmall
	}

	dt := b[n]
	if dt != Map {
		return n, nil, fmt.Errorf("type mismatch: expected Map, got %s", GetDataTypeName(dt))
	}
	n++

	s := int(b[n])
	n++

	if s != 2 && s != 4 && s != 8 {
		return n, nil, benc.ErrInvalidSize
	}
	if lb-2 < s {
		return n, nil, benc.ErrBufTooSmall
	}

	u := b[n : n+s]
	v := 0

	switch s {
	case 2:
		v = int(uint16(u[0]) | uint16(u[1])<<8)
	case 4:
		_ = u[3]
		v = int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	case 8:
		_ = u[7]
		v = int(uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
			uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56)
	}

	if lb-s-2 < v {
		return n, nil, benc.ErrInvalidData
	}

	n += s

	var k K
	var val V
	var err error

	ts := make(map[K]V, v)

	for i := 0; i < v; i++ {
		n, k, err = kUnmarshaler(n, b)
		if err != nil {
			return n, nil, fmt.Errorf("(key) at index %d: %s", i, err.Error())
		}

		n, val, err = vUnmarshaler(n, b)
		if err != nil {
			return n, nil, fmt.Errorf("(value) at index %d: %s", i, err.Error())
		}

		ts[k] = val
	}

	return n, ts, nil
}

//

func SkipByte(n int, b []byte) (int, error) {
	if len(b)-n < 2 {
		return n, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Byte {
		return n, fmt.Errorf("type mismatch: expected Byte, got %s", GetDataTypeName(dt))
	}
	return n + 2, nil
}

func SizeByte() int {
	return 2
}

func MarshalByte(n int, b []byte, byt byte) int {
	b[n] = Byte
	b[n+1] = byt
	return n + 2
}

func UnmarshalByte(n int, b []byte) (int, byte, error) {
	if len(b)-n < 2 {
		return n, 0, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Byte {
		return n, 0, fmt.Errorf("type mismatch: expected Byte, got %s", GetDataTypeName(dt))
	}
	n++
	return n + 1, b[n], nil
}

//

func SkipByteSlice(n int, b []byte) (int, error) {
	lb := len(b) - n
	if lb < 2 {
		return n, benc.ErrBufTooSmall
	}

	dt := b[n]
	if dt != ByteSlice {
		return n, fmt.Errorf("type mismatch: expected ByteSlice, got %s", GetDataTypeName(dt))
	}
	n++

	s := int(b[n])
	n++

	if s != 2 && s != 4 && s != 8 {
		return n, benc.ErrInvalidSize
	}
	if lb-2 < s {
		return n, benc.ErrBufTooSmall
	}

	u := b[n : n+s]
	v := 0

	switch s {
	case 2:
		v = int(uint16(u[0]) | uint16(u[1])<<8)
	case 4:
		_ = u[3]
		v = int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	case 8:
		_ = u[7]
		v = int(uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
			uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56)
	}

	if lb-s-2 < v {
		return n, benc.ErrInvalidData
	}

	n += s
	return n + v, nil
}

func SizeByteSlice(bs []byte, ms ...int) (int, error) {
	s := 2
	v := len(bs)
	if len(ms) == 1 {
		s = ms[0]
	}

	switch s {
	case 2:
		if v > math.MaxUint16 {
			return 0, benc.ErrDataTooBig
		}
	case 4:
		if v > math.MaxUint32 {
			return 0, benc.ErrDataTooBig
		}
	case 8:
		break
	default:
		panic("[benc " + benc.BencVersion + "]: invalid `ms` provided in `SizeByteSlice`: allowed values, are: 2, 4 and 8")
	}

	return v + s + 2, nil
}

func MarshalByteSlice(n int, b []byte, bs []byte, ms ...int) (int, error) {
	b[n] = ByteSlice
	n++

	s := 2
	if len(ms) == 1 {
		s = ms[0]
	}

	b[n] = byte(s)
	n++

	v := len(bs)
	u := b[n : n+s]

	switch s {
	case 2:
		if v > math.MaxUint16 {
			return n - 1, benc.ErrDataTooBig
		}

		_ = u[1]
		u[0] = byte(v)
		u[1] = byte(v >> 8)
	case 4:
		if v > math.MaxUint32 {
			return n - 1, benc.ErrDataTooBig
		}

		_ = u[3]
		v32 := uint32(v)
		u[0] = byte(v32)
		u[1] = byte(v32 >> 8)
		u[2] = byte(v32 >> 16)
		u[3] = byte(v32 >> 24)
	case 8:
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
	default:
		panic("[benc " + benc.BencVersion + "]: invalid `ms` provided in `MarshalByteSlice`: allowed values, are: 2, 4 and 8")
	}

	n += s
	return n + copy(b[n:], bs), nil
}

func UnmarshalByteSlice(n int, b []byte) (int, []byte, error) {
	lb := len(b) - n
	if lb < 2 {
		return n, nil, benc.ErrBufTooSmall
	}

	dt := b[n]
	if dt != ByteSlice {
		return n, nil, fmt.Errorf("type mismatch: expected ByteSlice, got %s", GetDataTypeName(dt))
	}
	n++

	s := int(b[n])
	n++

	if s != 2 && s != 4 && s != 8 {
		return n, nil, benc.ErrInvalidSize
	}
	if lb-2 < s {
		return n, nil, benc.ErrBufTooSmall
	}

	u := b[n : n+s]
	v := 0

	switch s {
	case 2:
		v = int(uint16(u[0]) | uint16(u[1])<<8)
	case 4:
		_ = u[3]
		v = int(uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24)
	case 8:
		_ = u[7]
		v = int(uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
			uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56)
	}

	if lb-s-2 < v {
		return n, nil, benc.ErrInvalidData
	}

	n += s
	return n + v, b[n : n+v], nil
}

//

func SkipUInt64(n int, b []byte) (int, error) {
	if len(b)-n < 9 {
		return n, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != UInt64 {
		return n, fmt.Errorf("type mismatch: expected Uint64, got %s", GetDataTypeName(dt))
	}
	return n + 9, nil
}

func SizeUInt64() int {
	return 9
}

func MarshalUInt64(n int, b []byte, v uint64) int {
	b[n] = UInt64
	n++

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
	if len(b)-n < 9 {
		return n, 0, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != UInt64 {
		return n, 0, fmt.Errorf("type mismatch: expected Uint64, got %s", GetDataTypeName(dt))
	}
	n++
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, v, nil
}

//

func SkipUInt32(n int, b []byte) (int, error) {
	if len(b)-n < 5 {
		return n, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != UInt32 {
		return n, fmt.Errorf("type mismatch: expected Uint32, got %s", GetDataTypeName(dt))
	}
	return n + 5, nil
}

func SizeUInt32() int {
	return 5
}

func MarshalUInt32(n int, b []byte, v uint32) int {
	b[n] = UInt32
	n++

	u := b[n : n+4]
	_ = u[3]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	u[2] = byte(v >> 16)
	u[3] = byte(v >> 24)
	return n + 4
}

func UnmarshalUInt32(n int, b []byte) (int, uint32, error) {
	if len(b)-n < 5 {
		return n, 0, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != UInt32 {
		return n, 0, fmt.Errorf("type mismatch: expected Uint32, got %s", GetDataTypeName(dt))
	}
	n++
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, v, nil
}

//

func SkipUInt16(n int, b []byte) (int, error) {
	if len(b)-n < 3 {
		return n, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != UInt16 {
		return n, fmt.Errorf("type mismatch: expected Uint16, got %s", GetDataTypeName(dt))
	}
	return n + 3, nil
}

func SizeUInt16() int {
	return 3
}

func MarshalUInt16(n int, b []byte, v uint16) int {
	b[n] = UInt16
	n++

	u := b[n : n+2]
	_ = u[1]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	return n + 2
}

func UnmarshalUInt16(n int, b []byte) (int, uint16, error) {
	if len(b)-n < 3 {
		return n, 0, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != UInt16 {
		return n, 0, fmt.Errorf("type mismatch: expected Uint16, got %s", GetDataTypeName(dt))
	}
	n++
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, v, nil
}

//

func SkipInt64(n int, b []byte) (int, error) {
	if len(b)-n < 9 {
		return n, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Int64 {
		return n, fmt.Errorf("type mismatch: expected Int64, got %s", GetDataTypeName(dt))
	}
	return n + 9, nil
}

func SizeInt64() int {
	return 9
}

func MarshalInt64(n int, b []byte, v int64) int {
	b[n] = Int64
	n++

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
	if len(b)-n < 9 {
		return n, 0, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Int64 {
		return n, 0, fmt.Errorf("type mismatch: expected Int64, got %s", GetDataTypeName(dt))
	}
	n++
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, int64(DecodeZigZag(v)), nil
}

//

func SkipInt32(n int, b []byte) (int, error) {
	if len(b)-n < 5 {
		return n, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Int32 {
		return n, fmt.Errorf("type mismatch: expected Int32, got %s", GetDataTypeName(dt))
	}
	return n + 5, nil
}

func SizeInt32() int {
	return 5
}

func MarshalInt32(n int, b []byte, v int32) int {
	b[n] = Int32
	n++

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
	if len(b)-n < 5 {
		return n, 0, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Int32 {
		return n, 0, fmt.Errorf("type mismatch: expected Int32, got %s", GetDataTypeName(dt))
	}
	n++
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, int32(DecodeZigZag(v)), nil
}

//

func SkipInt16(n int, b []byte) (int, error) {
	if len(b)-n < 3 {
		return n, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Int16 {
		return n, fmt.Errorf("type mismatch: expected Int16, got %s", GetDataTypeName(dt))
	}
	return n + 3, nil
}

func SizeInt16() int {
	return 3
}

func MarshalInt16(n int, b []byte, v int16) int {
	b[n] = Int16
	n++

	v16 := uint16(EncodeZigZag(v))
	u := b[n : n+2]
	_ = u[1]
	u[0] = byte(v16)
	u[1] = byte(v16 >> 8)
	return n + 2
}

func UnmarshalInt16(n int, b []byte) (int, int16, error) {
	if len(b)-n < 3 {
		return n, 0, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Int16 {
		return n, 0, fmt.Errorf("type mismatch: expected Int16, got %s", GetDataTypeName(dt))
	}
	n++
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, int16(DecodeZigZag(v)), nil
}

//

func SkipFloat64(n int, b []byte) (int, error) {
	if len(b)-n < 9 {
		return n, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Float64 {
		return n, fmt.Errorf("type mismatch: expected Float64, got %s", GetDataTypeName(dt))
	}
	return n + 9, nil
}

func SizeFloat64() int {
	return 9
}

func MarshalFloat64(n int, b []byte, v float64) int {
	b[n] = Float64
	n++

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
	if len(b)-n < 9 {
		return n, 0, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Float64 {
		return n, 0, fmt.Errorf("type mismatch: expected Float64, got %s", GetDataTypeName(dt))
	}
	n++
	u := b[n : n+8]
	_ = u[7]
	v := uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 |
		uint64(u[4])<<32 | uint64(u[5])<<40 | uint64(u[6])<<48 | uint64(u[7])<<56
	return n + 8, math.Float64frombits(v), nil
}

//

func SkipFloat32(n int, b []byte) (int, error) {
	if len(b)-n < 5 {
		return n, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Float32 {
		return n, fmt.Errorf("type mismatch: expected Float32, got %s", GetDataTypeName(dt))
	}
	return n + 5, nil
}

func SizeFloat32() int {
	return 5
}

func MarshalFloat32(n int, b []byte, v float32) int {
	b[n] = Float32
	n++

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
	if len(b)-n < 5 {
		return n, 0, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Float32 {
		return n, 0, fmt.Errorf("type mismatch: expected Float32, got %s", GetDataTypeName(dt))
	}
	n++
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, math.Float32frombits(v), nil
}

//

func SkipBool(n int, b []byte) (int, error) {
	if len(b)-n < 2 {
		return n, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Bool {
		return n, fmt.Errorf("type mismatch: expected Bool, got %s", GetDataTypeName(dt))
	}
	return n + 2, nil
}

func SizeBool() int {
	return 2
}

func MarshalBool(n int, b []byte, v bool) int {
	b[n] = Bool
	n++
	var i byte
	if v {
		i = 1
	}
	b[n] = i
	return n + 1
}

func UnmarshalBool(n int, b []byte) (int, bool, error) {
	if len(b)-n < 2 {
		return n, false, benc.ErrBufTooSmall
	}
	dt := b[n]
	if dt != Bool {
		return n, false, fmt.Errorf("type mismatch: expected Bool, got %s", GetDataTypeName(dt))
	}
	n++
	return n + 1, uint8(b[n]) == 1, nil
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
