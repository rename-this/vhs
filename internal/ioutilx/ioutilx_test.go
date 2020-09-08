package ioutilx

import (
	"bufio"
	"io/ioutil"
	"testing"

	"github.com/go-errors/errors"
	"gotest.tools/assert"
)

type testCloser struct {
	err error
}

func (c *testCloser) Close() error { return c.err }

func TestClosers(t *testing.T) {
	cases := []struct {
		desc        string
		closers     Closers
		errsContain []string
	}{
		{
			desc: "no errors",
			closers: Closers{
				&testCloser{},
				&testCloser{},
				&testCloser{},
			},
		},
		{
			desc: "with errors",
			closers: Closers{
				&testCloser{err: errors.New("111")},
				&testCloser{err: errors.New("222")},
				&testCloser{err: errors.New("333")},
			},
			errsContain: []string{
				"111",
				"222",
				"333",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			err := c.closers.Close()
			if len(c.errsContain) == 0 {
				assert.NilError(t, err)
			} else {
				for _, e := range c.errsContain {
					assert.ErrorContains(t, err, e)
				}
			}
		})
	}
}

func TestNopWriteCloser(t *testing.T) {
	var (
		c   = NopWriteCloser(bufio.ReadWriter{})
		err = c.Close()
	)
	assert.NilError(t, err)
}

func TestNopReadCloserID(t *testing.T) {
	var (
		r  = NopReadCloserID(ioutil.NopCloser(bufio.ReadWriter{}))
		id = r.ID()
	)
	assert.Equal(t, "", id)
}
