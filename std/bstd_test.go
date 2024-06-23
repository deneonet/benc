package bstd

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"go.kine.bz/benc"
)

func SizeAll(sizers ...func() int) int {
	s := 0
	for _, sizer := range sizers {
		ts := sizer()
		if ts == 0 {
			return 0
		}
		s += ts
	}
	return s
}

func SkipAll(b []byte, skipers ...func(n int, b []byte) (int, error)) error {
	n := 0
	var err error
	for i, skiper := range skipers {
		n, err = skiper(n, b)
		if err != nil {
			return fmt.Errorf("(skip) at idx %d: error: %s", i, err.Error())
		}
	}
	if n != len(b) {
		return errors.New("skip failed: something doesn't match in the marshal- and skip progrss")
	}
	return nil
}

func SkipOnce_Verify(b []byte, skiper func(n int, b []byte) (int, error)) error {
	n := 0
	var err error
	n, err = skiper(n, b)
	if err != nil {
		return fmt.Errorf("skip: error: %s", err.Error())
	}
	if n != len(b) {
		return errors.New("skip failed: something doesn't match in the marshal- and skip progrss")
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
		return nil, errors.New("marshal failed: something doesn't match in the marshal- and size progrss")
	}
	return b, nil
}

func UnmarshalAll_Verify(b []byte, values []any, unmarshals ...func(n int, b []byte) (int, any, error)) error {
	n := 0
	var v any
	var err error
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
		if err != expected {
			return fmt.Errorf("(unmarshal) at idx %d: expected a %s error", i, expected)
		}
	}
	return nil
}

