package bstd

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestIndependentWireV1Vectors(t *testing.T) {
	var fixture struct {
		Valid   []struct{ Name, Hex string }
		Invalid []struct{ Name, Hex string }
	}
	data, err := os.ReadFile("../../testdata/conformance/wire-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, v := range fixture.Valid {
		t.Run(v.Name, func(t *testing.T) {
			if strconv.IntSize != 64 && (v.Name == "uint64_max_varint_64bit" || v.Name == "int64_min_varint_64bit" || v.Name == "int64_max_varint_64bit") {
				t.Skip("architecture-dependent varint")
			}
			want, err := hex.DecodeString(v.Hex)
			if err != nil {
				t.Fatal(err)
			}
			buf := make([]byte, len(want))
			var n int
			switch v.Name {
			case "uint_zero":
				n = MarshalUint(0, buf, 0)
			case "uint_127":
				n = MarshalUint(0, buf, 127)
			case "uint_128":
				n = MarshalUint(0, buf, 128)
			case "uint32_max_varint":
				n = MarshalUint(0, buf, 0xffffffff)
			case "uint64_max_varint_64bit":
				n = MarshalUint(0, buf, ^uint(0))
			case "uint64_max_fixed":
				n = MarshalUint64(0, buf, ^uint64(0))
			case "int_minus_one":
				n = MarshalInt(0, buf, -1)
			case "int64_min_varint_64bit":
				x := int64(-1 << 63)
				n = MarshalInt(0, buf, int(x))
			case "int64_max_varint_64bit":
				x := int64(1<<63 - 1)
				n = MarshalInt(0, buf, int(x))
			case "int64_min_fixed":
				n = MarshalInt64(0, buf, -1<<63)
			case "int64_max_fixed":
				n = MarshalInt64(0, buf, 1<<63-1)
			case "string_nul":
				n = MarshalString(0, buf, "A\x00")
			case "string_utf8":
				n = MarshalString(0, buf, "é")
			case "bytes":
				n = MarshalBytes(0, buf, []byte{0, 255})
			case "array":
				n = MarshalArray(0, buf, []int32{1, -2}, MarshalInt32)
			case "slice":
				n = MarshalSlice(0, buf, []int32{1, -2}, MarshalInt32)
			case "empty_slice":
				n = MarshalSlice(0, buf, []int32{}, MarshalInt32)
			case "map_single":
				n = MarshalMap(0, buf, map[string]int32{"a": 2}, MarshalString, MarshalInt32)
			case "nested":
				n = MarshalBool(MarshalInt32(0, buf, 7), buf, true)
			case "nil_pointer":
				n = MarshalPointer(0, buf, (*int32)(nil), MarshalInt32)
			case "present_pointer":
				x := int32(7)
				n = MarshalPointer(0, buf, &x, MarshalInt32)
			case "unix_nanosecond":
				n = MarshalTime(0, buf, time.Unix(0, 1))
			case "negative_nanosecond":
				n = MarshalTime(0, buf, time.Unix(0, -1))
			default:
				t.Fatalf("unhandled vector: %s", v.Name)
			}
			if n != len(want) || hex.EncodeToString(buf) != v.Hex {
				t.Fatalf("got %x at %d, want %x", buf, n, want)
			}
		})
	}
	for _, v := range fixture.Invalid {
		t.Run(v.Name, func(t *testing.T) {
			b, err := hex.DecodeString(v.Hex)
			if err != nil {
				t.Fatal(err)
			}
			var n int
			switch v.Name {
			case "missing_slice_terminator", "wrong_slice_terminator", "huge_slice":
				n, err = SkipSlice(0, b, SkipByte)
			case "overflowed_varint":
				n, _, err = UnmarshalUint(0, b)
			case "truncated_string", "invalid_utf8", "huge_string":
				n, _, err = UnmarshalString(0, b)
			case "truncated_map", "wrong_map_terminator":
				n, err = SkipMap(0, b, SkipByte, SkipByte)
			case "invalid_bool":
				n, err = SkipBool(0, b)
			case "duplicate_map_key":
				n, _, err = UnmarshalMap[byte, byte](0, b, UnmarshalByte, UnmarshalByte)
			default:
				t.Fatalf("unhandled invalid vector: %s", v.Name)
			}
			if err == nil || n != 0 {
				t.Fatalf("got (%d, %v), want error and zero cursor", n, err)
			}
			want := map[string]error{
				"missing_slice_terminator": ErrBufTooSmall, "wrong_slice_terminator": ErrInvalidTerminator,
				"overflowed_varint": ErrOverflow, "truncated_string": ErrBufTooSmall,
				"truncated_map": ErrBufTooSmall, "invalid_bool": ErrInvalidBool,
				"invalid_utf8": ErrInvalidUTF8, "wrong_map_terminator": ErrInvalidTerminator,
				"duplicate_map_key": ErrDuplicateKey, "huge_string": ErrDecodeLimit,
				"huge_slice": ErrDecodeLimit,
			}[v.Name]
			if !errors.Is(err, want) && !(strconv.IntSize == 32 && errors.Is(err, ErrOverflow) &&
				(v.Name == "huge_string" || v.Name == "huge_slice")) {
				t.Fatalf("got %v, want %v", err, want)
			}
		})
	}
}

