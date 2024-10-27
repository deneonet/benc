package bgenimpl

import (
	"testing"

	"github.com/deneonet/benc"
	bstd "github.com/deneonet/benc/std"
)

func TestTags(t *testing.T) {
	buf := []byte{0, 0, 0}
	MarshalTag(0, buf, 2, 2)
	if buf[0] != 2 && buf[1] != 2 {
		t.Fatal("1: no match")
	}

	_, id, ty, err := UnmarshalTag(0, buf)
	if err != nil {
		t.Fatal(err)
	}
	if ty != 2 && id != 2 {
		t.Fatalf("2: no match %d %d", ty, id)
	}

	MarshalTag(0, buf, 50, 255)
	if buf[0] != 50 && buf[1] != 255 {
		t.Fatal("3: no match")
	}

	n, err := SkipTag(0, buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatal("1: expected n of 2")
	}

	_, id, ty, err = UnmarshalTag(0, buf)
	if err != nil {
		t.Fatal(err)
	}
	if ty != 50 && id != 255 {
		t.Fatalf("4: no match %d %d", ty, id)
	}

	_, _, _, err = UnmarshalTag(0, []byte{})
	if err != benc.ErrBufTooSmall {
		t.Fatal("1: expected ErrBufTooSmall")
	}

	_, _, _, err = UnmarshalTag(0, []byte{1})
	if err != benc.ErrBufTooSmall {
		t.Fatal("2: expected ErrBufTooSmall")
	}

	_, err = SkipTag(0, []byte{})
	if err != benc.ErrBufTooSmall {
		t.Fatal("3: expected ErrBufTooSmall")
	}

	_, err = SkipTag(0, []byte{1})
	if err != benc.ErrBufTooSmall {
		t.Fatal("4: expected ErrBufTooSmall")
	}

	MarshalTag(0, buf, 50, 256)
	if buf[0] != 50 && buf[1] != 1 && buf[2] != 0 {
		t.Fatal("5: no match")
	}

	n, err = SkipTag(0, buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatal("1: expected n of 3")
	}

	_, id, ty, err = UnmarshalTag(0, buf)
	if err != nil {
		t.Fatal(err)
	}
	if ty != 50 && id != 256 {
		t.Fatalf("6: no match %d %d", ty, id)
	}

	MarshalTag(0, buf, 50, 256)
	if buf[0] != 50 && buf[1] != 1 && buf[2] != 0 {
		t.Fatal("7: no match")
	}
	ts := make([]byte, 2)
	ts[0] = buf[0]
	ts[1] = buf[1]

	_, err = SkipTag(0, ts)
	if err != benc.ErrBufTooSmall {
		t.Fatal("1: expected benc.ErrBufTooSmall")
	}

	_, _, _, err = UnmarshalTag(0, ts)
	if err != benc.ErrBufTooSmall {
		t.Fatal("2: expected benc.ErrBufTooSmall")
	}
}

func TestHandleCompatibility_Basic(t *testing.T) {
	_, _, err := HandleCompatibility(0, []byte{}, []uint16{}, 0)
	if err != ErrEof {
		t.Fatal("1: expected ErrEof")
	}

	_, _, err = HandleCompatibility(0, []byte{1}, []uint16{}, 0)
	if err != ErrEof {
		t.Fatal("2: expected ErrEof")
	}

	buf := make([]byte, 1024)
	MarshalTag(0, buf, 2, 10)

	n, ok, err := HandleCompatibility(0, buf, []uint16{}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("1: expected `ok`")
	}
	if n != 2 {
		t.Fatal("1: expected n of 2")
	}

	n, ok, err = HandleCompatibility(0, buf, []uint16{}, 9)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if n != 0 {
		t.Fatal("2: expected n of 0")
	}

	MarshalTag(0, buf, 2, 256)
	n, ok, err = HandleCompatibility(0, buf, []uint16{}, 255)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if n != 0 {
		t.Fatal("2: expected n of 0")
	}
}

func TestHandleCompatibility_Types(t *testing.T) {
	buf := make([]byte, 2)
	MarshalTag(0, buf, Fixed8, 1)

	_, ok, err := HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("1: expected ErrEof")
	}

	MarshalTag(0, buf, Fixed16, 1)
	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("1: expected ErrEof")
	}

	MarshalTag(0, buf, Fixed32, 1)
	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("1: expected ErrEof")
	}

	MarshalTag(0, buf, Fixed64, 1)
	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("1: expected ErrEof")
	}

	buf = make([]byte, 2+2)
	MarshalTag(0, buf, Container, 1)
	buf[2] = 1
	buf[3] = 1

	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("1: expected ErrEof")
	}

	buf = make([]byte, 2+2+1)
	MarshalTag(0, buf, Container, 1)
	buf[2] = 0
	buf[3] = 1
	buf[4] = 1

	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if err != ErrInvalidType {
		t.Fatal("1: expected ErrInvalidType")
	}

	buf = make([]byte, 2+1)
	MarshalTag(0, buf, Container, 1)
	buf[2] = 0

	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if err != benc.ErrBufTooSmall {
		t.Fatal("1: expected benc.ErrBufTooSmall")
	}

	buf = make([]byte, 2+2+2+1)
	MarshalTag(0, buf, Container, 1)
	MarshalTag(2, buf, Fixed8, 2)
	buf[5] = 1
	buf[6] = 1

	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("1: expected ErrEof")
	}

	buf = make([]byte, 2+1+2)
	MarshalTag(0, buf, Bytes, 1)
	bstd.MarshalBytes(2, buf, []byte{1, 2})

	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("1: expected ErrEof")
	}

	buf = make([]byte, 2+4+1+2)
	MarshalTag(0, buf, ArrayMap, 1)
	bstd.MarshalSlice(2, buf, []byte{1, 2}, bstd.MarshalByte)

	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("1: expected ErrEof")
	}

	buf = make([]byte, 2+2+1)
	MarshalTag(0, buf, Fixed8, 1)
	MarshalTag(3, buf, Fixed16, 2)

	n, ok, _ := HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if n != 3 {
		t.Fatal("1: expected n of 3")
	}

	MarshalTag(0, buf, 0, 1)
	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("1: unexpected `ok`")
	}
	if err != ErrInvalidType {
		t.Fatal("1: expected ErrInvalidType")
	}
}
