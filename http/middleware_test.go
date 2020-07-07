package http

import (
	"bytes"
	"io/ioutil"
	"testing"

	"gotest.tools/assert"
)

func TestProcess(t *testing.T) {
	cases := []struct {
		desc        string
		mware       *Middleware
		num         int
		req         *Request
		out         []*Request
		errContains string
	}{
		{
			desc:  "no exec path",
			mware: &Middleware{},
			req:   &Request{Host: "111"},
			out:   []*Request{{Host: "111"}},
		},
		{
			desc: "change host",
			mware: &Middleware{
				stdin:  ioutil.Discard,
				stdout: ioutil.NopCloser(bytes.NewBufferString("{\"host\":\"222\"}\n{\"host\":\"333\"}\n")),
			},
			num: 2,
			req: &Request{Host: "111"},
			out: []*Request{
				{Host: "222"},
				{Host: "333"},
			},
		},
		{
			desc: "bad JSON",
			mware: &Middleware{
				stdin:  ioutil.Discard,
				stdout: ioutil.NopCloser(bytes.NewBufferString("{\"host\":\n")),
			},
			num:         1,
			req:         &Request{Host: "111"},
			errContains: "failed to unmarshal request",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			for i := 0; i < c.num; i++ {
				req, err := c.mware.Exec(c.req)
				if c.errContains != "" {
					assert.ErrorContains(t, err, c.errContains)
				} else {
					assert.DeepEqual(t, c.out[i], req)
				}
			}
		})
	}
}
