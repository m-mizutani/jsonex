package jsonex

import (
	"bytes"
	"testing"
)

func TestNewBuffer(t *testing.T) {
	buf := newBuffer(1024)
	
	if buf.cap != 1024 {
		t.Errorf("newBuffer(1024).cap = %d, expected 1024", buf.cap)
	}
	if buf.len() != 0 {
		t.Errorf("newBuffer(1024).len() = %d, expected 0", buf.len())
	}
	if buf.pos != 0 {
		t.Errorf("newBuffer(1024).pos = %d, expected 0", buf.pos)
	}
}

func TestBufferWrite(t *testing.T) {
	buf := newBuffer(10)
	
	data := []byte("hello")
	buf.write(data)
	
	if buf.len() != 5 {
		t.Errorf("buffer.len() after write = %d, expected 5", buf.len())
	}
	if !bytes.Equal(buf.bytes(), data) {
		t.Errorf("buffer.bytes() = %v, expected %v", buf.bytes(), data)
	}
}

func TestBufferWriteByte(t *testing.T) {
	buf := newBuffer(10)
	
	buf.writeByte('A')
	buf.writeByte('B')
	
	if buf.len() != 2 {
		t.Errorf("buffer.len() after writeByte = %d, expected 2", buf.len())
	}
	
	expected := []byte("AB")
	if !bytes.Equal(buf.bytes(), expected) {
		t.Errorf("buffer.bytes() = %v, expected %v", buf.bytes(), expected)
	}
}

func TestBufferGrow(t *testing.T) {
	buf := newBuffer(5)
	
	// Write more data than initial capacity
	data := []byte("hello world")
	buf.write(data)
	
	if buf.len() != len(data) {
		t.Errorf("buffer.len() after grow = %d, expected %d", buf.len(), len(data))
	}
	if !bytes.Equal(buf.bytes(), data) {
		t.Errorf("buffer.bytes() after grow = %v, expected %v", buf.bytes(), data)
	}
}

func TestBufferReset(t *testing.T) {
	buf := newBuffer(10)
	
	buf.write([]byte("hello"))
	buf.reset()
	
	if buf.len() != 0 {
		t.Errorf("buffer.len() after reset = %d, expected 0", buf.len())
	}
	if buf.pos != 0 {
		t.Errorf("buffer.pos after reset = %d, expected 0", buf.pos)
	}
}

func TestBufferSlice(t *testing.T) {
	buf := newBuffer(10)
	buf.write([]byte("hello world"))
	
	slice := buf.slice(0, 5)
	expected := []byte("hello")
	if !bytes.Equal(slice, expected) {
		t.Errorf("buffer.slice(0, 5) = %v, expected %v", slice, expected)
	}
	
	slice = buf.slice(6, 11)
	expected = []byte("world")
	if !bytes.Equal(slice, expected) {
		t.Errorf("buffer.slice(6, 11) = %v, expected %v", slice, expected)
	}
	
	// Test invalid slice
	slice = buf.slice(-1, 5)
	if slice != nil {
		t.Errorf("buffer.slice(-1, 5) = %v, expected nil", slice)
	}
	
	slice = buf.slice(0, 20)
	if slice != nil {
		t.Errorf("buffer.slice(0, 20) = %v, expected nil", slice)
	}
}

func TestBufferPool(t *testing.T) {
	// Get buffer from pool
	buf1 := getBuffer()
	if buf1 == nil {
		t.Fatal("getBuffer() returned nil")
	}
	
	// Use the buffer
	buf1.write([]byte("test"))
	
	// Return to pool
	putBuffer(buf1)
	
	// Buffer should be reset after returning to pool
	if buf1.len() != 0 {
		t.Errorf("buffer.len() after putBuffer = %d, expected 0", buf1.len())
	}
	
	// Get another buffer (might be the same one)
	buf2 := getBuffer()
	if buf2 == nil {
		t.Fatal("getBuffer() returned nil")
	}
	
	// Should be clean
	if buf2.len() != 0 {
		t.Errorf("reused buffer.len() = %d, expected 0", buf2.len())
	}
	
	putBuffer(buf2)
}