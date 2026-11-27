package bstd

import (
	"encoding/binary"
	"errors"
	"math"
	"strconv"
	"time"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/exp/constraints"
)

var ErrBufTooSmall = errors.New("buffer too small")
var ErrReuseBufTooSmall = errors.New("reuse buffer too small")
var ErrOverflow = errors.New("varint overflows the native integer width")
var ErrVerifyUnmarshal = errors.New("check for a mistake in the unmarshal process")
var ErrVerifyMarshal = errors.New("check for a mistake in calculating the size or in the marshal process")
var ErrDecodeLimit = errors.New("decode resource limit exceeded")
var ErrInvalidTerminator = errors.New("invalid collection terminator")
var ErrInvalidBool = errors.New("invalid boolean byte")
var ErrInvalidUTF8 = errors.New("invalid UTF-8 string")
var ErrDuplicateKey = errors.New("duplicate map key")

// These limits apply to each decoded field/collection. Recursive callers must
// separately bound nesting depth: the callback API has no shared decode state.
const MaxDecodeBytes = 64 << 20
const MaxDecodeElements = 1 << 20

func hasBytes(b []byte, n, size int) bool {
	return n >= 0 && n <= len(b) && size >= 0 && size <= len(b)-n
}

func payloadLength(b []byte, n int, length uint) (int, error) {
	if length > MaxDecodeBytes || length > uint(^uint(0)>>1) {
		return 0, ErrDecodeLimit
	}
	if !hasBytes(b, n, int(length)) {
		return 0, ErrBufTooSmall
	}
	return int(length), nil
}

func collectionCount(b []byte, n int, count uint, minimumPerElement, elementBytes uintptr) (int, error) {
	if count > MaxDecodeElements || count > uint(^uint(0)>>1) ||
		(elementBytes != 0 && count > uint(MaxDecodeBytes/elementBytes)) {
		return 0, ErrDecodeLimit
	}
	if !hasBytes(b, n, 4) {
		return 0, ErrBufTooSmall
	}
	if count > uint((len(b)-n-4)/int(minimumPerElement)) {
		return 0, ErrBufTooSmall
	}
	return int(count), nil
}

func checkTerminator(n int, b []byte) (int, error) {
	if !hasBytes(b, n, 4) {
		return 0, ErrBufTooSmall
	}
	for _, v := range b[n : n+4] {
		if v != 1 {
			return 0, ErrInvalidTerminator
		}
	}
	return n + 4, nil
}

func validProgress(before, after int, b []byte) bool {
	return after > before && hasBytes(b, after, 0)
}

type SizeFunc[T any] func(t T) int
type MarshalFunc[T any] func(n int, b []byte, t T) int

// Returns the new offset 'n' after skipping the marshalled string.
// For unsafe string unmarshalling too.
//
// Possible errors returned:
//   - ErrOverflow          - varint overflowed a N-bit unsigned integer.
//   - ErrBufTooSmall       - 'buf' was too small to skip the marshalled string.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipString(n int, b []byte) (int, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, err
	}
	s, err := payloadLength(b, n, us)
	if err != nil {
		return 0, err
	}
	if !utf8.Valid(b[n : n+s]) {
		return 0, ErrInvalidUTF8
	}
	return n + s, nil
}

// Returns the bytes needed to marshal a string.
// For unsafe string marshalling too.
func SizeString(str string) int {
	v := len(str)
	return v + SizeUint(uint(v))
}

// Returns the new offset 'n' after marshalling the string.
//
// !- Panics, if 'b' is too small.
func MarshalString(n int, b []byte, str string) int {
	if !utf8.ValidString(str) {
		panic(ErrInvalidUTF8)
	}
	if !hasBytes(b, n, SizeString(str)) {
		panic(ErrBufTooSmall)
	}
	n = MarshalUint(n, b, uint(len(str)))
	if !hasBytes(b, n, len(str)) {
		panic(ErrBufTooSmall)
	}
	return n + copy(b[n:], str)
}

// Returns the new offset 'n', as well as the string, that got unmarshalled.
//
// Possible errors returned:
//   - ErrOverflow          - varint overflowed a N-bit unsigned integer.
//   - ErrBufTooSmall       - 'buf' was too small to unmarshal the slice.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalString(n int, b []byte) (int, string, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, "", err
	}
	s, err := payloadLength(b, n, us)
	if err != nil {
		return 0, "", err
	}
	if !utf8.Valid(b[n : n+s]) {
		return 0, "", ErrInvalidUTF8
	}
	return n + s, string(b[n : n+s]), nil
}