func TestMalformedDecodeHasBoundedErrors(t *testing.T) {
	// The fifth byte 0f is legal on 32-bit machines (uint32 max).
	uint32MaxVarint := []byte{255, 255, 255, 255, 15}
	if n, got, err := UnmarshalUint(0, uint32MaxVarint); err != nil || n != 5 || got != 0xffffffff {
		t.Fatalf("uint32 varint: n=%d value=%d err=%v", n, got, err)
	}
	if n, err := SkipVarint(0, uint32MaxVarint); err != nil || n != 5 {
		t.Fatalf("skip uint32 varint: n=%d err=%v", n, err)
	}
	maxLength := []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 1}
	for _, offset := range []int{-1, 2} {
		if n, _, err := UnmarshalString(offset, []byte{0}); n != 0 || !errors.Is(err, ErrBufTooSmall) {
			t.Fatalf("string offset %d: %d, %v", offset, n, err)
		}
		if n, err := SkipByte(offset, []byte{0}); n != 0 || !errors.Is(err, ErrBufTooSmall) {
			t.Fatalf("byte offset %d: %d, %v", offset, n, err)
		}
		if n, _, err := UnmarshalUint(offset, []byte{0}); n != 0 || !errors.Is(err, ErrBufTooSmall) {
			t.Fatalf("varint offset %d: %d, %v", offset, n, err)
		}
	}
	if n, _, err := UnmarshalString(0, maxLength); n != 0 || !(errors.Is(err, ErrDecodeLimit) || errors.Is(err, ErrOverflow)) {
		t.Fatalf("huge string: %d, %v", n, err)
	}
	if n, _, err := UnmarshalSlice[byte](0, maxLength, UnmarshalByte); n != 0 || !(errors.Is(err, ErrDecodeLimit) || errors.Is(err, ErrOverflow)) {
		t.Fatalf("huge slice: %d, %v", n, err)
	}
	if n, _, err := UnmarshalMap[byte, byte](0, maxLength, UnmarshalByte, UnmarshalByte); n != 0 || !(errors.Is(err, ErrDecodeLimit) || errors.Is(err, ErrOverflow)) {
		t.Fatalf("huge map: %d, %v", n, err)
	}
	for _, b := range [][]byte{{0}, {0, 0, 0, 0, 0}} {
		if n, _, err := UnmarshalSlice[byte](0, b, UnmarshalByte); n != 0 || err == nil {
			t.Fatalf("slice %x: %d, %v", b, n, err)
		}
		if n, _, err := UnmarshalMap[byte, byte](0, b, UnmarshalByte, UnmarshalByte); n != 0 || err == nil {
			t.Fatalf("map %x: %d, %v", b, n, err)
		}
	}
	if _, err := SkipBool(0, []byte{2}); !errors.Is(err, ErrInvalidBool) {
		t.Fatal(err)
	}
	if n, _, err := UnmarshalSlice[byte](0, []byte{1, 7, 1, 1, 1, 1}, func(int, []byte) (int, byte, error) { return 100, 0, nil }); n != 0 || !errors.Is(err, ErrVerifyUnmarshal) {
		t.Fatalf("invalid callback cursor: %d, %v", n, err)
	}
}

