package bgenimpl

import (
	"encoding/binary"
	"errors"
	"strconv"
	"testing"

	"github.com/deneonet/benc"
	bstd "github.com/deneonet/benc/std"
)

func TestTags(t *testing.T) {
	buf := []byte{0, 0, 0}
	MarshalTag(0, buf, 2, 2)
	if buf[0] != 2 && buf[1] != 2 {
		t.Fatal("no match")
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
		t.Fatal("expected n of 2")
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
		t.Fatal("expected ErrBufTooSmall")
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
		t.Fatal("expected n of 3")
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
		t.Fatal("expected benc.ErrBufTooSmall")
	}

	_, _, _, err = UnmarshalTag(0, ts)
	if err != benc.ErrBufTooSmall {
		t.Fatal("2: expected benc.ErrBufTooSmall")
	}
}

func TestHandleCompatibility_Basic(t *testing.T) {
	_, _, err := HandleCompatibility(0, []byte{}, []uint16{}, 0)
	if err != ErrEof {
		t.Fatal("expected ErrEof")
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
		t.Fatal("expected `ok`")
	}
	if n != 2 {
		t.Fatal("expected n of 2")
	}

	n, ok, err = HandleCompatibility(0, buf, []uint16{}, 9)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("unexpected `ok`")
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
		t.Fatal("unexpected `ok`")
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
		t.Fatal("unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("expected ErrEof")
	}

	MarshalTag(0, buf, Fixed16, 1)
	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("expected ErrEof")
	}

	MarshalTag(0, buf, Fixed32, 1)
	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("expected ErrEof")
	}

	MarshalTag(0, buf, Fixed64, 1)
	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("expected ErrEof")
	}

	MarshalTag(0, buf, Varint, 1)
	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("unexpected `ok`")
	}
	if err != benc.ErrBufTooSmall {
		t.Fatal("expected ErrBufTooSmall")
	}

	buf = make([]byte, 2+2)
	MarshalTag(0, buf, Container, 1)
	buf[2] = 1
	buf[3] = 1

	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("expected ErrEof")
	}

	buf = make([]byte, 2+2+1)
	MarshalTag(0, buf, Container, 1)
	buf[2] = 0
	buf[3] = 1
	buf[4] = 1

	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("unexpected `ok`")
	}
	if err != ErrInvalidType {
		t.Fatal("expected ErrInvalidType")
	}

	buf = make([]byte, 2+1)
	MarshalTag(0, buf, Container, 1)
	buf[2] = 0

	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("unexpected `ok`")
	}
	if err != benc.ErrBufTooSmall {
		t.Fatal("expected benc.ErrBufTooSmall")
	}

	buf = make([]byte, 2+2+2+1)
	MarshalTag(0, buf, Container, 1)
	MarshalTag(2, buf, Fixed8, 2)
	buf[5] = 1
	buf[6] = 1

	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("expected ErrEof")
	}

	buf = make([]byte, 2+1+2)
	MarshalTag(0, buf, Bytes, 1)
	bstd.MarshalBytes(2, buf, []byte{1, 2})

	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("expected ErrEof")
	}

	buf = make([]byte, 2+4+1+2)
	MarshalTag(0, buf, ArrayMap, 1)
	bstd.MarshalSlice(2, buf, []byte{1, 2}, bstd.MarshalByte)

	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("unexpected `ok`")
	}
	if err != ErrEof {
		t.Fatal("expected ErrEof")
	}

	buf = make([]byte, 2+2+1)
	MarshalTag(0, buf, Fixed8, 1)
	MarshalTag(3, buf, Fixed16, 2)

	n, ok, _ := HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("unexpected `ok`")
	}
	if n != 3 {
		t.Fatal("expected n of 3")
	}

	MarshalTag(0, buf, 0, 1)
	_, ok, err = HandleCompatibility(0, buf, []uint16{1}, 0)
	if ok {
		t.Fatal("unexpected `ok`")
	}
	if err != ErrInvalidType {
		t.Fatal("expected ErrInvalidType")
	}
}

var maxVarintLenMap = map[int]int{
	64: binary.MaxVarintLen64,
	32: binary.MaxVarintLen32,
}

var maxVarintLen = maxVarintLenMap[strconv.IntSize]

func TestEnums(t *testing.T) {
	v := 150
	size := SizeEnum(v)
	if size != 2 {
		t.Errorf("expected size 2, got %d", size)
	}

	buf := []byte{172, 2}
	n, err := SkipEnum(0, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("expected offset 2, got %d", n)
	}

	buf = []byte{172}
	_, err = SkipEnum(0, buf)
	if !errors.Is(err, benc.ErrBufTooSmall) {
		t.Errorf("expected ErrBufTooSmall, got %v", err)
	}

	buf = make([]byte, 10)
	n = MarshalEnum(0, buf, 150)
	if n != 2 {
		t.Errorf("expected offset 2, got %d", n)
	}
	t.Log(buf[1])
	if buf[0] != 172 || buf[1] != 2 {
		t.Errorf("unexpected buffer contents: %v", buf[:n])
	}

	buf = []byte{172, 2}
	n, v, err = UnmarshalEnum[int](0, buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("expected offset 2, got %d", n)
	}
	if v != 150 {
		t.Errorf("expected value 150, got %d", v)
	}

	buf = []byte{172}
	_, _, err = UnmarshalEnum[int](0, buf)
	if !errors.Is(err, benc.ErrBufTooSmall) {
		t.Errorf("expected ErrBufTooSmall, got %v", err)
	}

	buf = make([]byte, maxVarintLen+1)
	for i := 0; i < maxVarintLen+1; i++ {
		buf[i] = 0x80
	}
	buf[maxVarintLen] = 0x02
	_, _, err = UnmarshalEnum[int](0, buf)
	if !errors.Is(err, benc.ErrOverflow) {
		t.Errorf("expected ErrOverflow, got %v", err)
	}
}