// https://github.com/golang/go/issues/53003#issuecomment-1145241692
//
// see s2b
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

// Returns the new offset 'n' after marshalling the string.
// Uses unsafe operations to convert the string to bytes.
//
// !- Panics, if 'b' is too small.
func MarshalUnsafeString(n int, b []byte, str string) int {
	if !utf8.ValidString(str) {
		panic(ErrInvalidUTF8)
	}
	if !hasBytes(b, n, SizeString(str)) {
		panic(ErrBufTooSmall)
	}
	n = MarshalUint(n, b, uint(len(str)))
	if !hasBytes(b, n, len(str)) {
		panic(ErrBufTooSmall)
	}
	return n + copy(b[n:], s2b(str))
}

// Returns the new offset 'n', as well as the string, that got unmarshalled.
// Uses unsafe operations to convert the bytes to a string.
//
// Possible errors returned:
//   - ErrOverflow          - varint overflowed a N-bit unsigned integer.
//   - ErrBufTooSmall       - 'buf' was too small to unmarshal the slice.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalUnsafeString(n int, b []byte) (int, string, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, "", err
	}
	s, err := payloadLength(b, n, us)
	if err != nil {
		return 0, "", err
	}
	if s == 0 {
		return n, "", nil
	}
	if !utf8.Valid(b[n : n+s]) {
		return 0, "", ErrInvalidUTF8
	}
	return n + s, b2s(b[n : n+s]), nil
}

// Returns the new offset 'n' after skipping the marshalled slice.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'buf' was too small to skip the marshalled slice.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
// It now accepts a skipper function to correctly skip each element and much faster on large sets.
func SkipSlice(n int, b []byte, skipElement func(n int, b []byte) (int, error)) (int, error) {
	if skipElement == nil {
		panic("benc: nil slice skipper")
	}
	// 1. Unmarshal the number of elements in the slice.
	n, elementCount, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, err
	}
	if _, err = collectionCount(b, n, elementCount, 1, 0); err != nil {
		return 0, err
	}

	// 2. Loop 'elementCount' times, calling the provided skipper for each element.
	for i := uint(0); i < elementCount; i++ {
		prev := n
		n, err = skipElement(n, b)
		if err != nil {
			return 0, err
		}
		if !validProgress(prev, n, b) {
			return 0, ErrVerifyUnmarshal
		}
	}

	// 3. Finally, skip the 4-byte terminator.
	return checkTerminator(n, b)
}

// Returns the bytes needed to marshal a slice with a dynamic element size.
func SizeSlice[T any](slice []T, sizer SizeFunc[T]) (s int) {
	v := len(slice)
	s += 4 + SizeUint(uint(v))

	for _, t := range slice {
		s += sizer(t)
	}

	return
}

// Returns the bytes needed to marshal a slice with a fixed element size.
func SizeFixedSlice[T any](slice []T, elemSize int) (s int) {
	v := len(slice)
	s += 4 + SizeUint(uint(v)) + v*elemSize
	return
}

// Returns the new offset 'n' after marshalling the slice.
//
// !- Panics, if 'b' is too small.
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

// Returns the new offset 'n', as well as the slice, that got unmarshalled.
//
// Possible errors returned:
//   - ErrOverflow          - varint overflowed a N-bit unsigned integer.
//   - ErrBufTooSmall       - 'buf' was too small to unmarshal the slice.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalSlice[T any](n int, b []byte, unmarshaler interface{}) (int, []T, error) {
	switch unmarshaler.(type) {
	case func(int, []byte) (int, T, error), func(int, []byte, *T) (int, error):
	default:
		panic("benc: invalid `unmarshaler` provided in `UnmarshalSlice`")
	}
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, nil, err
	}
	var t T
	s, err := collectionCount(b, n, us, 1, unsafe.Sizeof(t))
	if err != nil {
		return 0, nil, err
	}
	ts := make([]T, s)

	switch p := unmarshaler.(type) {
	case func(n int, b []byte) (int, T, error):
		for i := 0; i < s; i++ {
			prev := n
			n, t, err = p(n, b)
			if err != nil {
				return 0, nil, err
			}
			if !validProgress(prev, n, b) {
				return 0, nil, ErrVerifyUnmarshal
			}

			ts[i] = t
		}
	case func(n int, b []byte, v *T) (int, error):
		for i := 0; i < s; i++ {
			prev := n
			n, err = p(n, b, &ts[i])
			if err != nil {
				return 0, nil, err
			}
			if !validProgress(prev, n, b) {
				return 0, nil, ErrVerifyUnmarshal
			}
		}
	default:
		panic("benc: invalid `unmarshaler` provided in `UnmarshalSlice`")
	}

	n, err = checkTerminator(n, b)
	if err != nil {
		return 0, nil, err
	}
	return n, ts, nil
}