func TestStrictUTF8AndTransactionalStringWrite(t *testing.T) {
	invalid := []byte{1, 0xff}
	if n, _, err := UnmarshalString(0, invalid); n != 0 || !errors.Is(err, ErrInvalidUTF8) {
		t.Fatalf("invalid UTF-8: offset=%d error=%v", n, err)
	}
	if n, err := SkipString(0, invalid); n != 0 || !errors.Is(err, ErrInvalidUTF8) {
		t.Fatalf("skip UTF-8: offset=%d error=%v", n, err)
	}
	buf := []byte{0xaa}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected short write panic")
			}
		}()
		MarshalString(0, buf, "too long")
	}()
	if buf[0] != 0xaa {
		t.Fatalf("partial write: %x", buf)
	}
}

func TestBoundedMalformedCorpus(t *testing.T) {
	seed := uint32(1)
	for size := 0; size <= 16; size++ {
		for i := 0; i < 128; i++ {
			b := make([]byte, size)
			for j := range b {
				seed = seed*1664525 + 1013904223
				b[j] = byte(seed >> 24)
			}
			for _, decode := range []func() (int, error){
				func() (int, error) { return SkipSlice(0, b, SkipByte) },
				func() (int, error) { n, _, err := UnmarshalSlice[byte](0, b, UnmarshalByte); return n, err },
				func() (int, error) { return SkipMap(0, b, SkipByte, SkipByte) },
				func() (int, error) {
					n, _, err := UnmarshalMap[byte, byte](0, b, UnmarshalByte, UnmarshalByte)
					return n, err
				},
				func() (int, error) { n, _, err := UnmarshalString(0, b); return n, err },
				func() (int, error) { n, _, err := UnmarshalBytesCopied(0, b); return n, err },
			} {
				n, err := decode()
				if (err != nil && n != 0) || n < 0 || n > len(b) {
					t.Fatalf("malformed input %x: offset=%d err=%v", b, n, err)
				}
			}
		}
	}
}

func TestDecodeDepthBudget(t *testing.T) {
	budget := NewDecodeDepth(16)
	var recursive func(int, []byte, *DecodeDepth) (int, error)
	recursive = func(n int, b []byte, d *DecodeDepth) (int, error) {
		return SkipPointerWithDepth(n, b, d, recursive)
	}
	deep := make([]byte, 18)
	for i := range deep {
		deep[i] = 1
	}
	if n, err := SkipPointerWithDepth(0, deep, budget, recursive); n != 0 || !errors.Is(err, ErrDecodeLimit) {
		t.Fatalf("deep pointer: offset=%d error=%v", n, err)
	}
	// An error must release the budget for an independent following message.
	shallow := []byte{1, 0}
	if n, err := SkipPointerWithDepth(0, shallow, budget, recursive); n != len(shallow) || err != nil {
		t.Fatalf("shallow pointer after error: offset=%d error=%v", n, err)
	}
	var read func(int, []byte, *DecodeDepth) (int, byte, error)
	read = func(n int, b []byte, d *DecodeDepth) (int, byte, error) {
		n, _, err := UnmarshalPointerWithDepth(n, b, d, read)
		return n, 1, err
	}
	if n, _, err := UnmarshalPointerWithDepth(0, deep, budget, read); n != 0 || !errors.Is(err, ErrDecodeLimit) {
		t.Fatalf("deep unmarshal: offset=%d error=%v", n, err)
	}
}
