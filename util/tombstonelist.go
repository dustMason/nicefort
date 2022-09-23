package util

import (
	"errors"
	"github.com/gammazero/deque"
)

type TombstoneList[T any] struct {
	Values     []T
	tombstones *deque.Deque[int]
	capacity   int
}

func NewTombstoneList[T any](capacity int, items ...T) *TombstoneList[T] {
	tl := &TombstoneList[T]{
		Values:     make([]T, 0, capacity),
		tombstones: deque.New[int](),
		capacity:   capacity,
	}
	if len(items) > 0 {
		tl.Values = items
	}
	return tl
}

func (t *TombstoneList[T]) Append(e T) (int, error) {
	if len(t.Values) == t.capacity && t.tombstones.Len() == 0 {
		return -1, errors.New("no more capacity")
	}
	if len(t.Values) == t.capacity {
		i := t.tombstones.PopBack()
		t.Values[i] = e
		return i, nil
	}
	t.Values = append(t.Values, e)
	return len(t.Values) - 1, nil
}

func (t *TombstoneList[T]) Set(i int, e T) {
	t.Values[i] = e
}

func (t *TombstoneList[T]) Get(i int) T {
	var empty T
	if i > len(t.Values)-1 {
		return empty
	}
	return t.Values[i]
}

func (t *TombstoneList[T]) Remove(i int) {
	var empty T
	t.Values[i] = empty
	t.tombstones.PushBack(i)
}