// Returns the bytes needed to marshal a fixed size array (passed as a slice).
func SizeArray[T any](s []T, sizer SizeFunc[T]) (sz int) {
	for _, t := range s {
		sz += sizer(t)
	}
	return
}

// Returns the new offset 'n' after marshalling the fixed size array (passed as a slice).
func MarshalArray[T any](n int, b []byte, s []T, marshaler MarshalFunc[T]) int {
	for _, t := range s {
		n = marshaler(n, b, t)
	}
	return n
}

// Unmarshals a fixed size array into the provided slice 's'.
func UnmarshalArray[T any](n int, b []byte, s []T, unmarshaler interface{}) (int, error) {
	if !hasBytes(b, n, 0) {
		return 0, ErrBufTooSmall
	}
	var err error
	switch p := unmarshaler.(type) {
	case func(n int, b []byte) (int, T, error):
		for i := range s {
			prev := n
			n, s[i], err = p(n, b)
			if err != nil {
				return 0, err
			}
			if !validProgress(prev, n, b) {
				return 0, ErrVerifyUnmarshal
			}
		}
	case func(n int, b []byte, v *T) (int, error):
		for i := range s {
			prev := n
			n, err = p(n, b, &s[i])
			if err != nil {
				return 0, err
			}
			if !validProgress(prev, n, b) {
				return 0, ErrVerifyUnmarshal
			}
		}
	default:
		panic("benc: invalid `unmarshaler` provided in `UnmarshalArray`")
	}
	return n, nil
}

// Returns the bytes needed to marshal a fixed size byte array.
func SizeByteArray(n int) int {
	return n
}

// Returns the new offset 'n' after marshalling the fixed size byte array.
func MarshalByteArray(n int, b []byte, src []byte) int {
	if !hasBytes(b, n, len(src)) {
		panic(ErrBufTooSmall)
	}
	return n + copy(b[n:], src)
}

// Unmarshals a fixed size byte array into the provided slice 'dst'.
func UnmarshalByteArray(n int, b []byte, dst []byte) (int, error) {
	if !hasBytes(b, n, len(dst)) {
		return 0, ErrBufTooSmall
	}
	copy(dst, b[n:])
	return n + len(dst), nil
}

// Returns the new offset 'n' after skipping the marshalled map.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'buf' was too small to skip the marshalled map.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
// It now accepts a skipper function to correctly skip each element and much faster on large sets.
func SkipMap(n int, b []byte, skipKey func(n int, b []byte) (int, error), skipValue func(n int, b []byte) (int, error)) (int, error) {
	if skipKey == nil || skipValue == nil {
		panic("benc: nil map skipper")
	}
	// 1. Unmarshal the number of key-value pairs in the map.
	n, pairCount, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, err
	}
	if _, err = collectionCount(b, n, pairCount, 2, 0); err != nil {
		return 0, err
	}

	// 2. Loop 'pairCount' times, skipping one key and one value in each iteration.
	for i := uint(0); i < pairCount; i++ {
		// Skip the key
		prev := n
		n, err = skipKey(n, b)
		if err != nil {
			return 0, err
		}
		if !validProgress(prev, n, b) {
			return 0, ErrVerifyUnmarshal
		}
		// Skip the value
		prev = n
		n, err = skipValue(n, b)
		if err != nil {
			return 0, err
		}
		if !validProgress(prev, n, b) {
			return 0, ErrVerifyUnmarshal
		}
	}

	// 3. Finally, skip the 4-byte terminator.
	return checkTerminator(n, b)
}

