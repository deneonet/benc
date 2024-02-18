package bstd

import (
	"bytes"
	"github.com/deneonet/benc/bpre"
	"testing"
	"time"

	"github.com/deneonet/benc/bmd"
	"github.com/deneonet/benc/btag"
	"github.com/deneonet/benc/bunsafe"
)

func TestDataTypes_StringTag_Metadata(t *testing.T) {
	s := bmd.SizeBool()
	s += bmd.SizeByte()
	s += bmd.SizeFloat32()
	s += bmd.SizeFloat64()
	s += bmd.SizeInt()
	s += bmd.SizeInt16()
	s += bmd.SizeInt32()
	s += bmd.SizeInt64()
	s += bmd.SizeString("H")
	s += bmd.SizeTime()
	s += bmd.SizeUInt()
	s += bmd.SizeUInt16()
	s += bmd.SizeUInt32()
	s += bmd.SizeInt64()
	s += bmd.SizeByteSlice([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	n, buf := btag.SMarshalMD(s, "v1")
	n = bmd.MarshalBool(n, buf, true)
	n = bmd.MarshalByte(n, buf, 1)
	n = bmd.MarshalFloat32(n, buf, 1)
	n = bmd.MarshalFloat64(n, buf, 1)
	n = bmd.MarshalInt(n, buf, 1)
	n = bmd.MarshalInt16(n, buf, 1)
	n = bmd.MarshalInt32(n, buf, 1)
	n = bmd.MarshalInt64(n, buf, 1)
	n = bunsafe.MarshalStringMD(n, buf, "H")
	n = bmd.MarshalTime(n, buf, time.Now())
	n = bmd.MarshalFloat64(n, buf, 0)
	n = bmd.MarshalUInt16(n, buf, 0)
	n = bmd.MarshalUInt32(n, buf, 0)
	n = bmd.MarshalUInt64(n, buf, 0)
	n = bmd.MarshalByteSlice(n, buf, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	n, _, err := btag.UUnmarshalMD(0, buf)
	if err == nil {
		t.Fatal("should return error")
	}
	n, _, err = bmd.UnmarshalByte(n, buf)
	if err == nil {
		t.Fatal("should return error")
	}
	n, _, err = bmd.UnmarshalBool(n, buf)
	if err == nil {
		t.Fatal("should return error")
	}
	n, _, err = bmd.UnmarshalInt32(n, buf)
	if err == nil {
		t.Fatal("should return error")
	}
	n, _, err = bmd.UnmarshalUInt(n, buf)
	if err == nil {
		t.Fatal("should return error")
	}
	n, _, err = bmd.UnmarshalInt64(n, buf)
	if err == nil {
		t.Fatal("should return error")
	}
	n, _, err = bmd.UnmarshalUInt16(n, buf)
	if err == nil {
		t.Fatal("should return error")
	}
	n, _, err = bmd.UnmarshalFloat32(n, buf)
	if err == nil {
		t.Fatal("should return error")
	}
	n, _, err = bmd.UnmarshalFloat64(n, buf)
	if err == nil {
		t.Fatal("should return error")
	}
	n, _, err = bmd.UnmarshalUInt64(n, buf)
	if err == nil {
		t.Fatal("should return error")
	}
	n, _, err = bmd.UnmarshalInt(n, buf)
	if err == nil {
		t.Fatal("should return error")
	}
	n, _, err = bmd.UnmarshalTime(n, buf)
	if err == nil {
		t.Fatal("should return error")
	}
	n, _, err = btag.SUnmarshalMD(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalBool(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalByte(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := VerifyUnmarshal(n, buf); err == nil {
		// will always fail
		t.Fatal("should return error")
	}
}
func TestDataTypes_UIntTag_Metadata(t *testing.T) {
	s := bmd.SizeBool()
	s += bmd.SizeByte()
	s += bmd.SizeFloat32()
	s += bmd.SizeFloat64()
	s += bmd.SizeInt()
	s += bmd.SizeInt16()
	s += bmd.SizeInt32()
	s += bmd.SizeInt64()
	s += bmd.SizeString("H")
	s += bmd.SizeTime()
	s += bmd.SizeUInt()
	s += bmd.SizeUInt16()
	s += bmd.SizeUInt32()
	s += bmd.SizeInt64()
	s += bmd.SizeByteSlice([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	n, buf := btag.UMarshalMD(s, 1)
	n = bmd.MarshalBool(n, buf, true)
	n = bmd.MarshalByte(n, buf, 1)
	n = bmd.MarshalFloat32(n, buf, 1)
	n = bmd.MarshalFloat64(n, buf, 1)
	n = bmd.MarshalInt(n, buf, 1)
	n = bmd.MarshalInt16(n, buf, 1)
	n = bmd.MarshalInt32(n, buf, 1)
	n = bmd.MarshalInt64(n, buf, 1)
	n = bunsafe.MarshalStringMD(n, buf, "H")
	n = bmd.MarshalTime(n, buf, time.Now())
	n = bmd.MarshalUInt(n, buf, 0)
	n = bmd.MarshalUInt16(n, buf, 0)
	n = bmd.MarshalUInt32(n, buf, 0)
	n = bmd.MarshalUInt64(n, buf, 0)
	n = bmd.MarshalByteSlice(n, buf, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
	n, tag, err := btag.UUnmarshalMD(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	if tag != 1 {
		t.Fatal("tag doesn't match")
	}
	n, _, err = bmd.UnmarshalBool(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalByte(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalFloat32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalFloat64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bunsafe.UnmarshalStringMD(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalTime(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalUInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalUInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalUInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalUInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bmd.UnmarshalByteSlice(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}

func TestSkippingDataTypes_StringTag_Metadata(t *testing.T) {
	s := bmd.SizeBool()
	s += bmd.SizeByte()
	s += bmd.SizeFloat32()
	s += bmd.SizeFloat64()
	s += bmd.SizeInt()
	s += bmd.SizeInt16()
	s += bmd.SizeInt32()
	s += bmd.SizeInt64()
	s += bmd.SizeString("H")
	s += bmd.SizeTime()
	s += bmd.SizeUInt()
	s += bmd.SizeUInt16()
	s += bmd.SizeUInt32()
	s += bmd.SizeUInt64()
	s += bmd.SizeByteSlice([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	n, buf := btag.SMarshalMD(s, "v1")

	n = bmd.MarshalBool(n, buf, true)
	n = bmd.MarshalByte(n, buf, 1)
	n = bmd.MarshalFloat32(n, buf, 1)
	n = bmd.MarshalFloat64(n, buf, 1)
	n = bmd.MarshalInt(n, buf, 1)
	n = bmd.MarshalInt16(n, buf, 1)
	n = bmd.MarshalInt32(n, buf, 1)
	n = bmd.MarshalInt64(n, buf, 1)
	n = bunsafe.MarshalStringMD(n, buf, "H")
	n = bmd.MarshalTime(n, buf, time.Now())
	n = bmd.MarshalUInt(n, buf, 0)
	n = bmd.MarshalUInt16(n, buf, 0)
	n = bmd.MarshalUInt32(n, buf, 0)
	n = bmd.MarshalUInt64(n, buf, 0)
	n = bmd.MarshalByteSlice(n, buf, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
	n, err := bmd.SkipStringTag(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipBool(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipByte(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipFloat32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipFloat64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipString(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipTime(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipUInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipUInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipUInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipUInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipByteSlice(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}
func TestSkippingDataTypes_UIntTag_Metadata(t *testing.T) {
	s := bmd.SizeBool()
	s += bmd.SizeByte()
	s += bmd.SizeFloat32()
	s += bmd.SizeFloat64()
	s += bmd.SizeInt()
	s += bmd.SizeInt16()
	s += bmd.SizeInt32()
	s += bmd.SizeInt64()
	s += bmd.SizeString("H")
	s += bmd.SizeTime()
	s += bmd.SizeUInt()
	s += bmd.SizeUInt16()
	s += bmd.SizeUInt32()
	s += bmd.SizeInt64()
	s += bmd.SizeByteSlice([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	n, buf := btag.UMarshalMD(s, 1)
	n = bmd.MarshalBool(n, buf, true)
	n = bmd.MarshalByte(n, buf, 1)
	n = bmd.MarshalFloat32(n, buf, 1)
	n = bmd.MarshalFloat64(n, buf, 1)
	n = bmd.MarshalInt(n, buf, 1)
	n = bmd.MarshalInt16(n, buf, 1)
	n = bmd.MarshalInt32(n, buf, 1)
	n = bmd.MarshalInt64(n, buf, 1)
	n = bunsafe.MarshalStringMD(n, buf, "H")
	n = bmd.MarshalTime(n, buf, time.Now())
	n = bmd.MarshalUInt(n, buf, 0)
	n = bmd.MarshalUInt16(n, buf, 0)
	n = bmd.MarshalUInt32(n, buf, 0)
	n = bmd.MarshalUInt64(n, buf, 0)
	n = bmd.MarshalByteSlice(n, buf, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	n, err := bmd.SkipUIntTag(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, err = bmd.SkipBool(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipByte(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipFloat32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipFloat64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipString(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipTime(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipUInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipUInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipUInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipUInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = bmd.SkipByteSlice(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}

func TestDataTypes_StringTag(t *testing.T) {
	s := SizeBool()
	s += SizeByte()
	s += SizeFloat32()
	s += SizeFloat64()
	s += SizeInt()
	s += SizeInt16()
	s += SizeInt32()
	s += SizeInt64()
	s += SizeString("H")
	s += SizeTime()
	s += SizeUInt()
	s += SizeUInt16()
	s += SizeUInt32()
	s += SizeUInt64()
	s += SizeByteSlice([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	n, buf := btag.SMarshal(s, "v1")
	n = MarshalBool(n, buf, true)
	n = MarshalByte(n, buf, 1)
	n = MarshalFloat32(n, buf, 1)
	n = MarshalFloat64(n, buf, 1)
	n = MarshalInt(n, buf, 1)
	n = MarshalInt16(n, buf, 1)
	n = MarshalInt32(n, buf, 1)
	n = MarshalInt64(n, buf, 1)
	n = bunsafe.MarshalString(n, buf, "H")
	n = MarshalTime(n, buf, time.Now())
	n = MarshalUInt(n, buf, 0)
	n = MarshalUInt16(n, buf, 0)
	n = MarshalUInt32(n, buf, 0)
	n = MarshalUInt64(n, buf, 0)
	n = MarshalByteSlice(n, buf, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	n, tag, err := btag.SUnmarshal(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	if tag != "v1" {
		t.Fatal("tag doesn't match")
	}
	n, _, err = UnmarshalBool(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalByte(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalFloat32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalFloat64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bunsafe.UnmarshalString(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalTime(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalUInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalUInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalUInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalUInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalByteSlice(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}
func TestDataTypes_UIntTag(t *testing.T) {
	s := SizeBool()
	s += SizeByte()
	s += SizeFloat32()
	s += SizeFloat64()
	s += SizeInt()
	s += SizeInt16()
	s += SizeInt32()
	s += SizeInt64()
	s += SizeString("H")
	s += SizeTime()
	s += SizeUInt()
	s += SizeUInt16()
	s += SizeUInt32()
	s += SizeInt64()
	n, buf := btag.UMarshal(s, 1)
	n = MarshalBool(n, buf, true)
	n = MarshalByte(n, buf, 1)
	n = MarshalFloat32(n, buf, 1)
	n = MarshalFloat64(n, buf, 1)
	n = MarshalInt(n, buf, 1)
	n = MarshalInt16(n, buf, 1)
	n = MarshalInt32(n, buf, 1)
	n = MarshalInt64(n, buf, 1)
	n = bunsafe.MarshalString(n, buf, "H")
	n = MarshalTime(n, buf, time.Now())
	n = MarshalUInt(n, buf, 0)
	n = MarshalUInt16(n, buf, 0)
	n = MarshalUInt32(n, buf, 0)
	n = MarshalUInt64(n, buf, 0)
	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
	n, tag, err := btag.UUnmarshal(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	if tag != 1 {
		t.Fatal("tag doesn't match")
	}
	n, _, err = UnmarshalBool(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalByte(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalFloat32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalFloat64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = bunsafe.UnmarshalString(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalTime(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalUInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalUInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalUInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, _, err = UnmarshalUInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}

func TestSkippingDataTypes_StringTag(t *testing.T) {
	s := SizeBool()
	s += SizeByte()
	s += SizeFloat32()
	s += SizeFloat64()
	s += SizeInt()
	s += SizeInt16()
	s += SizeInt32()
	s += SizeInt64()
	s += SizeString("H")
	s += SizeTime()
	s += SizeUInt()
	s += SizeUInt16()
	s += SizeUInt32()
	s += SizeInt64()
	s += SizeByteSlice([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	n, buf := btag.SMarshal(s, "v1")

	n = MarshalBool(n, buf, true)
	n = MarshalByte(n, buf, 1)
	n = MarshalFloat32(n, buf, 1)
	n = MarshalFloat64(n, buf, 1)
	n = MarshalInt(n, buf, 1)
	n = MarshalInt16(n, buf, 1)
	n = MarshalInt32(n, buf, 1)
	n = MarshalInt64(n, buf, 1)
	n = bunsafe.MarshalString(n, buf, "H")
	n = MarshalTime(n, buf, time.Now())
	n = MarshalUInt(n, buf, 0)
	n = MarshalUInt16(n, buf, 0)
	n = MarshalUInt32(n, buf, 0)
	n = MarshalUInt64(n, buf, 0)
	n = MarshalByteSlice(n, buf, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	n, err := SkipStringTag(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipBool(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipByte(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipFloat32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipFloat64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipString(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipTime(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipUInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipUInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipUInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipUInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipByteSlice(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}
func TestSkippingDataTypes_UIntTag(t *testing.T) {
	s := SizeBool()
	s += SizeByte()
	s += SizeFloat32()
	s += SizeFloat64()
	s += SizeInt()
	s += SizeInt16()
	s += SizeInt32()
	s += SizeInt64()
	s += SizeString("H")
	s += SizeTime()
	s += SizeUInt()
	s += SizeUInt16()
	s += SizeUInt32()
	s += SizeInt64()

	n, buf := btag.UMarshal(s, 1)
	n = MarshalBool(n, buf, true)
	n = MarshalByte(n, buf, 1)
	n = MarshalFloat32(n, buf, 1)
	n = MarshalFloat64(n, buf, 1)
	n = MarshalInt(n, buf, 1)
	n = MarshalInt16(n, buf, 1)
	n = MarshalInt32(n, buf, 1)
	n = MarshalInt64(n, buf, 1)
	n = bunsafe.MarshalString(n, buf, "H")
	n = MarshalTime(n, buf, time.Now())
	n = MarshalUInt(n, buf, 0)
	n = MarshalUInt16(n, buf, 0)
	n = MarshalUInt32(n, buf, 0)
	n = MarshalUInt64(n, buf, 0)

	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	n, err := SkipUIntTag(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, err = SkipBool(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipByte(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipFloat32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipFloat64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipString(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipTime(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipUInt(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipUInt16(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipUInt32(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipUInt64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}

func TestSliceAndMap(t *testing.T) {
	sliceData := []string{"DATA_1", "DATA_2"}

	mapData := make(map[string]float64)
	mapData["DATA_1"] = 13531.523400123
	mapData["DATA_2"] = 2561.1512312313

	s := SizeSlice(sliceData, SizeString)
	s += SizeMap(mapData, SizeString, SizeFloat64)

	n, buf := Marshal(s)
	n = MarshalSlice(n, buf, sliceData, MarshalString)
	n = MarshalMap(n, buf, mapData, MarshalString, MarshalFloat64)
	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	var err error
	n, sliceData, err = UnmarshalSlice(0, buf, UnmarshalString)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, mapData, err = UnmarshalMap(n, buf, UnmarshalString, UnmarshalFloat64)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	if sliceData[0] != "DATA_1" || sliceData[1] != "DATA_2" {
		t.Fatal("slice doesn't match")
	}

	if mapData["DATA_1"] != 13531.523400123 || mapData["DATA_2"] != 2561.1512312313 {
		t.Fatal("map doesn't match")
	}
}
func TestSkippingSliceAndMap(t *testing.T) {
	sliceData := []string{"DATA_1", "DATA_2"}

	mapData := make(map[string]float64)
	mapData["DATA_1"] = 13531.523400123
	mapData["DATA_2"] = 2561.1512312313

	s := SizeSlice(sliceData, SizeString)
	s += SizeMap(mapData, SizeString, SizeInt)

	n, buf := Marshal(s)
	n = MarshalSlice(n, buf, sliceData, MarshalString)
	n = MarshalMap(n, buf, mapData, MarshalString, MarshalFloat64)
	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	var err error
	n, err = SkipSlice(0, buf, SkipString)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, err = SkipMap(n, buf, SkipString, SkipFloat64)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}

func TestSliceAndMap_Metadata(t *testing.T) {
	sliceData := []string{"DATA_1", "DATA_2"}

	mapData := make(map[string]float64)
	mapData["DATA_1"] = 13531.523400123
	mapData["DATA_2"] = 2561.1512312313

	s := bmd.SizeSlice(sliceData, bmd.SizeString)
	s += bmd.SizeMap(mapData, bmd.SizeString, bmd.SizeFloat64)

	n, buf := Marshal(s)
	n = bmd.MarshalSlice(n, buf, sliceData, bmd.MarshalString)
	n = bmd.MarshalMap(n, buf, mapData, bmd.MarshalString, bmd.MarshalFloat64)
	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	var err error
	n, sliceData, err = bmd.UnmarshalSlice(0, buf, bmd.UnmarshalString)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, mapData, err = bmd.UnmarshalMap(n, buf, bmd.UnmarshalString, bmd.UnmarshalFloat64)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	if sliceData[0] != "DATA_1" || sliceData[1] != "DATA_2" {
		t.Fatal("slice doesn't match")
	}

	if mapData["DATA_1"] != 13531.523400123 || mapData["DATA_2"] != 2561.1512312313 {
		t.Fatal("map doesn't match")
	}
}
func TestSkippingSliceAndMap_Metadata(t *testing.T) {
	sliceData := []string{"DATA_1", "DATA_2"}

	mapData := make(map[string]float64)
	mapData["DATA_1"] = 13531.523400123
	mapData["DATA_2"] = 2561.1512312313

	s := bmd.SizeSlice(sliceData, bmd.SizeString)
	s += bmd.SizeMap(mapData, bmd.SizeString, bmd.SizeInt)

	n, buf := Marshal(s)
	n = bmd.MarshalSlice(n, buf, sliceData, bmd.MarshalString)
	n = bmd.MarshalMap(n, buf, mapData, bmd.MarshalString, bmd.MarshalFloat64)
	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	var err error
	n, err = bmd.SkipSlice(0, buf, bmd.SkipString)
	if err != nil {
		t.Fatal(err.Error())
	}

	n, err = bmd.SkipMap(n, buf, bmd.SkipString, bmd.SkipFloat64)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}

func TestMessageFraming(t *testing.T) {
	var buffer bytes.Buffer

	s := SizeString("Hello World!")
	s += SizeFloat64()

	n, buf := MarshalMF(s)
	n = MarshalString(n, buf, "Hello World!")
	n = MarshalFloat64(n, buf, 1231.5131)
	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	// Write the byte slice containing the encoded data twice into buffer
	// = two concatenated BENC encoded byte slices
	buffer.Write(buf)
	buffer.Write(buf)

	// Extracts the two concatenated byte slices, into a slice of byte slices
	data, err := UnmarshalMF(buffer.Bytes())
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, bs := range data {
		var helloWorld string
		n, helloWorld, err = bunsafe.UnmarshalString(0, bs)
		if err != nil {
			t.Fatal(err.Error())
		}
		if helloWorld != "Hello World!" {
			t.Fatal("helloWorld: string doesn't match")
		}

		var randomFloat64 float64
		n, randomFloat64, err = UnmarshalFloat64(n, bs)
		if err != nil {
			t.Fatal(err.Error())
		}
		if randomFloat64 != 1231.5131 {
			t.Fatal("randomFloat64: float64 doesn't match")
		}
	}

	if err := VerifyUnmarshalMF(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}
func TestMessageFraming_StringTag(t *testing.T) {
	var buffer bytes.Buffer

	s := SizeString("Hello World!")
	s += SizeFloat64()

	n, buf := btag.SMarshalMF(s, "v1")
	n = MarshalString(n, buf, "Hello World!")
	n = MarshalFloat64(n, buf, 1231.5131)
	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	// Write the byte slice containing the encoded data twice into buffer
	// = two concatenated BENC encoded byte slices
	buffer.Write(buf)
	buffer.Write(buf)

	// Extracts the two concatenated byte slices, into a slice of byte slices
	data, err := UnmarshalMF(buffer.Bytes())
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, bs := range data {
		var tag string
		n, tag, err = btag.SUnmarshal(0, bs)
		if err != nil {
			t.Fatal(err.Error())
		}
		if tag != "v1" {
			t.Fatal("tag: string doesn't match")
		}

		var helloWorld string
		n, helloWorld, err = UnmarshalString(n, bs)
		if err != nil {
			t.Fatal(err.Error())
		}
		if helloWorld != "Hello World!" {
			t.Fatal("helloWorld: string doesn't match")
		}

		var randomFloat64 float64
		n, randomFloat64, err = UnmarshalFloat64(n, bs)
		if err != nil {
			t.Fatal(err.Error())
		}
		if randomFloat64 != 1231.5131 {
			t.Fatal("randomFloat64: float64 doesn't match")
		}
	}

	if err := VerifyUnmarshalMF(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}
func TestMessageFraming_UIntTag(t *testing.T) {
	var buffer bytes.Buffer

	s := SizeString("Hello World!")
	s += SizeFloat64()

	n, buf := btag.UMarshalMF(s, 1)
	n = MarshalString(n, buf, "Hello World!")
	n = MarshalFloat64(n, buf, 1231.5131)
	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	// Write the byte slice containing the encoded data twice into buffer
	// = two concatenated BENC encoded byte slices
	buffer.Write(buf)
	buffer.Write(buf)

	bpre.UnmarshalMF(2000)

	// Extracts the two concatenated byte slices, into a slice of byte slices
	data, err := UnmarshalMF(buffer.Bytes())
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, bs := range data {
		var tag uint16
		n, tag, err = btag.UUnmarshal(0, bs)
		if err != nil {
			t.Fatal(err.Error())
		}
		if tag != 1 {
			t.Fatal("tag: string doesn't match")
		}

		var helloWorld string
		n, helloWorld, err = UnmarshalString(n, bs)
		if err != nil {
			t.Fatal(err.Error())
		}
		if helloWorld != "Hello World!" {
			t.Fatal("helloWorld: string doesn't match")
		}

		var randomFloat64 float64
		n, randomFloat64, err = UnmarshalFloat64(n, bs)
		if err != nil {
			t.Fatal(err.Error())
		}
		if randomFloat64 != 1231.5131 {
			t.Fatal("randomFloat64: float64 doesn't match")
		}
	}

	if err := VerifyUnmarshalMF(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}

func TestPreAllocation(t *testing.T) {
	// pre-allocates a byte slice of size 1000
	bpre.Marshal(1000)

	s := SizeString("Hello World!")
	s += SizeFloat64()

	// doesn't allocate any memory now, because it takes from the pre-allocated byte slice the size needed
	n, buf := Marshal(s)
	n = bunsafe.MarshalString(n, buf, "Hello World!")
	n = MarshalFloat64(n, buf, 1231.5131)
	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	n, err := SkipString(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, err = SkipFloat64(n, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
	bpre.Reset()
}
func TestOutOfOrderDeserialization(t *testing.T) {
	s := SizeString("Hello World!")
	s += SizeFloat64()
	s += SizeFloat32()

	n, buf := Marshal(s)

	// Marshal - Order:
	// Hello World! : bunsafe.UnmarshalString(...)
	// 1231.5131 : UnmarshalFloat64(...)
	// 1231.5132 : UnmarshalFloat32(...)

	n = bunsafe.MarshalString(n, buf, "Hello World!")
	n = MarshalFloat64(n, buf, 1231.5131)
	n = MarshalFloat32(n, buf, 1231.5132)
	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}

	// Unmarshal - Order:
	// 1231.5131 : UnmarshalFloat64(...)
	// Hello World! : bunsafe.UnmarshalString(...)
	// 1231.5132 : UnmarshalFloat32(...)

	n, err := SkipString(0, buf)
	if err != nil {
		t.Fatal(err.Error())
	}

	var randomFloat64 float64
	n, randomFloat64, err = UnmarshalFloat64(n, buf)

	if err != nil {
		t.Fatal(err.Error())
	}
	if randomFloat64 != 1231.5131 {
		t.Fatal("randomFloat64: float64 doesn't match")
	}

	var helloWorld string
	_, helloWorld, err = bunsafe.UnmarshalString(0, buf)

	if err != nil {
		t.Fatal(err.Error())
	}
	if helloWorld != "Hello World!" {
		t.Fatal("helloWorld: string doesn't match")
	}

	var randomFloat32 float32
	n, randomFloat32, err = UnmarshalFloat32(n, buf)

	if err != nil {
		t.Fatal(err.Error())
	}
	if randomFloat32 != 1231.5132 {
		t.Fatal("randomFloat32: float32 doesn't match")
	}

	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}
