package bstd

import "testing"

// Mutate complete and malformed wire values across independent decoder paths.
// Every successful decoder must stay within the input; failures must not
// report a committed cursor. Fuzzing also catches panics and runaway sizes.
func FuzzWireDecode(f *testing.F) {
	for _, data := range [][]byte{
		{}, {0}, {1}, {1, 'a'}, {1, 0xff},
		{2, 1, 2, 1, 1, 1, 1},
		{1, 7, 8, 1, 1, 1, 1},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 2},
	} {
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			t.Skip()
		}
		check := func(name string, n int, err error) {
			t.Helper()
			if err != nil && n != 0 || err == nil && (n < 0 || n > len(data)) {
				t.Fatalf("%s: cursor=%d input=%d err=%v", name, n, len(data), err)
			}
		}
		n, _, err := UnmarshalString(0, data)
		check("string", n, err)
		n, err = SkipString(0, data)
		check("skip string", n, err)
		n, _, err = UnmarshalBytesCopied(0, data)
		check("bytes", n, err)
		n, _, err = UnmarshalUint(0, data)
		check("varint", n, err)
		n, _, err = UnmarshalSlice[byte](0, data, UnmarshalByte)
		check("slice", n, err)
		n, _, err = UnmarshalMap[byte, byte](0, data, UnmarshalByte, UnmarshalByte)
		check("map", n, err)
	})
}