// Returns the bytes needed to marshal a map.
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

// Returns the new offset 'n' after marshalling the map.
//
// !- Panics, if 'b' is too small.
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

// Returns the new offset 'n', as well as the map, that got unmarshalled.
//
// Possible errors returned:
//   - ErrOverflow          - varint overflowed a N-bit unsigned integer.
//   - ErrBufTooSmall       - 'buf' was too small to unmarshal the map.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalMap[K comparable, V any](n int, b []byte, kUnmarshaler interface{}, vUnmarshaler interface{}) (int, map[K]V, error) {
	switch kUnmarshaler.(type) {
	case func(int, []byte) (int, K, error), func(int, []byte, *K) (int, error):
	default:
		panic("benc: invalid `kUnmarshaler` provided in `UnmarshalMap`")
	}
	switch vUnmarshaler.(type) {
	case func(int, []byte) (int, V, error), func(int, []byte, *V) (int, error):
	default:
		panic("benc: invalid `vUnmarshaler` provided in `UnmarshalMap`")
	}
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, nil, err
	}
	var k K
	var v V
	s, err := collectionCount(b, n, us, 2, unsafe.Sizeof(k)+unsafe.Sizeof(v)+16)
	if err != nil {
		return 0, nil, err
	}
	ts := make(map[K]V, s)

	for range s {
		beforeKey := n
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
		if !validProgress(beforeKey, n, b) {
			return 0, nil, ErrVerifyUnmarshal
		}

		beforeValue := n
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
			// FIX: Corrected copy-paste error in panic message.
			panic("benc: invalid `vUnmarshaler` provided in `UnmarshalMap`")
		}
		if !validProgress(beforeValue, n, b) {
			return 0, nil, ErrVerifyUnmarshal
		}

		if _, exists := ts[k]; exists {
			return 0, nil, ErrDuplicateKey
		}
		ts[k] = v
	}

	n, err = checkTerminator(n, b)
	if err != nil {
		return 0, nil, err
	}
	return n, ts, nil
}

// Returns the new offset 'n' after skipping the marshalled byte.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'buf' was too small to skip the marshalled byte.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipByte(n int, b []byte) (int, error) {
	if !hasBytes(b, n, 1) {
		return 0, ErrBufTooSmall
	}
	return n + 1, nil
}

// Returns the bytes needed to marshal a byte.
func SizeByte() int {
	return 1
}

// Returns the new offset 'n' after marshalling the byte.
//
// !- Panics, if 'b' is too small.
func MarshalByte(n int, b []byte, byt byte) int {
	b[n] = byt
	return n + 1
}

// Returns the new offset 'n', as well as the byte, that got unmarshalled.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'buf' was too small to unmarshal the byte.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalByte(n int, b []byte) (int, byte, error) {
	if !hasBytes(b, n, 1) {
		return 0, 0, ErrBufTooSmall
	}
	return n + 1, b[n], nil
}

// Returns the new offset 'n' after skipping the marshalled byte slice.
//
// Possible errors returned:
//   - ErrOverflow          - varint overflowed a N-bit unsigned integer.
//   - ErrBufTooSmall       - 'buf' was too small to skip the marshalled byte slice.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipBytes(n int, b []byte) (int, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, err
	}
	s, err := payloadLength(b, n, us)
	if err != nil {
		return 0, err
	}
	return n + s, nil
}

// Returns the bytes needed to marshal a byte slice.
func SizeBytes(bs []byte) int {
	v := len(bs)
	return v + SizeUint(uint(v))
}

// Returns the new offset 'n' after marshalling the byte slice.
//
// !- Panics, if 'b' is too small.
func MarshalBytes(n int, b []byte, bs []byte) int {
	if !hasBytes(b, n, SizeBytes(bs)) {
		panic(ErrBufTooSmall)
	}
	n = MarshalUint(n, b, uint(len(bs)))
	if !hasBytes(b, n, len(bs)) {
		panic(ErrBufTooSmall)
	}
	return n + copy(b[n:], bs)
}

