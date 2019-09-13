package smms

import (
	"io"
	"sync"
)

// 抄自 github.com/VictoriaMetrics/VictoriaMetrics

var (
	// Verify byteBuffer implements the given interfaces.
	_ io.Writer = &byteBuffer{}
)

// byteBuffer implements a simple byte buffer.
type byteBuffer struct {
	// B is the underlying byte slice.
	B []byte
}

// Reset resets bb.
func (bb *byteBuffer) Reset() {
	bb.B = bb.B[:0]
}

// Write appends p to bb.
func (bb *byteBuffer) Write(p []byte) (int, error) {
	bb.B = append(bb.B, p...)
	return len(p), nil
}

// NewReader returns new reader for the given bb.
func (bb *byteBuffer) NewReader() io.Reader {
	return &reader{
		bb: bb,
	}
}

type reader struct {
	bb *byteBuffer

	// readOffset is the offset in bb.B for read.
	readOffset int
}

// Read reads up to len(p) bytes from bb.
func (r *reader) Read(p []byte) (int, error) {
	var err error
	n := copy(p, r.bb.B[r.readOffset:])
	if n < len(p) {
		err = io.EOF
	}
	r.readOffset += n
	return n, err
}

// MustClose closes bb for subsequent re-use.
func (r *reader) MustClose() {
	r.bb = nil
	r.readOffset = 0
}

// byteBufferPool is a pool of ByteBuffers.
type byteBufferPool struct {
	p sync.Pool
}

// Get obtains a byteBuffer from bbp.
func (bbp *byteBufferPool) Get() *byteBuffer {
	bbv := bbp.p.Get()
	if bbv == nil {
		return &byteBuffer{}
	}
	return bbv.(*byteBuffer)
}

// Put puts bb into bbp.
func (bbp *byteBufferPool) Put(bb *byteBuffer) {
	bb.Reset()
	bbp.p.Put(bb)
}
