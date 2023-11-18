package bstd

import (
	"bytes"
	"testing"
	"time"

	"github.com/deneonet/benc/bpre"
	"github.com/deneonet/benc/btag"
	"github.com/deneonet/benc/bunsafe"
)

func TestDataTypes(t *testing.T) {
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
	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
	n, tag, err := UnmarshalStringTag(0, buf)
	checkErr(t, err)
	if tag != "v1" {
		t.Fatal("tag doesn't match")
	}
	n, _, err = UnmarshalBool(n, buf)
	checkErr(t, err)
	n, _, err = UnmarshalByte(n, buf)
	checkErr(t, err)
	n, _, err = UnmarshalFloat32(n, buf)
	checkErr(t, err)
	n, _, err = UnmarshalFloat64(n, buf)
	checkErr(t, err)
	n, _, err = UnmarshalInt(n, buf)
	checkErr(t, err)
	n, _, err = UnmarshalInt16(n, buf)
	checkErr(t, err)
	n, _, err = UnmarshalInt32(n, buf)
	checkErr(t, err)
	n, _, err = UnmarshalInt64(n, buf)
	checkErr(t, err)
	n, _, err = bunsafe.UnmarshalString(n, buf)
	checkErr(t, err)
	n, _, err = UnmarshalTime(n, buf)
	checkErr(t, err)
	n, _, err = UnmarshalUInt(n, buf)
	checkErr(t, err)
	n, _, err = UnmarshalUInt16(n, buf)
	checkErr(t, err)
	n, _, err = UnmarshalUInt32(n, buf)
	checkErr(t, err)
	n, _, err = UnmarshalUInt64(n, buf)
	checkErr(t, err)
	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
}

func TestSliceMap(t *testing.T) {
	data := []string{"WWWWWWWWWWWWWWWWWWWWWWWWWWWW!", "hhhhhhhhhhhhhhhhhhhhhhhhhhhh"}
	m := make(map[string]int)
	m["WWWWWWWWWWWWWWWWWWWWWWWWWWWW"] = 1022323232323232323
	m["hhhhhhhhhhhhhhhhhhhhhhhhhhhh"] = 23232323232323

	s := SizeSlice(data, SizeString)
	s += SizeMap(m, SizeString, SizeInt)
	n, buf := Marshal(s)
	n = MarshalSlice(n, buf, data, MarshalString)
	n = MarshalMap(n, buf, m, MarshalString, MarshalInt)
	if err := VerifyMarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
	n, data2, err := UnmarshalSlice(0, buf, UnmarshalString)
	if err != nil {
		t.Fatal(err.Error())
	}
	n, m, err = UnmarshalMap(n, buf, UnmarshalString, UnmarshalInt)
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := VerifyUnmarshal(n, buf); err != nil {
		t.Fatal(err.Error())
	}
	if data2[0] != "WWWWWWWWWWWWWWWWWWWWWWWWWWWW!" || data2[1] != "hhhhhhhhhhhhhhhhhhhhhhhhhhhh" {
		t.Fatal("slice doesn't match")
	}
	if m["hhhhhhhhhhhhhhhhhhhhhhhhhhhh"] != 23232323232323 || m["WWWWWWWWWWWWWWWWWWWWWWWWWWWW"] != 1022323232323232323 {
		t.Fatal("map doesn't match")
	}
}

func TestMessageFraming(t *testing.T) {
	var bytes bytes.Buffer
	n, buf := MarshalMF(7)
	n = bunsafe.MarshalString(n, buf, "Hello")
	bytes.Write(buf)
	bytes.Write(buf)

	data, _ := UnmarshalMF(bytes.Bytes())
	for _, bs := range data {
		_, d2, _ := bunsafe.UnmarshalString(n, bs)
		if d2 != "Hello" {
			t.Fatal("unmarshal string don't match")
		}
	}

	bpre.UnmarshalMF(100)
	bpre.Marshal(100)

	bytes.Reset()
	n, buf = btag.SMarshalMF(7, "v1")
	n = bunsafe.MarshalString(n, buf, "Hello")
	bytes.Write(buf)
	bytes.Write(buf)

	data, _ = UnmarshalMF(bytes.Bytes())
	for _, bs := range data {
		_, tag, _ := UnmarshalStringTag(0, bs)
		if tag != "v1" {
			t.Fatal("tag don't match")
		}
		_, d2, _ := bunsafe.UnmarshalString(n, bs)
		if d2 != "Hello" {
			t.Fatal("unmarshal string don't match")
		}
	}

	bytes.Reset()
	n, buf = btag.UMarshalMF(7, 1)
	n = bunsafe.MarshalString(n, buf, "Hello")
	bytes.Write(buf)
	bytes.Write(buf)

	data, _ = UnmarshalMF(bytes.Bytes())
	for _, bs := range data {
		_, tag, _ := UnmarshalUIntTag(0, bs)
		if tag != 1 {
			t.Fatal("tag don't match")
		}
		_, d2, _ := bunsafe.UnmarshalString(n, bs)
		if d2 != "Hello" {
			t.Fatal("unmarshal string don't match")
		}
	}
}

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err.Error())
	}
}
