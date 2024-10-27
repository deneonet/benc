package benc

import (
	"errors"
	"sync"
)

var ErrBufTooSmall = errors.New("buffer too small")
var ErrReuseBufTooSmall = errors.New("reuse buffer too small")
var ErrOverflow = errors.New("varint overflows a 64-bit integer")
var ErrVerifyUnmarshal = errors.New("check for a mistake in the unmarshal process")
var ErrVerifyMarshal = errors.New("check for a mistake in calculating the size or in the marshal process")

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

// Verifies that the length of the buffer equals n
func VerifyMarshal(n int, b []byte) error {
	if n != len(b) {
		return ErrVerifyMarshal
	}
	return nil
}

// Verifies that the length of the buffer equals n
func VerifyUnmarshal(n int, b []byte) error {
	if n != len(b) {
		return ErrVerifyUnmarshal
	}
	return nil
}