func SkipAll_VerifyError(expected error, buffers [][]byte, skipers ...func(n int, b []byte) (int, error)) error {
	var err error
	for i, skiper := range skipers {
		_, err = skiper(0, buffers[i])
		if err != expected {
			return fmt.Errorf("(skip) at idx %d: expected a %s error", i, expected)
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
	s := SizeAll(SizeBool, SizeBool, SizeByte, SizeFloat32, SizeFloat64, SizeInt16, SizeInt32, SizeInt64, SizeUInt16, SizeUInt32, SizeUInt64,
		sizeTestStr, sizeTestStr, func() int {
			return SizeBytes(testBs)
		})

	values := []any{true, false, byte(128), rand.Float32(), rand.Float64(), int16(16), rand.Int31(), rand.Int63(), uint16(160), rand.Uint32(), rand.Uint64(), testStr, testStr, testBs}
	buf, err := MarshalAll(s, values,
		func(n int, b []byte, v any) int { return MarshalBool(n, b, v.(bool)) },
		func(n int, b []byte, v any) int { return MarshalBool(n, b, v.(bool)) },
		func(n int, b []byte, v any) int { return MarshalByte(n, b, v.(byte)) },
		func(n int, b []byte, v any) int { return MarshalFloat32(n, b, v.(float32)) },
		func(n int, b []byte, v any) int { return MarshalFloat64(n, b, v.(float64)) },
		func(n int, b []byte, v any) int { return MarshalInt16(n, b, v.(int16)) },
		func(n int, b []byte, v any) int { return MarshalInt32(n, b, v.(int32)) },
		func(n int, b []byte, v any) int { return MarshalInt64(n, b, v.(int64)) },
		func(n int, b []byte, v any) int { return MarshalUInt16(n, b, v.(uint16)) },
		func(n int, b []byte, v any) int { return MarshalUInt32(n, b, v.(uint32)) },
		func(n int, b []byte, v any) int { return MarshalUInt64(n, b, v.(uint64)) },
		func(n int, b []byte, v any) int { return MarshalString(n, b, v.(string)) },
		func(n int, b []byte, v any) int { return MarshalUnsafeString(n, b, v.(string)) },
		func(n int, b []byte, v any) int { return MarshalBytes(n, b, v.([]byte)) },
	)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipAll(buf, SkipBool, SkipBool, SkipByte, SkipFloat32, SkipFloat64, SkipInt16, SkipInt32, SkipInt64, SkipUInt16, SkipUInt32, SkipUInt64, SkipString, SkipString, SkipBytes); err != nil {
		t.Fatal(err.Error())
	}

	if err = UnmarshalAll_Verify(buf, values,
		func(n int, b []byte) (int, any, error) { return UnmarshalBool(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBool(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByte(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalFloat32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalFloat64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt16(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUInt16(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUInt32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUInt64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytes(n, b) },
	); err != nil {
		t.Fatal(err.Error())
	}
}

func TestErrBufTooSmall(t *testing.T) {
	buffers := [][]byte{{}, {}, {1, 2, 3}, {1, 2, 3, 4, 5, 6, 7}, {1}, {1, 2, 3}, {1, 2, 3, 4, 5, 6, 7}, {1}, {1, 2, 3}, {1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7}}
	if err := UnmarshalAll_VerifyError(benc.ErrBufTooSmall, buffers,
		func(n int, b []byte) (int, any, error) { return UnmarshalBool(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByte(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalFloat32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalFloat64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt16(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUInt16(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUInt32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUInt64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytes(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytes(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytes(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytes(n, b) },
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

	skipSliceOfBytes := func(n int, b []byte) (int, error) { return SkipSlice(n, b) }
	skipMapOfBytes := func(n int, b []byte) (int, error) { return SkipMap(n, b, SkipByte, SkipByte) }
	if err := SkipAll_VerifyError(benc.ErrBufTooSmall, buffers, SkipBool, SkipByte, SkipFloat32, SkipFloat64, SkipInt16, SkipInt32, SkipInt64, SkipUInt16, SkipUInt32, SkipUInt64, SkipString, SkipString, SkipString, SkipString, SkipString, SkipString, SkipString, SkipString, SkipBytes, SkipBytes, SkipBytes, SkipBytes, skipSliceOfBytes, skipSliceOfBytes, skipSliceOfBytes, skipSliceOfBytes, skipMapOfBytes, skipMapOfBytes, skipMapOfBytes, skipMapOfBytes); err != nil {
		t.Fatal(err.Error())
	}
}

func TestErrBufTooSmall_2(t *testing.T) {
	buffers := [][]byte{{}, {2, 0}, {}, {2, 0}, {}, {2, 0}, {0, 0, 0}, {10, 0, 0, 0, 1}, {10, 0, 0, 0, 1, 2, 3}}
	if err := UnmarshalAll_VerifyError(benc.ErrBufTooSmall, buffers,
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytes(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytes(n, b) },
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
	); err != nil {
		t.Fatal(err.Error())
	}

	skipSliceOfBytes := func(n int, b []byte) (int, error) { return SkipSlice(n, b) }
	skipMapOfBytes := func(n int, b []byte) (int, error) { return SkipMap(n, b, SkipByte, SkipByte) }
	if err := SkipAll_VerifyError(benc.ErrBufTooSmall, buffers, SkipString, SkipString, SkipString, SkipString, SkipString, SkipString, SkipBytes, SkipBytes, SkipBytes, skipSliceOfBytes, skipSliceOfBytes, skipSliceOfBytes, skipMapOfBytes, skipMapOfBytes, skipMapOfBytes); err != nil {
		t.Fatal(err.Error())
	}
}

func TestSlices(t *testing.T) {
	slice := []string{"sliceelement1", "sliceelement2", "sliceelement3", "sliceelement4", "sliceelement5"}
	s := SizeSlice(slice, SizeString)
	buf := make([]byte, s)
	MarshalSlice(0, buf, slice, MarshalString)

	if err := SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
		return SkipSlice(n, b)
	}); err != nil {
		t.Fatal(err.Error())
	}

	_, retSlice, err := UnmarshalSlice[string](0, buf, UnmarshalString)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retSlice, slice) {
		t.Fatal("no match!")
	}

	t.Logf("org %v\ndec %v", slice, retSlice)
}

/*
func TestMaps(t *testing.T) {
	m := make(map[string]string)
	m["mapkey1"] = "mapvalue1"
	m["mapkey2"] = "mapvalue2"
	m["mapkey3"] = "mapvalue3"
	m["mapkey4"] = "mapvalue4"
	m["mapkey5"] = "mapvalue5"

	s, err := SizeMap(m, SizeString, SizeString)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, buf := benc.Marshal(s)
	_, err = MarshalMap(n, buf, m, MarshalString, MarshalString)

	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
		return SkipMap(n, b, SkipString, SkipString)
	}); err != nil {
		t.Fatal(err.Error())
	}

	_, retMap, err := UnmarshalMap(0, buf, UnmarshalString, UnmarshalString)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retMap, m) {
		t.Fatal("no match!")
	}

	t.Logf("org %v\ndec %v", m, retMap)
}*/

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
		t.Fatal("no match!")
	}

	t.Logf("org %v\ndec %v", str, retStr)
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
		t.Fatal("no match!")
	}

	t.Logf("org %v\ndec %v", str, retStr)
}
