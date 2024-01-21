package cache

import "time"

type Entry[V any] interface {
	Key() string
	Value() V
	Expiration() time.Duration
}

type entry[V any] struct {
	key        string
	value      V
	expiration time.Duration
}

func NewEntry[V any](key string, value V, expiration time.Duration) Entry[V] {
	return &entry[V]{
		key:        key,
		value:      value,
		expiration: expiration,
	}
}

func (e *entry[V]) Key() string {
	return e.key
}

func (e *entry[V]) Value() V {
	return e.value
}

func (e *entry[V]) Expiration() time.Duration {
	return e.expiration
}
