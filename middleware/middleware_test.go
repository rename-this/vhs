package middleware

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	"github.com/rename-this/vhs/session"
	"gotest.tools/v3/assert"
)

type leopard struct {
	NumSpots int
}

func TestExec(t *testing.T) {
	cases := []struct {
		desc        string
		m           Middleware
		header      []byte
		num         int
		l           *leopard
		out         []*leopard
		errContains string
	}{
		{
			desc: "no exec path",
			m:    &mware{},
			l:    &leopard{NumSpots: 111},
			out:  []*leopard{{NumSpots: 111}},
		},
		{
			desc: "change host",
			m: &mware{
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
			desc: "change host with header",
			m: &mware{
				stdin:  ioutil.Discard,
				stdout: ioutil.NopCloser(bytes.NewBufferString("{\"NumSpots\":222}\n{\"NumSpots\":333}\n")),
			},
			header: []byte{'1'},
			num:    2,
			l:      &leopard{NumSpots: 111},
			out: []*leopard{
				{NumSpots: 222},
				{NumSpots: 333},
			},
		},
		{
			desc: "bad JSON",
			m: &mware{
				stdin:  ioutil.Discard,
				stdout: ioutil.NopCloser(bytes.NewBufferString("{\"NumSpots\":")),
			},
			num:         1,
			l:           &leopard{NumSpots: 111},
			errContains: "failed to unmarshal",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			ctx := session.NewContexts(nil, nil, nil)
			for i := 0; i < c.num; i++ {
				req, err := c.m.Exec(ctx, c.header, c.l)
				if c.errContains != "" {
					assert.ErrorContains(t, err, c.errContains)
				} else {
					assert.DeepEqual(t, c.out[i], req)
				}
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	cases := []struct {
		desc        string
		command     string
		ns          []interface{}
		expected    []interface{}
		errContains string
	}{
		{
			desc:    "success",
			command: "../testdata/double.bash",
			ns: []interface{}{
				&leopard{1}, &leopard{2}, &leopard{3},
			},
			expected: []interface{}{
				&leopard{2}, &leopard{4}, &leopard{6},
			},
		},
		{
			desc:    "no command",
			command: "",
		},
		{
			desc:    "unsupported JSON data",
			command: "../testdata/double.bash",
			ns: []interface{}{
				make(chan struct{}),
			},
			errContains: "unsupported type: chan struct {}",
		},
		{
			desc:        "bad middleware",
			command:     "../testdata/cannot_exec.go",
			errContains: "failed to start middleware command",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			err := func() error {
				ctx := session.NewContexts(nil, nil, nil)
				m, err := New(ctx, c.command)
				assert.NilError(t, err)

				if c.command == "" {
					assert.Equal(t, nil, m)
					return nil
				}

				if err := m.Start(); err != nil {
					return err
				}

				go m.Wait()

				assert.NilError(t, err)

				time.Sleep(800 * time.Millisecond)

				out := make([]interface{}, 0, len(c.ns))
				for _, n := range c.ns {
					o, err := m.Exec(ctx, nil, n)
					if err != nil {
						return err
					}
					out = append(out, o)
				}

				m.Close()
				ctx.Cancel()

				assert.DeepEqual(t, out, c.expected)

				return nil
			}()

			if c.errContains != "" {
				assert.ErrorContains(t, err, c.errContains)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
