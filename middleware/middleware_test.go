package middleware

import (
	"bytes"
	"io/ioutil"
	"testing"

	"gotest.tools/assert"
)

type leopard struct {
	NumSpots int
}

func TestExec(t *testing.T) {
	cases := []struct {
		desc        string
		mware       *Middleware
		num         int
		l           *leopard
		out         []*leopard
		errContains string
	}{
		{
			desc:  "no exec path",
			mware: &Middleware{},
			l:     &leopard{NumSpots: 111},
			out:   []*leopard{{NumSpots: 111}},
		},
		{
			desc: "change host",
			mware: &Middleware{
				stdin:  ioutil.Discard,
				stdout: ioutil.NopCloser(bytes.NewBufferString("{\"NumSpots\":222}\n{\"NumSpots\":333}\n")),
			},
			num: 2,
			l:   &leopard{NumSpots: 111},
			out: []*leopard{
				{NumSpots: 222},
				{NumSpots: 333},
			},
		},
		{
			desc: "bad JSON",
			mware: &Middleware{
				stdin:  ioutil.Discard,
				stdout: ioutil.NopCloser(bytes.NewBufferString("{\"NumSpots\":"))},
			num:         1,
			l:           &leopard{NumSpots: 111},
			errContains: "failed to unmarshal",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			for i := 0; i < c.num; i++ {
				req, err := c.mware.Exec(c.l)
				if c.errContains != "" {
					assert.ErrorContains(t, err, c.errContains)
				} else {
					assert.DeepEqual(t, c.out[i], req)
				}
			}
		})
	}
}
