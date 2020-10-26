package ioutilx

import (
	"bufio"
	"testing"

	"gotest.tools/v3/assert"
)

func TestNopWriteCloser(t *testing.T) {
	var (
		c   = NopWriteCloser(bufio.ReadWriter{})
		err = c.Close()
	)
	assert.NilError(t, err)
}