// UnmarshalBytesCopied returns a new allocated slice with a copy of the marshalled byte slice inside `b`.
//
// Possible errors returned:
//   - ErrOverflow          - varint overflowed a N-bit unsigned integer.
//   - ErrBufTooSmall       - 'buf' was too small to unmarshal the unsigned integer.
//
// It's slower than UnmarshalBytesCropped, but modifications to `b` won't affect the returned byte slice.
func UnmarshalBytesCopied(n int, b []byte) (int, []byte, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, nil, err
	}
	s, err := payloadLength(b, n, us)
	if err != nil {
		return 0, nil, err
	}
	cb := make([]byte, s)
	copy(cb, b[n:n+s])
	return n + s, cb, nil
}

// UnmarshalBytesCropped returns a cropped slice of the marshalled byte slice inside `b`.
//
// Possible errors returned:
//   - ErrOverflow          - varint overflowed a N-bit unsigned integer.
//   - ErrBufTooSmall       - 'buf' was too small to unmarshal the unsigned integer.
//
// It's faster than UnmarshalBytesCopied, but modifications to `b` will affect the returned byte slice.
func UnmarshalBytesCropped(n int, b []byte) (int, []byte, error) {
	n, us, err := UnmarshalUint(n, b)
	if err != nil {
		return 0, nil, err
	}
	s, err := payloadLength(b, n, us)
	if err != nil {
		return 0, nil, err
	}
	return n + s, b[n : n+s], nil
}

var maxVarintLenMap = map[int]int{
	64: binary.MaxVarintLen64,
	32: binary.MaxVarintLen32,
}

var maxVarintLen = maxVarintLenMap[strconv.IntSize]
var maxVarintLastByte = byte((1 << (strconv.IntSize - 7*(maxVarintLen-1))) - 1)

// Returns the new offset 'n' after skipping the marshalled varint.
//
// Possible errors returned:
//   - ErrOverflow          - varint overflowed a N-bit unsigned integer.
//   - ErrBufTooSmall       - 'buf' was too small to skip the marshalled varint.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipVarint(n int, buf []byte) (int, error) {
	if !hasBytes(buf, n, 0) {
		return 0, ErrBufTooSmall
	}
	for i, b := range buf[n:] {
		if i == maxVarintLen {
			return 0, ErrOverflow
		}
		if b < 0x80 {
			if i == maxVarintLen-1 && b > maxVarintLastByte {
				return 0, ErrOverflow
			}
			return n + i + 1, nil
		}
	}
	return 0, ErrBufTooSmall
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
//   - ErrOverflow          - varint overflowed a N-bit integer.
//   - ErrBufTooSmall       - 'buf' was too small to unmarshal the integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalInt(n int, buf []byte) (int, int, error) {
	if !hasBytes(buf, n, 0) {
		return 0, 0, ErrBufTooSmall
	}
	var x uint
	var s uint
	for i, b := range buf[n:] {
		if i == maxVarintLen {
			return 0, 0, ErrOverflow
		}
		if b < 0x80 {
			if i == maxVarintLen-1 && b > maxVarintLastByte {
				return 0, 0, ErrOverflow
			}
			return n + i + 1, int(decodeZigZag(x | uint(b)<<s)), nil
		}
		x |= uint(b&0x7f) << s
		s += 7
	}
	return 0, 0, ErrBufTooSmall
}

// Returns the new offset 'n' after skipping the marshalled unsigned integer.
func SkipUint(n int, b []byte) (int, error) {
	return SkipVarint(n, b)
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
//   - ErrOverflow          - varint overflowed a N-bit unsigned integer
//   - ErrBufTooSmall       - 'buf' was too small to unmarshal the unsigned integer//
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalUint(n int, buf []byte) (int, uint, error) {
	if !hasBytes(buf, n, 0) {
		return 0, 0, ErrBufTooSmall
	}
	var x uint
	var s uint
	for i, b := range buf[n:] {
		if i == maxVarintLen {
			return 0, 0, ErrOverflow
		}
		if b < 0x80 {
			if i == maxVarintLen-1 && b > maxVarintLastByte {
				return 0, 0, ErrOverflow
			}
			return n + i + 1, x | uint(b)<<s, nil
		}
		x |= uint(b&0x7f) << s
		s += 7
	}
	return 0, 0, ErrBufTooSmall
}

// Returns the new offset 'n' after skipping the marshalled uintptr.
func SkipUintptr(n int, b []byte) (int, error) {
	return SkipUint(n, b)
}

// Returns the bytes needed to marshal a uintptr.
func SizeUintptr(v uintptr) int {
	return SizeUint(uint(v))
}

// Returns the new offset 'n' after marshalling the uintptr.
func MarshalUintptr(n int, b []byte, v uintptr) int {
	return MarshalUint(n, b, uint(v))
}

// Returns the new offset 'n', as well as the uintptr, that got unmarshalled.
func UnmarshalUintptr(n int, b []byte) (int, uintptr, error) {
	n, v, err := UnmarshalUint(n, b)
	return n, uintptr(v), err
}

// Returns the new offset 'n' after skipping the marshalled 64-bit unsigned integer.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'b' was too small to skip the marshalled 64-bit unsigned integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipUint64(n int, b []byte) (int, error) {
	if !hasBytes(b, n, 8) {
		return 0, ErrBufTooSmall
	}
	return n + 8, nil
}

