package benc

import "testing"

func TestBufPool(t *testing.T) {
	bufPool := NewBufPool()
	_, err := bufPool.Marshal(1024, func(b []byte) (n int) {
		return
	})

	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestBufPoolWithSize(t *testing.T) {
	bufPool := NewBufPool(WithBufferSize(1025))
	if bufPool.BufSize != 1025 {
		t.Fatal("size doesn't match!")
	}
	_, err := bufPool.Marshal(1024, func(b []byte) (n int) {
		return
	})

	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestBufPoolError(t *testing.T) {
	bufPool := NewBufPool(WithBufferSize(0))

	// reuse buffer too small for size of `1`
	_, err := bufPool.Marshal(1, func(b []byte) (n int) {
		return
	})

	if err != ErrReuseBufTooSmall {
		t.Fatal("expected a benc.ErrReuseBufTooSmall error!")
	}
}

func TestVerifyMarshalAndUnmarshal(t *testing.T) {
	if err := VerifyMarshal(3, []byte{1, 2, 3}); err != nil {
		t.Fatal("benc.VerifyMarshal error: " + err.Error())
	}

	if err := VerifyUnmarshal(3, []byte{1, 2, 3}); err != nil {
		t.Fatal("benc.VerifyUnmarshal error: " + err.Error())
	}

	if err := VerifyMarshal(2, []byte{1, 2, 3}); err == nil {
		t.Fatal("(benc.VerifyMarshal) expected an error")
	}

	if err := VerifyUnmarshal(2, []byte{1, 2, 3}); err == nil {
		t.Fatal("(benc.VerifyUnmarshal) expected an error")
	}
}
