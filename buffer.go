package jsonex

import (
	"sync"
)

// buffer represents an internal byte buffer (unexported)
type buffer struct {
	data []byte
	cap  int
	pos  int
}

// newBuffer creates a new buffer with the specified capacity
func newBuffer(capacity int) *buffer {
	return &buffer{
		data: make([]byte, 0, capacity),
		cap:  capacity,
		pos:  0,
	}
}

// grow increases the buffer capacity if needed
func (b *buffer) grow(n int) {
	if len(b.data)+n <= cap(b.data) {
		return
	}
	newCap := cap(b.data) * 2
	if newCap < len(b.data)+n {
		newCap = len(b.data) + n
	}
	newData := make([]byte, len(b.data), newCap)
	copy(newData, b.data)
	b.data = newData
}

// write appends data to the buffer
func (b *buffer) write(data []byte) {
	b.grow(len(data))
	b.data = append(b.data, data...)
}

// writeByte appends a single byte to the buffer
func (b *buffer) writeByte(c byte) {
	b.grow(1)
	b.data = append(b.data, c)
}

// bytes returns the current buffer contents
func (b *buffer) bytes() []byte {
	return b.data
}

// len returns the current buffer length
func (b *buffer) len() int {
	return len(b.data)
}

// reset clears the buffer for reuse
func (b *buffer) reset() {
	b.data = b.data[:0]
	b.pos = 0
}

// slice returns a slice of the buffer from start to end
func (b *buffer) slice(start, end int) []byte {
	if start < 0 || end > len(b.data) || start > end {
		return nil
	}
	return b.data[start:end]
}

// bufferPool provides pooled buffers for memory efficiency
var bufferPool = sync.Pool{
	New: func() interface{} {
		return newBuffer(4096) // default buffer size
	},
}

// getBuffer gets a buffer from the pool
func getBuffer() *buffer {
	return bufferPool.Get().(*buffer)
}

// putBuffer returns a buffer to the pool
func putBuffer(b *buffer) {
	b.reset()
	bufferPool.Put(b)
}