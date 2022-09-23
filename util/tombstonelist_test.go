package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTombstoneList_Capacity(t *testing.T) {
	tl := NewTombstoneList[string](3)
	i := 0
	for i < 3 {
		assert.NoErrorf(t, tl.Append("ok"), "err")
		i++
	}
	assert.Error(t, tl.Append("boom"), "err")
}

func TestTombstoneList_FindAvailableSlot(t *testing.T) {
	tl := NewTombstoneList[string](3)
	i := 0
	for i < 3 {
		_ = tl.Append("ok")
		i++
	}
	tl.Remove(1)
	assert.NoErrorf(t, tl.Append("boom"), "err")
	vs := tl.Values
	expected := []string{"ok", "boom", "ok"}
	assert.ElementsMatch(t, vs, expected)
}

func TestTombstoneList_RemoveAll(t *testing.T) {
	tl := NewTombstoneList[*string](3)
	i := 0
	for i < 3 {
		s := "ok"
		_ = tl.Append(&s)
		i++
	}
	tl.Remove(0)
	tl.Remove(1)
	tl.Remove(2)
	vs := tl.Values
	expected := []*string{nil, nil, nil}
	assert.ElementsMatch(t, expected, vs)
	s := "new"
	_ = tl.Append(&s)
	vs = tl.Values
	expected = []*string{&s, nil, nil}
	assert.ElementsMatch(t, expected, vs)
}
