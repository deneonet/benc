package bidv

import (
	"strings"
	"testing"

	"github.com/deneonet/benc"
	bstd "github.com/deneonet/benc/std"
)

type Test struct {
	err string
}

const TestId = AllowedStartId

func (t *Test) Size() int {
	return bstd.SizeString(t.err)
}

func (t *Test) Marshal(n int, b []byte) int {
	return bstd.MarshalString(n, b, t.err)
}

func (t *Test) Unmarshal(tn int, b []byte) (n int, err error) {
	n, t.err, err = bstd.UnmarshalString(tn, b)
	return
}

func TestBasic(t *testing.T) {
	var id uint = 10
	str := "Hello World!"
	s := Size(id, bstd.SizeString(str))
	b := make([]byte, s)
	n := Marshal(0, b, id)
	bstd.MarshalString(n, b, str)

	n, err := Skip(0, b, id, bstd.SkipString)
	if err != nil {
		t.Fatal(err)
	}
	if n != s {
		t.Fatal("skip: unexpected n")
	}

	n, deserStr, err := Unmarshal[string](0, b, id, bstd.UnmarshalString)
	if err != nil {
		t.Fatal(err)
	}
	if deserStr != str {
		t.Log(deserStr)
		t.Fatal("no match")
	}
	if n != s {
		t.Fatal("unmarshal: unexpected n")
	}
}

func TestStruct(t *testing.T) {
	data := Test{
		err: "None",
	}

	s := Size(TestId, data.Size())
	b := make([]byte, s)
	n := Marshal(0, b, TestId)
	data.Marshal(n, b)

	n, deserData, err := Unmarshal[Test](0, b, TestId, func(n int, b []byte, test *Test) (int, error) {
		return test.Unmarshal(n, b)
	})
	if err != nil {
		t.Fatal(err)
	}
	if deserData.err != data.err {
		t.Log(deserData.err)
		t.Fatal("no match")
	}
	if n != s {
		t.Fatal("unmarshal: unexpected n")
	}
}

func TestErrBufTooSmall(t *testing.T) {
	var id uint = 10
	_, err := Skip(0, []byte{}, id, bstd.SkipString)
	if err != benc.ErrBufTooSmall {
		t.Fatal(err)
	}

	_, _, err = Unmarshal[string](0, []byte{}, id, bstd.UnmarshalString)
	if err != benc.ErrBufTooSmall {
		t.Fatal(err)
	}
}

func TestErrIdMismatch(t *testing.T) {
	var id uint = 10
	str := "Hello World!"
	s := Size(id, bstd.SizeString(str))
	b := make([]byte, s)
	n := Marshal(0, b, id)
	bstd.MarshalString(n, b, str)

	_, err := Skip(0, b, 9, bstd.SkipString)
	if err == nil {
		t.Fatal("skip: expected error")
	}
	if !strings.HasPrefix(err.Error(), "id mismatch:") {
		t.Fatal("skip: expected ID mismatch error")
	}

	_, _, err = Unmarshal[string](0, b, 9, bstd.UnmarshalString)
	if err == nil {
		t.Fatal("unmarshal: expected error")
	}
	if !strings.HasPrefix(err.Error(), "id mismatch:") {
		t.Fatal("unmarshal: expected ID mismatch error")
	}
}

func TestDefaultIdNicknames(t *testing.T) {
	_ = GetDefaultIdNickname(2)
	_ = GetDefaultIdNickname(3)
	_ = GetDefaultIdNickname(4)
	_ = GetDefaultIdNickname(5)
	_ = GetDefaultIdNickname(6)
	_ = GetDefaultIdNickname(7)
	_ = GetDefaultIdNickname(8)
	_ = GetDefaultIdNickname(9)
	_ = GetDefaultIdNickname(10)
	_ = GetDefaultIdNickname(11)
	_ = GetDefaultIdNickname(12)
	_ = GetDefaultIdNickname(13)
	_ = GetDefaultIdNickname(14)
	_ = GetDefaultIdNickname(15)
	_ = GetDefaultIdNickname(16)
}
