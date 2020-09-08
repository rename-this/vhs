package safebuffer

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestSafeBuffer(t *testing.T) {
	buf := &SafeBuffer{}

	n, err := buf.Write([]byte("111"))
	assert.NilError(t, err)
	assert.Equal(t, n, 3)

	p := make([]byte, 3)
	n, err = buf.Read(p)
	assert.NilError(t, err)
	assert.Equal(t, n, 3)
	assert.Equal(t, string(p), "111")

	n, err = buf.Write([]byte("2222"))
	assert.NilError(t, err)
	assert.Equal(t, n, 4)
	assert.Equal(t, string(buf.Bytes()), "2222")
}