// Returns the bytes needed to marshal a 64-bit unsigned integer.
func SizeUint64(_ uint64) int {
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
//   - ErrBufTooSmall       - 'b' was too small to unmarshal the 64-bit unsigned integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalUint64(n int, b []byte) (int, uint64, error) {
	if !hasBytes(b, n, 8) {
		return 0, 0, ErrBufTooSmall
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
//   - ErrBufTooSmall       - 'b' was too small to skip the marshalled 32-bit unsigned integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipUint32(n int, b []byte) (int, error) {
	if !hasBytes(b, n, 4) {
		return 0, ErrBufTooSmall
	}
	return n + 4, nil
}

// Returns the bytes needed to marshal a 32-bit unsigned integer.
func SizeUint32(_ uint32) int {
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
//   - ErrBufTooSmall       - 'b' was too small to unmarshal the 32-bit unsigned integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalUint32(n int, b []byte) (int, uint32, error) {
	if !hasBytes(b, n, 4) {
		return 0, 0, ErrBufTooSmall
	}
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, v, nil
}

// Returns the new offset 'n' after skipping the marshalled 16-bit unsigned integer.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'b' was too small to skip the marshalled 16-bit unsigned integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipUint16(n int, b []byte) (int, error) {
	if !hasBytes(b, n, 2) {
		return 0, ErrBufTooSmall
	}
	return n + 2, nil
}

// Returns the bytes needed to marshal a 16-bit unsigned integer.
func SizeUint16(_ uint16) int {
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
//   - ErrBufTooSmall       - 'b' was too small to unmarshal the 16-bit unsigned integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalUint16(n int, b []byte) (int, uint16, error) {
	if !hasBytes(b, n, 2) {
		return 0, 0, ErrBufTooSmall
	}
	u := b[n : n+2]
	_ = u[1]
	v := uint16(u[0]) | uint16(u[1])<<8
	return n + 2, v, nil
}

// Returns the new offset 'n' after skipping the marshalled 64-bit integer.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'b' was too small to skip the marshalled 64-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipInt64(n int, b []byte) (int, error) {
	if !hasBytes(b, n, 8) {
		return 0, ErrBufTooSmall
	}
	return n + 8, nil
}

// Returns the bytes needed to marshal a 64-bit integer.
func SizeInt64(_ int64) int {
	return 8
}

// Returns the new offset 'n' after marshalling the 64-bit integer.
func MarshalInt64(n int, b []byte, v int64) int {
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

// Returns the new offset 'n', as well as the 64-bit integer, that got unmarshalled.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'b' was too small to unmarshal the 64-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalInt64(n int, b []byte) (int, int64, error) {
	if !hasBytes(b, n, 8) {
		return 0, 0, ErrBufTooSmall
	}
	u := b[n : n+8]
	_ = u[7]
	v := int64(u[0]) | int64(u[1])<<8 | int64(u[2])<<16 | int64(u[3])<<24 |
		int64(u[4])<<32 | int64(u[5])<<40 | int64(u[6])<<48 | int64(u[7])<<56
	return n + 8, v, nil
}

// Returns the new offset 'n' after skipping the marshalled 32-bit integer.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'b' was too small to skip the marshalled 32-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipInt32(n int, b []byte) (int, error) {
	if !hasBytes(b, n, 4) {
		return 0, ErrBufTooSmall
	}
	return n + 4, nil
}

