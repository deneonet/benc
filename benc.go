package benc

import (
	"errors"
	"sync"
)

var ErrBufTooSmall = errors.New("buffer too small")
var ErrInvalidData = errors.New("decoded size is too big")
var ErrDataTooBig = errors.New("data too big for max size")
var ErrReuseBufTooSmall = errors.New("reuse buffer too small")
var ErrOverflow = errors.New("varint overflows a 64-bit integer")
var ErrInvalidSize = errors.New("decoded size is invalid, neither 2, 4 nor 8")

const BencVersion = "v1.1.0"

const (
	Bytes2 int = 2
	Bytes4 int = 4
	Bytes8 int = 8
)

type optFunc func(*Opts)

type Opts struct {
	bufSize uint
}

func defaultOpts() Opts {
	return Opts{
		bufSize: 1024,
	}
}

type BufPool struct {
	BufSize uint
	p       sync.Pool
}

func WithBufferSize(bufSize uint) optFunc {
	return func(o *Opts) {
		o.bufSize = bufSize
	}
}

func NewBufPool(opts ...optFunc) *BufPool {
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	bp := &BufPool{
		BufSize: o.bufSize,
		p: sync.Pool{
			New: func() interface{} {
				s := make([]byte, o.bufSize)
				return &s
			},
		},
	}
	return bp
}

// Initialises the marshal process, it reuses the buffers from a buf pool instance
//
// s = size of the data in bytes, retrieved by using the benc `Size...` methods
func (bp *BufPool) Marshal(s int, f func(b []byte) (n int)) ([]byte, error) {
	ptr := bp.p.Get().(*[]byte)
	slice := *ptr

	if s > len(slice) {
		return nil, ErrReuseBufTooSmall
	}

	b := slice[:s]
	f(b)
	*ptr = slice
	bp.p.Put(ptr)

	return b, nil
}

// Initialises the marshal process, it creates a new buffer of the size `s`
//
// s = size of the data in bytes, retrieved by using the benc `Size...` methods
func Marshal(s int) (int, []byte) {
	return 0, make([]byte, s)
}

type Container interface {
	Size(id uint16) (int, error)
	Marshal(n int, b []byte, id uint16) (int, error)
	Unmarshal(n int, b []byte, id uint16) (int, error)
}

// Marshals a benc container, which is generated using bencgen
//
// c = the container
func MarshalCtr(c Container) (b []byte, err error) {
	s, err := c.Size(0)
	if err != nil {
		return nil, err
	}

	b = make([]byte, s)
	if _, err = c.Marshal(0, b, 0); err != nil {
		return
	}
	return
}

// Unmarshals a benc container, which is generated using bencgen
//
// c = the container, b = the bytes to unmarshal
func UnmarshalCtr(b []byte, c Container) (err error) {
	_, err = c.Unmarshal(0, b, 0)
	return
}

// Verifies that the length of the buffer equals n
// TODO: make constant errors
func VerifyMarshal(n int, b []byte) error {
	if n != len(b) {
		return errors.New("check for a mistake in calculating the size or in the marshal process")
	}
	return nil
}

// Verifies that the length of the buffer equals n
func VerifyUnmarshal(n int, b []byte) error {
	if n != len(b) {
		return errors.New("check for a mistake in the unmarshal process")
	}
	return nil
}

//

// unfinished; reason: max size of each byte slice `math.maxUint16` + no buffer reusing support
func UnmarshalMF(b []byte) ([][]byte, error) {
	var n uint16

	var dec [][]byte
	for {
		if 2 > len(b[n:]) {
			return dec, nil
		}
		u := b[n : n+2]
		_ = u[1]
		size := uint16(u[0]) | uint16(u[1])<<8
		n += 2
		if int(size) > len(b[n:]) {
			return nil, ErrBufTooSmall
		}
		dec = append(dec, b[n:n+size])
		n += size
	}

	/* for buffer reusing

	for i := 0; i < len(b); i++ {
		if 2 > len(b[n:]) {
			return dec[:i], nil
		}
		u := b[n : n+2]
		_ = u[1]
		size := uint16(u[0]) | uint16(u[1])<<8
		n += 2
		if int(size) > len(b[n:]) {
			return nil, ErrBytesTooSmall
		}
		dec[i] = b[n : n+size]
		n += size
	}

	return dec[:i], nil*/
}

// unfinished; reason: max size of each byte slice `math.maxUint16` + no buffer reusing support
func MarshalMF(s int) (int, []byte) {
	b := make([]byte, s+2)
	v := uint16(s)
	_ = b[1]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	return 2, b
}
