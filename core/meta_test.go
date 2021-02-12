package core

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestMeta(t *testing.T) {
	type fox struct {
		name string
	}

	m := NewMeta("id", map[string]interface{}{
		"str": "111",
		"num": 111,
		"fox": &fox{name: "Tails"},
	})

	assert.Equal(t, "id", m.SourceID)

	n, ok := m.GetString("str")
	assert.Assert(t, ok)
	assert.Equal(t, n, "111")

	f, ok := m.Get("fox")
	assert.Assert(t, ok)
	assert.Equal(t, "Tails", f.(*fox).name)

	_, ok = m.GetString("nosuchkey")
	assert.Assert(t, !ok)

	m.Set("num", 111111)
	x, ok := m.Get("num")
	assert.Assert(t, ok)
	assert.Equal(t, x.(int), 111111)

	m2 := NewMeta("", nil)
	assert.Assert(t, m2.values != nil)

	em := EmptyMeta(nil)
	assert.Assert(t, em.Meta() != nil)
}