// Returns the bytes needed to marshal a 32-bit integer.
func SizeInt32(_ int32) int {
	return 4
}

// Returns the new offset 'n' after marshalling the 32-bit integer.
func MarshalInt32(n int, b []byte, v int32) int {
	u := b[n : n+4]
	_ = u[3]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	u[2] = byte(v >> 16)
	u[3] = byte(v >> 24)
	return n + 4
}

// Returns the new offset 'n', as well as the 32-bit integer, that got unmarshalled.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'b' was too small to unmarshal the 32-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalInt32(n int, b []byte) (int, int32, error) {
	if !hasBytes(b, n, 4) {
		return 0, 0, ErrBufTooSmall
	}
	u := b[n : n+4]
	_ = u[3]
	v := int32(u[0]) | int32(u[1])<<8 | int32(u[2])<<16 | int32(u[3])<<24
	return n + 4, v, nil
}

// Returns the new offset 'n' after skipping the marshalled 16-bit integer.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'b' was too small to skip the marshalled 16-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipInt16(n int, b []byte) (int, error) {
	if !hasBytes(b, n, 2) {
		return 0, ErrBufTooSmall
	}
	return n + 2, nil
}

// Returns the bytes needed to marshal a 16-bit integer.
func SizeInt16(_ int16) int {
	return 2
}

// Returns the new offset 'n' after marshalling the 16-bit integer.
func MarshalInt16(n int, b []byte, v int16) int {
	u := b[n : n+2]
	_ = u[1]
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	return n + 2
}

// Returns the new offset 'n', as well as the 16-bit integer, that got unmarshalled.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'b' was too small to unmarshal the 16-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalInt16(n int, b []byte) (int, int16, error) {
	if !hasBytes(b, n, 2) {
		return 0, 0, ErrBufTooSmall
	}
	u := b[n : n+2]
	_ = u[1]
	v := int16(u[0]) | int16(u[1])<<8
	return n + 2, v, nil
}

// Returns the new offset 'n' after skipping the marshalled 8-bit integer.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'b' was too small to skip the marshalled 16-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipInt8(n int, b []byte) (int, error) {
	return SkipByte(n, b)
}

// Returns the bytes needed to marshal a 8-bit integer.
func SizeInt8(_ int8) int {
	return 1
}

// Returns the new offset 'n' after marshalling the 8-bit integer.
func MarshalInt8(n int, b []byte, v int8) int {
	return MarshalByte(n, b, byte(v))
}

// Returns the new offset 'n', as well as the 8-bit integer, that got unmarshalled.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'b' was too small to unmarshal the 8-bit integer.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalInt8(n int, b []byte) (int, int8, error) {
	n, bi8, err := UnmarshalByte(n, b)
	if err != nil {
		return 0, -1, err
	}

	return n, int8(bi8), err
}

// Returns the new offset 'n' after skipping the marshalled 64-bit float.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'b' was too small to skip the marshalled 64-bit float.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipFloat64(n int, b []byte) (int, error) {
	if !hasBytes(b, n, 8) {
		return 0, ErrBufTooSmall
	}
	return n + 8, nil
}

// Returns the bytes needed to marshal a 64-bit float.
func SizeFloat64(_ float64) int {
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
//   - ErrBufTooSmall       - 'b' was too small to unmarshal the 64-bit float.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalFloat64(n int, b []byte) (int, float64, error) {
	if !hasBytes(b, n, 8) {
		return 0, 0, ErrBufTooSmall
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
//   - ErrBufTooSmall       - 'b' was too small to skip the marshalled 32-bit float.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipFloat32(n int, b []byte) (int, error) {
	if !hasBytes(b, n, 4) {
		return 0, ErrBufTooSmall
	}
	return n + 4, nil
}

// Returns the bytes needed to marshal a 32-bit float.
func SizeFloat32(_ float32) int {
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
//   - ErrBufTooSmall       - 'b' was too small to unmarshal the 32-bit float.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalFloat32(n int, b []byte) (int, float32, error) {
	if !hasBytes(b, n, 4) {
		return 0, 0, ErrBufTooSmall
	}
	u := b[n : n+4]
	_ = u[3]
	v := uint32(u[0]) | uint32(u[1])<<8 | uint32(u[2])<<16 | uint32(u[3])<<24
	return n + 4, math.Float32frombits(v), nil
}

// Returns the new offset 'n' after skipping the marshalled bool.
//
// Possible errors returned:
//   - ErrBufTooSmall       - 'b' was too small to skip the marshalled bool.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func SkipBool(n int, b []byte) (int, error) {
	if !hasBytes(b, n, 1) {
		return 0, ErrBufTooSmall
	}
	if b[n] > 1 {
		return 0, ErrInvalidBool
	}
	return n + 1, nil
}

