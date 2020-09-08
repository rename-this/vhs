package safebuffer

import (
	"bytes"
	"sync"
)

// SafeBuffer is a bytes.Buffer that is safe
// to be used by multiple goroutines. This is meant
// to be used for testing purposes.
type SafeBuffer struct {
	b  bytes.Buffer
	mu sync.Mutex
}

// Read reads the buffer.
func (b *SafeBuffer) Read(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Read(p)
}

// Write writes the buffer.
func (b *SafeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

// Bytes get the bytes from the buffer
func (b *SafeBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Bytes()
}
