package bstd

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/deneonet/benc"
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
	n, b := benc.Marshal(s)
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
		ts, err := SizeString(testStr)
		if err != nil {
			t.Fatal("at size string: error: " + err.Error())
			ts = 0
		}
		return ts
	}

	testBs := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	s := SizeAll(SizeBool, SizeBool, SizeByte, SizeFloat32, SizeFloat64, SizeInt16, SizeInt32, SizeInt64, SizeUInt16, SizeUInt32, SizeUInt64,
		sizeTestStr, sizeTestStr, func() int {
			ts, err := SizeByteSlice(testBs)
			if err != nil {
				t.Fatal("at size byteslice: error: " + err.Error())
				ts = 0
			}
			return ts
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
		func(n int, b []byte, v any) int {
			tn, err := MarshalString(n, b, v.(string))
			if err != nil {
				t.Fatal("at string: error: " + err.Error())
				tn = 0
			}
			return tn
		},
		func(n int, b []byte, v any) int {
			tn, err := MarshalUnsafeString(n, b, v.(string))
			if err != nil {
				t.Fatal("at unsafe string: error: " + err.Error())
				tn = 0
			}
			return tn
		},
		func(n int, b []byte, v any) int {
			tn, err := MarshalByteSlice(n, b, v.([]byte))
			if err != nil {
				t.Fatal("at byte slice: error: " + err.Error())
				tn = 0
			}
			return tn
		},
	)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipAll(buf, SkipBool, SkipBool, SkipByte, SkipFloat32, SkipFloat64, SkipInt16, SkipInt32, SkipInt64, SkipUInt16, SkipUInt32, SkipUInt64, SkipString, SkipString, SkipByteSlice); err != nil {
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
		func(n int, b []byte) (int, any, error) { return UnmarshalByteSlice(n, b) },
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
		func(n int, b []byte) (int, any, error) { return UnmarshalByteSlice(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByteSlice(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByteSlice(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByteSlice(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice(n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice(n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice(n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice(n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalMap(n, b, UnmarshalByte, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalMap(n, b, UnmarshalByte, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalMap(n, b, UnmarshalByte, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalMap(n, b, UnmarshalByte, UnmarshalByte) },
	); err != nil {
		t.Fatal(err.Error())
	}

	skipSliceOfBytes := func(n int, b []byte) (int, error) { return SkipSlice(n, b, SkipByte) }
	skipMapOfBytes := func(n int, b []byte) (int, error) { return SkipMap(n, b, SkipByte, SkipByte) }
	if err := SkipAll_VerifyError(benc.ErrBufTooSmall, buffers, SkipBool, SkipByte, SkipFloat32, SkipFloat64, SkipInt16, SkipInt32, SkipInt64, SkipUInt16, SkipUInt32, SkipUInt64, SkipString, SkipString, SkipString, SkipString, SkipString, SkipString, SkipString, SkipString, SkipByteSlice, SkipByteSlice, SkipByteSlice, SkipByteSlice, skipSliceOfBytes, skipSliceOfBytes, skipSliceOfBytes, skipSliceOfBytes, skipMapOfBytes, skipMapOfBytes, skipMapOfBytes, skipMapOfBytes); err != nil {
		t.Fatal(err.Error())
	}
}

func TestErrInvalidData(t *testing.T) {
	buffers := [][]byte{{2, 1, 1}, {4, 1, 2, 3, 4}, {8, 1, 2, 3, 4, 5, 6, 7, 8}, {2, 1, 1}, {4, 1, 2, 3, 4}, {8, 1, 2, 3, 4, 5, 6, 7, 8}, {2, 1, 1}, {4, 1, 2, 3, 4}, {8, 1, 2, 3, 4, 5, 6, 7, 8}, {2, 1, 1}, {4, 1, 2, 3, 4}, {8, 1, 2, 3, 4, 5, 6, 7, 8}, {2, 1, 1}, {4, 1, 2, 3, 4}, {8, 1, 2, 3, 4, 5, 6, 7, 8}}
	if err := UnmarshalAll_VerifyError(benc.ErrInvalidData, buffers,
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByteSlice(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByteSlice(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByteSlice(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice(n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice(n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice(n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalMap(n, b, UnmarshalByte, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalMap(n, b, UnmarshalByte, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalMap(n, b, UnmarshalByte, UnmarshalByte) },
	); err != nil {
		t.Fatal(err.Error())
	}

	skipSliceOfBytes := func(n int, b []byte) (int, error) { return SkipSlice(n, b, SkipByte) }
	skipMapOfBytes := func(n int, b []byte) (int, error) { return SkipMap(n, b, SkipByte, SkipByte) }
	if err := SkipAll_VerifyError(benc.ErrInvalidData, buffers, SkipString, SkipString, SkipString, SkipString, SkipString, SkipString, SkipByteSlice, SkipByteSlice, SkipByteSlice, skipSliceOfBytes, skipSliceOfBytes, skipSliceOfBytes, skipMapOfBytes, skipMapOfBytes, skipMapOfBytes); err != nil {
		t.Fatal(err.Error())
	}
}

func TestErrInvalidSize(t *testing.T) {
	buffers := [][]byte{{5}, {5}, {5}, {5}, {5}}
	if err := UnmarshalAll_VerifyError(benc.ErrInvalidSize, buffers,
		func(n int, b []byte) (int, any, error) { return UnmarshalString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalUnsafeString(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalByteSlice(n, b) },
		func(n int, b []byte) (int, any, error) { return UnmarshalSlice(n, b, UnmarshalByte) },
		func(n int, b []byte) (int, any, error) { return UnmarshalMap(n, b, UnmarshalByte, UnmarshalByte) },
	); err != nil {
		t.Fatal(err.Error())
	}

	if err := SkipAll_VerifyError(benc.ErrInvalidSize, buffers, SkipString, SkipString, SkipByteSlice, func(n int, b []byte) (int, error) { return SkipSlice(n, b, SkipByte) }, func(n int, b []byte) (int, error) { return SkipMap(n, b, SkipByte, SkipByte) }); err != nil {
		t.Fatal(err.Error())
	}
}

func TestSlices(t *testing.T) {
	slice := []string{"sliceelement1", "sliceelement2", "sliceelement3", "sliceelement4", "sliceelement5"}
	s, err := SizeSlice(slice, SizeString)

	if err != nil {
		t.Fatal(err.Error())
	}

	n, buf := benc.Marshal(s)
	_, err = MarshalSlice(n, buf, slice, MarshalString)

	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
		return SkipSlice(n, b, SkipString)
	}); err != nil {
		t.Fatal(err.Error())
	}

	_, retSlice, err := UnmarshalSlice(0, buf, UnmarshalString)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retSlice, slice) {
		t.Fatal("no match!")
	}

	t.Logf("org %v\ndec %v", slice, retSlice)
}

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
}

func TestLongSlices(t *testing.T) {
	slice := make([]uint16, math.MaxUint16+1)

	_, err := SizeSlice(slice, SizeUInt16)
	if err != benc.ErrDataTooBig {
		t.Fatal("size slice should return a `benc.ErrDataTooBig` error")
	}

	s, err := SizeSlice(slice, SizeUInt16, benc.Bytes4)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, buf := benc.Marshal(s)

	_, err = MarshalSlice(n, buf, slice, MarshalUInt16)
	if err != benc.ErrDataTooBig {
		t.Fatal("marshal slice should return a `benc.ErrDataTooBig` error")
	}

	_, err = MarshalSlice(n, buf, slice, MarshalUInt16, benc.Bytes4)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
		return SkipSlice(n, b, SkipUInt16)
	}); err != nil {
		t.Fatal(err.Error())
	}

	_, retSlice, err := UnmarshalSlice(0, buf, UnmarshalUInt16)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retSlice, slice) {
		t.Fatal("no match!")
	}

	t.Logf("org %v\ndec %v", slice, retSlice)
}

func generateBigMap() map[uint32]uint16 {
	m := make(map[uint32]uint16)
	for i := 0; i < math.MaxUint16+2; i++ {
		m[uint32(rand.Intn(math.MaxUint32))] = 0
	}
	if len(m) <= math.MaxUint16 {
		return generateBigMap()
	}
	return m
}

func TestBigMaps(t *testing.T) {
	m := generateBigMap()

	_, err := SizeMap(m, SizeUInt32, SizeUInt16)
	if err != benc.ErrDataTooBig {
		t.Fatal("size map should return a `benc.ErrDataTooBig` error")
	}

	s, err := SizeMap(m, SizeUInt32, SizeUInt16, benc.Bytes4)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, buf := benc.Marshal(s)

	_, err = MarshalMap(n, buf, m, MarshalUInt32, MarshalUInt16)
	if err != benc.ErrDataTooBig {
		t.Fatal("marshal map should return a `benc.ErrDataTooBig` error")
	}

	_, err = MarshalMap(n, buf, m, MarshalUInt32, MarshalUInt16, benc.Bytes4)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
		return SkipMap(n, b, SkipUInt32, SkipUInt16)
	}); err != nil {
		t.Fatal(err.Error())
	}

	_, retMap, err := UnmarshalMap(0, buf, UnmarshalUInt32, UnmarshalUInt16)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retMap, m) {
		t.Fatal("no match!")
	}

	t.Logf("org %v\ndec %v", m, retMap)
}

func generateString(l int) string {
	s := ""
	for i := 0; i < l; i++ {
		s += "T"
	}
	return s
}

func TestLongStrings(t *testing.T) {
	str := generateString(math.MaxUint16 + 1)

	_, err := SizeString(str)
	if err != benc.ErrDataTooBig {
		t.Fatal("size string should return a `benc.ErrDataTooBig` error")
	}

	s, err := SizeString(str, benc.Bytes4)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, buf := benc.Marshal(s)

	_, err = MarshalString(n, buf, str)
	if err != benc.ErrDataTooBig {
		t.Fatal("marshal string should return a `benc.ErrDataTooBig` error")
	}

	_, err = MarshalString(n, buf, str, benc.Bytes4)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipOnce_Verify(buf, SkipString); err != nil {
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

func TestLongUnsafeStrings(t *testing.T) {
	str := generateString(math.MaxUint16 + 1)

	_, err := SizeString(str)
	if err != benc.ErrDataTooBig {
		t.Fatal("size string should return a `benc.ErrDataTooBig` error")
	}

	s, err := SizeString(str, benc.Bytes4)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, buf := benc.Marshal(s)

	_, err = MarshalUnsafeString(n, buf, str)
	if err != benc.ErrDataTooBig {
		t.Fatal("marshal string should return a `benc.ErrDataTooBig` error")
	}

	_, err = MarshalUnsafeString(n, buf, str, benc.Bytes4)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipOnce_Verify(buf, SkipString); err != nil {
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

func TestLongByteSlices(t *testing.T) {
	slice := make([]byte, math.MaxUint16+1)

	_, err := SizeByteSlice(slice)
	if err != benc.ErrDataTooBig {
		t.Fatal("size byte slice should return a `benc.ErrDataTooBig` error")
	}

	s, err := SizeByteSlice(slice, benc.Bytes4)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, buf := benc.Marshal(s)

	_, err = MarshalByteSlice(n, buf, slice)
	if err != benc.ErrDataTooBig {
		t.Fatal("marshal byte slice should return a `benc.ErrDataTooBig` error")
	}

	_, err = MarshalByteSlice(n, buf, slice, benc.Bytes4)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipOnce_Verify(buf, SkipByteSlice); err != nil {
		t.Fatal(err.Error())
	}

	_, retSlice, err := UnmarshalByteSlice(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retSlice, slice) {
		t.Fatal("no match!")
	}

	t.Logf("org %v\ndec %v", slice, retSlice)
}

func TestEmptySlices(t *testing.T) {
	slice := []string{}
	s, err := SizeSlice(slice, SizeString)

	if err != nil {
		t.Fatal(err.Error())
	}

	n, buf := benc.Marshal(s)
	_, err = MarshalSlice(n, buf, slice, MarshalString)

	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipOnce_Verify(buf, func(n int, b []byte) (int, error) {
		return SkipSlice(n, b, SkipString)
	}); err != nil {
		t.Fatal(err.Error())
	}

	_, retSlice, err := UnmarshalSlice(0, buf, UnmarshalString)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retSlice, slice) {
		t.Fatal("no match!")
	}

	t.Logf("org %v\ndec %v", slice, retSlice)
}

func TestEmptyByteSlice(t *testing.T) {
	str := []byte{}

	s, err := SizeByteSlice(str)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, buf := benc.Marshal(s)
	_, err = MarshalByteSlice(n, buf, str)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipOnce_Verify(buf, SkipByteSlice); err != nil {
		t.Fatal(err.Error())
	}

	_, retStr, err := UnmarshalByteSlice(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(retStr, str) {
		t.Fatal("no match!")
	}

	t.Logf("org %v\ndec %v", str, retStr)
}

func TestEmptyString(t *testing.T) {
	str := ""

	s, err := SizeString(str)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, buf := benc.Marshal(s)
	_, err = MarshalString(n, buf, str)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipOnce_Verify(buf, SkipString); err != nil {
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

	s, err := SizeString(str)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, buf := benc.Marshal(s)
	_, err = MarshalUnsafeString(n, buf, str)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err = SkipOnce_Verify(buf, SkipString); err != nil {
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

func TestMessageFraming(t *testing.T) {
	var buf bytes.Buffer

	testStr := "Hello World!"
	s, err := SizeString(testStr)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, b := benc.MarshalMF(s)
	_, err = MarshalString(n, b, testStr)
	if err != nil {
		t.Fatal(err.Error())
	}

	// concatenated bytes of serialized "Hello World!" in the benc format
	buf.Write(b)
	buf.Write(b)

	_, err = benc.UnmarshalMF([]byte{1, 2})
	if err != benc.ErrBufTooSmall {
		t.Fatal("expected a `benc.ErrBufTooSmall` error")
	}

	_, err = benc.UnmarshalMF([]byte{1, 2, 3})
	if err != benc.ErrBufTooSmall {
		t.Fatal("expected a `benc.ErrBufTooSmall` error")
	}

	unconcatenatedBytes, err := benc.UnmarshalMF(buf.Bytes())
	if err != nil {
		t.Fatal(err.Error())
	}

	var i int
	for _, bytes := range unconcatenatedBytes {
		_, str, err := UnmarshalString(0, bytes)
		if err != nil {
			t.Fatal(err.Error())
		}

		if str != testStr {
			t.Fatal("no match!")
		}
		i++
	}

	if i != 2 {
		t.Fatal("i is not 2!")
	}
}
