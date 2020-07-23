package sink

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

type polarBear struct {
	format Format
	food   []interface{}
}

func (b *polarBear) Flush() error { return nil }

func (b *polarBear) Init(_ context.Context, format Format) {
	b.format = format
}

func (b *polarBear) Write(n interface{}) error {
	var err error
	n, err = b.format.Do(n)
	if err != nil {
		return err
	}
	b.food = append(b.food, n)
	return nil
}

type meat struct {
	Chewed bool
}

type mouth struct{}

func (*mouth) Do(n interface{}) (interface{}, error) {
	m := n.(*meat)
	m.Chewed = true
	return m, nil
}

func TestSink(t *testing.T) {
	cases := []struct {
		desc string
		s    Sink
		f    Format
		data []interface{}
		out  []interface{}
	}{
		{
			desc: "simple",
			s:    &polarBear{},
			f:    &mouth{},
			data: []interface{}{
				&meat{},
				&meat{},
				&meat{},
			},
			out: []interface{}{
				&meat{Chewed: true},
				&meat{Chewed: true},
				&meat{Chewed: true},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			c.s.Init(context.TODO(), c.f)
			for _, d := range c.data {
				c.s.Write(d)
			}
			bear := c.s.(*polarBear)
			assert.DeepEqual(t, bear.food, c.out)
		})
	}
}
