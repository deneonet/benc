package bstd

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/deneonet/benc"
)

func SizeAll(sizers ...func() int) (s int) {
	for _, sizer := range sizers {
		ts := sizer()
		if ts == 0 {
			return 0
		}
		s += ts
	}

	return
}

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
			return fmt.Errorf("(skip) at idx %d: expected a %s error, got %s", i, expected, err)
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

	values := []any{
		true,
		byte(128),
		rand.Float32(),
		rand.Float64(),
		int(math.MaxInt),
		int16(16),
		rand.Int31(),
		rand.Int63(),
		uint(math.MaxUint),
		uint16(160),
		rand.Uint32(),
		rand.Uint64(),
		testStr,
		testStr,
		testBs,
	}

	s := SizeAll(SizeBool, SizeByte, SizeFloat32, SizeFloat64, func() int { return SizeInt(math.MaxInt) }, SizeInt16, SizeInt32, SizeInt64, func() int { return SizeUint(math.MaxUint) }, SizeUint16, SizeUint32, SizeUint64,
		sizeTestStr, sizeTestStr, sizeTestBs)

	b, err := MarshalAll(s, values,
		func(n int, b []byte, v any) int { return MarshalBool(n, b, v.(bool)) },
		func(n int, b []byte, v any) int { return MarshalByte(n, b, v.(byte)) },
		func(n int, b []byte, v any) int { return MarshalFloat32(n, b, v.(float32)) },
		func(n int, b []byte, v any) int { return MarshalFloat64(n, b, v.(float64)) },
		func(n int, b []byte, v any) int { return MarshalInt(n, b, v.(int)) },
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
	)

	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipAll(b, SkipBool, SkipByte, SkipFloat32, SkipFloat64, SkipVarint, SkipInt16, SkipInt32, SkipInt64, SkipVarint, SkipUint16, SkipUint32, SkipUint64, SkipString, SkipString, SkipBytes); err != nil {
		t.Fatal(err.Error())
	}

	if err = UnmarshalAll(b, values,
		func(n int, b []byte) (int, any, error) { return UnmarshalBool(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByte(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalFloat32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalFloat64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt(n, b) },
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
	); err != nil {
		t.Fatal(err.Error())
	}
}

func TestErrBufTooSmall(t *testing.T) {
	buffers := [][]byte{{}, {}, {1, 2, 3}, {1, 2, 3, 4, 5, 6, 7}, {}, {1}, {1, 2, 3}, {1, 2, 3, 4, 5, 6, 7}, {}, {1}, {1, 2, 3}, {1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7}, {}, {2, 0}, {4, 1, 2, 3}, {8, 1, 2, 3, 4, 5, 6, 7}}
	if err := UnmarshalAll_VerifyError(benc.ErrBufTooSmall, buffers,
		func(n int, b []byte) (int, any, error) { return UnmarshalBool(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByte(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalFloat32(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalFloat64(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalInt(n, b) },
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
	skipMapOfBytes := func(n int, b []byte) (int, error) { return SkipMap(n, b) }
	if err := SkipAll_VerifyError(benc.ErrBufTooSmall, buffers, SkipBool, SkipByte, SkipFloat32, SkipFloat64, SkipVarint, SkipInt16, SkipInt32, SkipInt64, SkipVarint, SkipUint16, SkipUint32, SkipUint64, SkipString, SkipString, SkipString, SkipString, SkipString, SkipString, SkipString, SkipString, SkipBytes, SkipBytes, SkipBytes, SkipBytes, skipSliceOfBytes, skipSliceOfBytes, skipSliceOfBytes, skipSliceOfBytes, skipMapOfBytes, skipMapOfBytes, skipMapOfBytes, skipMapOfBytes); err != nil {
		t.Fatal(err.Error())
	}
}

func TestErrBufTooSmall_2(t *testing.T) {
	buffers := [][]byte{{}, {2, 0}, {}, {2, 0}, {}, {2, 0}, {}, {10, 0, 0, 0, 1}, {}, {10, 0, 0, 0, 1}}
	if err := UnmarshalAll_VerifyError(benc.ErrBufTooSmall, buffers,
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCropped(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalBytesCropped(n, b) },
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

	skipSliceOfBytes := func(n int, b []byte) (int, error) { return SkipSlice(n, b) }
	skipMapOfBytes := func(n int, b []byte) (int, error) { return SkipMap(n, b) }
	if err := SkipAll_VerifyError(benc.ErrBufTooSmall, buffers, SkipString, SkipString, SkipString, SkipString, SkipBytes, SkipBytes, skipSliceOfBytes, skipSliceOfBytes, skipMapOfBytes, skipMapOfBytes); err != nil {
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
	fmt.Println(buf)

	if err := SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
		return SkipMap(n, b)
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
	fmt.Println(buf)

	if err := SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
		return SkipMap(n, b)
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