// Returns the bytes needed to marshal a bool.
func SizeBool(_ bool) int {
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
//   - ErrBufTooSmall       - 'b' was too small to unmarshal the bool.
//
// If a error is returned, n (the int returned) equals zero ( 0 ).
func UnmarshalBool(n int, b []byte) (int, bool, error) {
	if !hasBytes(b, n, 1) {
		return 0, false, ErrBufTooSmall
	}
	if b[n] > 1 {
		return 0, false, ErrInvalidBool
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

// Time functions
func SkipTime(n int, b []byte) (int, error) {
	return SkipInt64(n, b)
}

func SizeTime() int {
	return 8 // int64 for UnixNano
}

func MarshalTime(n int, b []byte, t time.Time) int {
	if t.Before(time.Unix(0, math.MinInt64)) || t.After(time.Unix(0, math.MaxInt64)) {
		panic("benc: time exceeds signed 64-bit Unix nanoseconds")
	}
	return MarshalInt64(n, b, t.UnixNano())
}

func UnmarshalTime(n int, b []byte) (int, time.Time, error) {
	n, nano, err := UnmarshalInt64(n, b)
	if err != nil {
		return n, time.Time{}, err
	}
	return n, time.Unix(0, nano), nil
}

// Pointer fields by adding a boolean prefix

// Skips over a marshalled pointer field. It reads the boolean prefix and,
// if true, calls the provided element skipper to skip the value's data.
func SkipPointer(n int, b []byte, skipElement func(n int, b []byte) (int, error)) (int, error) {
	// 1. Unmarshal the boolean flag that indicates if a value is present.
	n, hasValue, err := UnmarshalBool(n, b)
	if err != nil {
		return 0, err
	}

	// 2. If a value was marshalled (hasValue is true), skip it.
	if hasValue {
		if skipElement == nil {
			panic("benc: nil pointer skipper")
		}
		prev := n
		n, err = skipElement(n, b)
		if err != nil {
			return 0, err
		}
		if !validProgress(prev, n, b) {
			return 0, ErrVerifyUnmarshal
		}
		return n, nil
	}

	// 3. If no value was marshalled, we're done.
	return n, nil
}

func SizePointer[T any](v *T, sizeFn SizeFunc[T]) int {
	if v != nil {
		return SizeBool(false) + sizeFn(*v)
	}
	return SizeBool(false)
}

func MarshalPointer[T any](n int, b []byte, v *T, marshalFn MarshalFunc[T]) int {
	n = MarshalBool(n, b, v != nil)

	if v != nil {
		n = marshalFn(n, b, *v)
	}

	return n
}

func UnmarshalPointer[T any](n int, b []byte, unmarshaler interface{}) (int, *T, error) {
	// 1. Unmarshal the boolean flag.
	n, hasValue, err := UnmarshalBool(n, b)
	if err != nil {
		// On error, return a nil pointer and zero offset, consistent with library style.
		return 0, nil, err
	}

	// 2. If the flag is false, it's a nil pointer. Return immediately.
	if !hasValue {
		return n, nil, nil
	}

	// 3. Only if the flag is true, create and unmarshal the value.
	var t T
	prev := n
	switch p := unmarshaler.(type) {
	case func(n int, b []byte) (int, T, error):
		n, t, err = p(n, b)
		if err != nil {
			return 0, nil, err
		}
	case func(n int, b []byte, v *T) (int, error):
		n, err = p(n, b, &t)
		if err != nil {
			return 0, nil, err
		}
	default:
		// It's good practice to handle unexpected unmarshaler types.
		panic("benc: invalid `unmarshaler` provided in `UnmarshalPointer`")
	}
	if !validProgress(prev, n, b) {
		return 0, nil, ErrVerifyUnmarshal
	}

	return n, &t, nil
}
