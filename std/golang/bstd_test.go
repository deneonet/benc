package bstd

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func SkipAll(b []byte, skipers ...func(n int, b []byte) (int, error)) (err error) {
	n := 0

	for i, skiper := range skipers {
		n, err = skiper(n, b)
		if err != nil {
			return fmt.Errorf("(skip) at idx %d: error: %s", i, err.Error())
		}
	}

	if n != len(b) {
		return errors.New("skip failed: something doesn't match in the marshal- and skip progress")
	}
	return nil
}

func SkipOnce_Verify(b []byte, skiper func(n int, b []byte) (int, error)) error {
	n, err := skiper(0, b)

	if err != nil {
		return fmt.Errorf("skip: error: %s", err.Error())
	}

	if n != len(b) {
		return errors.New("skip failed: something doesn't match in the marshal- and skip progress")
	}
	return nil
}

func MarshalAll(s int, values []any, marshals ...func(n int, b []byte, v any) int) ([]byte, error) {
	n := 0
	b := make([]byte, s)

	for i, marshal := range marshals {
		n = marshal(n, b, values[i])
		if n == 0 {
			// error already logged
			return nil, nil
		}
	}

	if n != len(b) {
		return nil, errors.New("marshal failed: something doesn't match in the marshal- and size progress")
	}

	return b, nil
}

func UnmarshalAll(b []byte, values []any, unmarshals ...func(n int, b []byte) (int, any, error)) error {
	n := 0
	var (
		v   any
		err error
	)

	for i, unmarshal := range unmarshals {
		n, v, err = unmarshal(n, b)
		if err != nil {
			return fmt.Errorf("(unmarshal) at idx %d: error: %s", i, err.Error())
		}
		if !reflect.DeepEqual(v, values[i]) {
			return fmt.Errorf("(unmarshal) at idx %d: no match: expected %v, got %v --- (%T - %T)", i, values[i], v, values[i], v)
		}
	}

	if n != len(b) {
		return errors.New("unmarshal failed: something doesn't match in the marshal- and unmarshal progrss")
	}

	return nil
}

func UnmarshalAll_VerifyError(expected error, buffers [][]byte, unmarshals ...func(n int, b []byte) (int, any, error)) error {
	var err error
	for i, unmarshal := range unmarshals {
		_, _, err = unmarshal(0, buffers[i])
		if !errors.Is(err, expected) {
			return fmt.Errorf("(unmarshal) at idx %d: expected a %v error, got %v", i, expected, err)
		}
	}
	return nil
}

func SkipAll_VerifyError(expected error, buffers [][]byte, skipers ...func(n int, b []byte) (int, error)) error {
	var err error
	for i, skiper := range skipers {
		_, err = skiper(0, buffers[i])
		if !errors.Is(err, expected) {
			return fmt.Errorf("(skip) at idx %d: expected a %v error, got %v", i, expected, err)
		}
	}
	return nil
}

func TestDataTypes(t *testing.T) {
	testStr := "Hello World!"
	sizeTestStr := func() int {
		return SizeString(testStr)
	}

	testBs := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sizeTestBs := func() int {
		return SizeBytes(testBs)
	}

	negIntVal := -12345
	posIntVal := int(math.MaxInt)

	values := []any{
		true,
		byte(128),
		rand.Float32(),
		rand.Float64(),
		posIntVal,
		negIntVal,
		int8(-1),
		int16(-1),
		rand.Int31(),
		rand.Int63(),
		uint(math.MaxUint),
		uint16(160),
		rand.Uint32(),
		rand.Uint64(),
		testStr,
		testStr,
		testBs,
		testBs,
	}

	s := SizeBool(false) + SizeByte() + SizeFloat32(1.0) + SizeFloat64(1.0) +
		SizeInt(posIntVal) +
		SizeInt(negIntVal) +
		SizeInt8(0) + SizeInt16(0) + SizeInt32(0) + SizeInt64(0) + SizeUint(math.MaxUint) +
		SizeUint16(0) + SizeUint32(0) + SizeUint64(0) +
		sizeTestStr() + sizeTestStr() + sizeTestBs() + sizeTestBs()

	b, err := MarshalAll(s, values,
		func(n int, b []byte, v any) int { return MarshalBool(n, b, v.(bool)) },
		func(n int, b []byte, v any) int { return MarshalByte(n, b, v.(byte)) },
		func(n int, b []byte, v any) int { return MarshalFloat32(n, b, v.(float32)) },
		func(n int, b []byte, v any) int { return MarshalFloat64(n, b, v.(float64)) },
		func(n int, b []byte, v any) int { return MarshalInt(n, b, v.(int)) },
		func(n int, b []byte, v any) int { return MarshalInt(n, b, v.(int)) },
		func(n int, b []byte, v any) int { return MarshalInt8(n, b, v.(int8)) },
		func(n int, b []byte, v any) int { return MarshalInt16(n, b, v.(int16)) },
		func(n int, b []byte, v any) int { return MarshalInt32(n, b, v.(int32)) },
		func(n int, b []byte, v any) int { return MarshalInt64(n, b, v.(int64)) },
		func(n int, b []byte, v any) int { return MarshalUint(n, b, v.(uint)) },
		func(n int, b []byte, v any) int { return MarshalUint16(n, b, v.(uint16)) },
		func(n int, b []byte, v any) int { return MarshalUint32(n, b, v.(uint32)) },
		func(n int, b []byte, v any) int { return MarshalUint64(n, b, v.(uint64)) },
		func(n int, b []byte, v any) int { return MarshalString(n, b, v.(string)) },
		func(n int, b []byte, v any) int { return MarshalUnsafeString(n, b, v.(string)) },
		func(n int, b []byte, v any) int { return MarshalBytes(n, b, v.([]byte)) },
		func(n int, b []byte, v any) int { return MarshalBytes(n, b, v.([]byte)) },
	)

	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipAll(b, SkipBool, SkipByte, SkipFloat32, SkipFloat64, SkipVarint, SkipVarint, SkipInt8, SkipInt16, SkipInt32, SkipInt64, SkipVarint, SkipUint16, SkipUint32, SkipUint64, SkipString, SkipString, SkipBytes, SkipBytes); err != nil {
		t.Fatal(err.Error())
	}

	if err = UnmarshalAll(b, values,
		func(n int, b []byte) (int, any, error) { return UnmarshalBool(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByte(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalFloat32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalFloat64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt8(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt16(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUint(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUint16(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUint32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUint64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCropped(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCopied(n, b) },
	); err != nil {
		t.Fatal(err.Error())
	}
}

func TestErrBufTooSmall(t *testing.T) {
	buffers := [][]byte{
		{}, {}, {1, 2, 3}, {1, 2, 3, 4, 5, 6, 7}, {}, {}, {1}, {1, 2, 3}, {1, 2, 3, 4, 5, 6, 7}, {},
		{1}, {1, 2, 3}, {1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7},
		{}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7},
		{}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7},
		{}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7},
	}
	if err := UnmarshalAll_VerifyError(ErrBufTooSmall, buffers,
		func(n int, b []byte) (int, any, error) { return UnmarshalBool(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByte(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalFloat32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalFloat64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt8(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt16(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUint(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUint16(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUint32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUint64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCropped(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCropped(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCropped(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCropped(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCopied(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCopied(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCopied(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCopied(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice[byte](n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice[byte](n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice[byte](n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice[byte](n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) {
			return UnmarshalMap[byte, byte](n, b, UnmarshalByte, UnmarshalByte)
		},
		func(n int, b []byte) (int, any, error) {
			return UnmarshalMap[byte, byte](n, b, UnmarshalByte, UnmarshalByte)
		},
		func(n int, b []byte) (int, any, error) {
			return UnmarshalMap[byte, byte](n, b, UnmarshalByte, UnmarshalByte)
		},
		func(n int, b []byte) (int, any, error) {
			return UnmarshalMap[byte, byte](n, b, UnmarshalByte, UnmarshalByte)
		},
	); err != nil {
		t.Fatal(err.Error())
	}

	skipSliceOfBytes := func(n int, b []byte) (int, error) { return SkipSlice(n, b, SkipByte) }
	skipMapOfBytes := func(n int, b []byte) (int, error) { return SkipMap(n, b, SkipByte, SkipByte) }

	if err := SkipAll_VerifyError(ErrBufTooSmall, buffers,
		SkipBool, SkipByte, SkipFloat32, SkipFloat64, SkipVarint, SkipInt8, SkipInt16, SkipInt32, SkipInt64,
		SkipVarint, SkipUint16, SkipUint32, SkipUint64, SkipString, SkipString, SkipString, SkipString,
		SkipString, SkipString, SkipString, SkipString, SkipBytes, SkipBytes, SkipBytes, SkipBytes,
		SkipBytes, SkipBytes, SkipBytes, SkipBytes, skipSliceOfBytes, skipSliceOfBytes, skipSliceOfBytes,
		skipSliceOfBytes, skipMapOfBytes, skipMapOfBytes, skipMapOfBytes, skipMapOfBytes,
	); err != nil {
		t.Fatal(err.Error())
	}
}

func TestErrBufTooSmall_2(t *testing.T) {
	buffers := [][]byte{{}, {2, 0}, {}, {2, 0}, {}, {2, 0}, {}, {10, 0, 0, 0, 1}, {}, {10, 0, 0, 0, 1}, {}, {10, 0, 0, 0, 1}}
	if err := UnmarshalAll_VerifyError(ErrBufTooSmall, buffers,
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCropped(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCropped(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCopied(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCopied(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice[byte](n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice[byte](n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) {
			return UnmarshalMap[byte, byte](n, b, UnmarshalByte, UnmarshalByte)
		},
		func(n int, b []byte) (int, any, error) {
			return UnmarshalMap[byte, byte](n, b, UnmarshalByte, UnmarshalByte)
		},
	); err != nil {
		t.Fatal(err.Error())
	}

	skipSliceOfBytes := func(n int, b []byte) (int, error) { return SkipSlice(n, b, SkipByte) }
	skipMapOfBytes := func(n int, b []byte) (int, error) { return SkipMap(n, b, SkipByte, SkipByte) }
	if err := SkipAll_VerifyError(ErrBufTooSmall, buffers, SkipString, SkipString, SkipString, SkipString, SkipBytes, SkipBytes, SkipBytes, SkipBytes, skipSliceOfBytes, skipSliceOfBytes, skipMapOfBytes, skipMapOfBytes); err != nil {
		t.Fatal(err.Error())
	}
}

func TestSlices(t *testing.T) {
	slice := []string{"sliceelement1", "sliceelement2", "sliceelement3", "sliceelement4", "sliceelement5"}
	s := SizeSlice(slice, SizeString)
	buf := make([]byte, s)
	MarshalSlice(0, buf, slice, MarshalString)

	if err := SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
		return SkipSlice(n, b, SkipString)
	}); err != nil {
		t.Fatal(err.Error())
	}

	_, retSlice, err := UnmarshalSlice[string](0, buf, UnmarshalString)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retSlice, slice) {
		t.Logf("org %v\ndec %v", slice, retSlice)
		t.Fatal("no match!")
	}
}

func TestMaps(t *testing.T) {
	m := make(map[string]string)
	m["mapkey1"] = "mapvalue1"
	m["mapkey2"] = "mapvalue2"
	m["mapkey3"] = "mapvalue3"
	m["mapkey4"] = "mapvalue4"
	m["mapkey5"] = "mapvalue5"

	s := SizeMap(m, SizeString, SizeString)
	buf := make([]byte, s)
	MarshalMap(0, buf, m, MarshalString, MarshalString)

	if err := SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
		return SkipMap(n, b, SkipString, SkipString)
	}); err != nil {
		t.Fatal(err.Error())
	}

	_, retMap, err := UnmarshalMap[string, string](0, buf, UnmarshalString, UnmarshalString)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retMap, m) {
		t.Logf("org %v\ndec %v", m, retMap)
		t.Fatal("no match!")
	}
}

func TestMaps_2(t *testing.T) {
	m := make(map[int32]string)
	m[1] = "mapvalue1"
	m[2] = "mapvalue2"
	m[3] = "mapvalue3"
	m[4] = "mapvalue4"
	m[5] = "mapvalue5"

	s := SizeMap(m, SizeInt32, SizeString)
	buf := make([]byte, s)
	MarshalMap(0, buf, m, MarshalInt32, MarshalString)

	if err := SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
		return SkipMap(n, b, SkipInt32, SkipString)
	}); err != nil {
		t.Fatal(err.Error())
	}

	_, retMap, err := UnmarshalMap[int32, string](0, buf, UnmarshalInt32, UnmarshalString)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retMap, m) {
		t.Logf("org %v\ndec %v", m, retMap)
		t.Fatal("no match!")
	}
}

func TestEmptyString(t *testing.T) {
	str := ""

	s := SizeString(str)
	buf := make([]byte, s)
	MarshalString(0, buf, str)

	if err := SkipOnce_Verify(buf, SkipString); err != nil {
		t.Fatal(err.Error())
	}

	_, retStr, err := UnmarshalString(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retStr, str) {
		t.Logf("org %v\ndec %v", str, retStr)
		t.Fatal("no match!")
	}
}

func TestLongString(t *testing.T) {
	str := ""
	for i := 0; i < math.MaxUint16+1; i++ {
		str += "H"
	}

	s := SizeString(str)
	buf := make([]byte, s)
	MarshalString(0, buf, str)

	if err := SkipOnce_Verify(buf, SkipString); err != nil {
		t.Fatal(err.Error())
	}

	_, retStr, err := UnmarshalString(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retStr, str) {
		t.Logf("org %v\ndec %v", str, retStr)
		t.Fatal("no match!")
	}
}

func TestEmptyUnsafeString(t *testing.T) {
	str := ""

	s := SizeString(str)
	buf := make([]byte, s)
	MarshalUnsafeString(0, buf, str)

	if err := SkipOnce_Verify(buf, SkipString); err != nil {
		t.Fatal(err.Error())
	}

	_, retStr, err := UnmarshalUnsafeString(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retStr, str) {
		t.Logf("org %v\ndec %v", str, retStr)
		t.Fatal("no match!")
	}
}

func getMaxVarintLen() int {
	if strconv.IntSize == 32 {
		return 5
	}
	return 10
}

func TestSkipVarint(t *testing.T) {
	maxLen := getMaxVarintLen()
	overflowEdgeCaseBytes := make([]byte, maxLen)
	for i := 0; i < maxLen-1; i++ {
		overflowEdgeCaseBytes[i] = 0x80
	}
	overflowEdgeCaseBytes[maxLen-1] = 0x02
	if strconv.IntSize == 32 {
		overflowEdgeCaseBytes[maxLen-1] = 0x10
	}

	tests := []struct {
		name    string
		buf     []byte
		n       int
		wantN   int
		wantErr error
	}{
		{"Valid single-byte varint", []byte{0x05}, 0, 1, nil},
		{"Valid multi-byte varint", []byte{0x80, 0x01}, 0, 2, nil},
		{"Buffer too small", []byte{0x80}, 0, 0, ErrBufTooSmall},
		{"Varint overflow", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}, 0, 0, ErrOverflow},
		{"Varint overflow edge case", overflowEdgeCaseBytes, 0, 0, ErrOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotN, gotErr := SkipVarint(tt.n, tt.buf)
			if gotN != tt.wantN || !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("SkipVarint() = (%d, %v), want (%d, %v)", gotN, gotErr, tt.wantN, tt.wantErr)
			}
		})
	}
}

func TestUnmarshalInt(t *testing.T) {
	maxLen := getMaxVarintLen()
	overflowEdgeCaseBytes := make([]byte, maxLen)
	for i := 0; i < maxLen-1; i++ {
		overflowEdgeCaseBytes[i] = 0x80
	}
	overflowEdgeCaseBytes[maxLen-1] = 0x02
	if strconv.IntSize == 32 {
		overflowEdgeCaseBytes[maxLen-1] = 0x10
	}

	tests := []struct {
		name    string
		buf     []byte
		n       int
		wantN   int
		wantVal int
		wantErr error
	}{
		{"Valid small int", []byte{0x02}, 0, 1, 1, nil},
		{"Valid negative int", []byte{0x03}, 0, 1, -2, nil},
		{"Valid multi-byte int", []byte{0xAC, 0x02}, 0, 2, 150, nil},
		{"Buffer too small", []byte{0x80}, 0, 0, 0, ErrBufTooSmall},
		{"Varint overflow", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}, 0, 0, 0, ErrOverflow},
		{"Varint overflow edge case", overflowEdgeCaseBytes, 0, 0, 0, ErrOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotN, gotVal, gotErr := UnmarshalInt(tt.n, tt.buf)
			if gotN != tt.wantN || gotVal != tt.wantVal || !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("UnmarshalInt() = (%d, %d, %v), want (%d, %d, %v)",
					gotN, gotVal, gotErr, tt.wantN, tt.wantVal, tt.wantErr)
			}
		})
	}
}

func TestTime(t *testing.T) {
	now := time.Unix(1663362895, 123456789)

	s := SizeTime()
	buf := make([]byte, s)
	MarshalTime(0, buf, now)

	if err := SkipOnce_Verify(buf, SkipTime); err != nil {
		t.Fatal(err.Error())
	}

	_, retTime, err := UnmarshalTime(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retTime, now) {
		t.Fatalf("no match: \norg %v\ndec %v", now, retTime)
	}
}

func TestPointer(t *testing.T) {
	t.Run("NonNilPointer", func(t *testing.T) {
		val := "hello world"
		ptr := &val

		s := SizePointer(ptr, SizeString)
		buf := make([]byte, s)
		MarshalPointer(0, buf, ptr, MarshalString)

		if err := SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
			return SkipPointer(n, b, SkipString)
		}); err != nil {
			t.Fatal(err.Error())
		}

		_, retPtr, err := UnmarshalPointer[string](0, buf, UnmarshalString)
		if err != nil {
			t.Fatal(err.Error())
		}

		if retPtr == nil {
			t.Fatal("expected a non-nil pointer, got nil")
		}
		if *retPtr != val {
			t.Fatalf("no match: expected %s, got %s", val, *retPtr)
		}
	})

	t.Run("NilPointer", func(t *testing.T) {
		var ptr *string = nil

		s := SizePointer(ptr, SizeString)
		buf := make([]byte, s)
		MarshalPointer(0, buf, ptr, MarshalString)

		if err := SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
			return SkipPointer(n, b, SkipString)
		}); err != nil {
			t.Fatal(err.Error())
		}

		_, retPtr, err := UnmarshalPointer[string](0, buf, UnmarshalString)
		if err != nil {
			t.Fatal(err.Error())
		}

		if retPtr != nil {
			t.Fatal("expected a nil pointer, got a non-nil one")
		}
	})
}

func TestUnmarshalUint(t *testing.T) {
	maxLen := getMaxVarintLen()
	overflowEdgeCaseBytes := make([]byte, maxLen)
	for i := 0; i < maxLen-1; i++ {
		overflowEdgeCaseBytes[i] = 0x80
	}
	overflowEdgeCaseBytes[maxLen-1] = 0x02
	if strconv.IntSize == 32 {
		overflowEdgeCaseBytes[maxLen-1] = 0x10
	}

	tests := []struct {
		name    string
		buf     []byte
		n       int
		wantN   int
		wantVal uint
		wantErr error
	}{
		{"Valid small uint", []byte{0x07}, 0, 1, 7, nil},
		{"Valid multi-byte uint", []byte{0xAC, 0x02}, 0, 2, 300, nil},
		{"Buffer too small", []byte{0x80}, 0, 0, 0, ErrBufTooSmall},
		{"Varint overflow", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}, 0, 0, 0, ErrOverflow},
		{"Varint overflow edge case", overflowEdgeCaseBytes, 0, 0, 0, ErrOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotN, gotVal, gotErr := UnmarshalUint(tt.n, tt.buf)
			if gotN != tt.wantN || gotVal != tt.wantVal || !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("UnmarshalUint() = (%d, %d, %v), want (%d, %d, %v)",
					gotN, gotVal, gotErr, tt.wantN, tt.wantVal, tt.wantErr)
			}
		})
	}
}

func TestTerminatorErrors(t *testing.T) {
	t.Run("SkipSliceTerminator", func(t *testing.T) {
		slice := []string{"a"}
		s := SizeSlice(slice, SizeString)
		buf := make([]byte, s)
		MarshalSlice(0, buf, slice, MarshalString)
		truncatedBuf := buf[:s-1]

		_, err := SkipSlice(0, truncatedBuf, SkipString)
		if !errors.Is(err, ErrBufTooSmall) {
			t.Errorf("Expected %v, got %v", ErrBufTooSmall, err)
		}
	})
	t.Run("SkipMapTerminator", func(t *testing.T) {
		m := map[string]string{"a": "b"}
		s := SizeMap(m, SizeString, SizeString)
		buf := make([]byte, s)
		MarshalMap(0, buf, m, MarshalString, MarshalString)
		truncatedBuf := buf[:s-1]

		_, err := SkipMap(0, truncatedBuf, SkipString, SkipString)
		if !errors.Is(err, ErrBufTooSmall) {
			t.Errorf("Expected %v, got %v", ErrBufTooSmall, err)
		}
	})
}

func TestPanics(t *testing.T) {
	testCases := []struct {
		name string
		fn   func()
	}{
		{"UnmarshalSlice", func() { _, _, _ = UnmarshalSlice[int](0, []byte{1}, "invalid") }},
		{"UnmarshalMapKey", func() { _, _, _ = UnmarshalMap[int, int](0, []byte{1}, "invalid", UnmarshalInt) }},
		{"UnmarshalMapValue", func() { _, _, _ = UnmarshalMap[int, int](0, []byte{1, 0}, UnmarshalInt, "invalid") }},
		{"UnmarshalPointer", func() { _, _, _ = UnmarshalPointer[int](0, []byte{1}, "invalid") }},
		{"SizeMapKey", func() { _ = SizeMap(map[int]int{1: 1}, "invalid", SizeInt) }},
		{"SizeMapValue", func() { _ = SizeMap(map[int]int{1: 1}, SizeInt, "invalid") }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic as expected")
				}
			}()
			tc.fn()
		})
	}
}

// COVERAGE: This final test function covers all remaining uncovered lines.
func TestFinalCoverage(t *testing.T) {
	// Covers SizeFixedSlice
	t.Run("SizeFixedSlice", func(t *testing.T) {
		slice := []int32{1, 2, 3}
		elemSize := SizeInt32(1)
		expected := SizeUint(uint(len(slice))) + len(slice)*elemSize + 4
		if got := SizeFixedSlice(slice, elemSize); got != expected {
			t.Errorf("SizeFixedSlice() = %v, want %v", got, expected)
		}
	})

	// Covers the pointer-based unmarshaler case in UnmarshalSlice
	t.Run("UnmarshalSliceWithPointerArg", func(t *testing.T) {
		slice := []int32{10, 20}
		s := SizeFixedSlice(slice, SizeInt32(10))
		buf := make([]byte, s)
		MarshalSlice(0, buf, slice, MarshalInt32)

		unmarshalInt32AsPointer := func(n int, b []byte, v *int32) (int, error) {
			newN, val, err := UnmarshalInt32(n, b)
			if err == nil {
				*v = val
			}
			return newN, err
		}
		_, retSlice, err := UnmarshalSlice[int32](0, buf, unmarshalInt32AsPointer)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(retSlice, slice) {
			t.Fatal("unmarshalled slice does not match original")
		}

		// Cover error path
		_, _, err = UnmarshalSlice[int32](0, []byte{1, 1, 2, 3}, unmarshalInt32AsPointer)
		if !errors.Is(err, ErrBufTooSmall) {
			t.Errorf("expected ErrBufTooSmall, got %v", err)
		}
	})

	// Covers the pointer-based unmarshaler cases in UnmarshalMap
	t.Run("UnmarshalMapWithPointerArgs", func(t *testing.T) {
		m := map[int16]int32{5: 50}
		s := SizeMap(m, SizeInt16, SizeInt32)
		buf := make([]byte, s)
		MarshalMap(0, buf, m, MarshalInt16, MarshalInt32)

		unmarshalInt16Ptr := func(n int, b []byte, v *int16) (int, error) {
			newN, val, err := UnmarshalInt16(n, b)
			if err == nil {
				*v = val
			}
			return newN, err
		}
		unmarshalInt32Ptr := func(n int, b []byte, v *int32) (int, error) {
			newN, val, err := UnmarshalInt32(n, b)
			if err == nil {
				*v = val
			}
			return newN, err
		}

		_, retMap, err := UnmarshalMap[int16, int32](0, buf, unmarshalInt16Ptr, unmarshalInt32Ptr)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(m, retMap) {
			t.Fatal("unmarshalled map does not match original")
		}

		// Cover error paths
		_, _, err = UnmarshalMap[int16, int32](0, []byte{1}, unmarshalInt16Ptr, unmarshalInt32Ptr) // Key fails
		if !errors.Is(err, ErrBufTooSmall) {
			t.Errorf("expected ErrBufTooSmall for key, got %v", err)
		}
		_, _, err = UnmarshalMap[int16, int32](0, []byte{1, 0, 0}, unmarshalInt16Ptr, unmarshalInt32Ptr) // Value fails
		if !errors.Is(err, ErrBufTooSmall) {
			t.Errorf("expected ErrBufTooSmall for value, got %v", err)
		}
	})

	// Covers error path in UnmarshalTime
	t.Run("UnmarshalTimeError", func(t *testing.T) {
		_, _, err := UnmarshalTime(0, []byte{1, 2, 3})
		if !errors.Is(err, ErrBufTooSmall) {
			t.Errorf("expected ErrBufTooSmall, got %v", err)
		}
	})

	// Covers error path in SkipPointer
	t.Run("SkipPointerError", func(t *testing.T) {
		_, err := SkipPointer(0, []byte{}, SkipByte)
		if !errors.Is(err, ErrBufTooSmall) {
			t.Errorf("expected ErrBufTooSmall, got %v", err)
		}
	})

	// Covers all error paths in UnmarshalPointer
	t.Run("UnmarshalPointerErrors", func(t *testing.T) {
		_, _, err := UnmarshalPointer[int](0, []byte{}, UnmarshalInt)
		if !errors.Is(err, ErrBufTooSmall) {
			t.Errorf("expected ErrBufTooSmall from UnmarshalBool, got %v", err)
		}

		_, _, err = UnmarshalPointer[int](0, []byte{1, 0x80}, UnmarshalInt)
		if !errors.Is(err, ErrBufTooSmall) {
			t.Errorf("expected ErrBufTooSmall from element unmarshaler, got %v", err)
		}

		unmarshalAsPointer := func(n int, b []byte, v *int) (int, error) {
			newN, _, err := UnmarshalInt(n, b)
			return newN, err
		}
		_, _, err = UnmarshalPointer[int](0, []byte{1, 0x80}, unmarshalAsPointer)
		if !errors.Is(err, ErrBufTooSmall) {
			t.Errorf("expected ErrBufTooSmall from pointer-based element unmarshaler, got %v", err)
		}
	})
}
