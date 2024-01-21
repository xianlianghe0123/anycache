package anycache

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"testing"

	"github.com/xianlianghe0123/anycache/cache"
)

var ctx = context.Background()

func TestSingle_Set(t *testing.T) {
	mapCache := NewMapCache[string]()
	fetcher := New[int, string](mapCache).
		WithNameSpace("test").
		WithGenKeyFunc(func(t int) string { return fmt.Sprintf("%d~", t) }).
		WithEmptyValue("").
		WithLoadFunc(func(ctx context.Context, key int) (string, error) {
			return "", nil
		}).
		Build()
	_ = fetcher.Set(ctx, 123, "123")
	if !maps.Equal(mapCache.Map, map[string]string{"test:123~": "123"}) {
		t.Fatalf("set mapCache failed")
	}
}

func TestBatch_MSet(t *testing.T) {
	mapCache := NewMapCache[string]()
	fetcher := New[int, string](mapCache).
		WithNameSpace("test").
		WithBatchLoadFunc(func(ctx context.Context, keys []int) ([]string, error) {
			t := make([]string, len(keys))
			for i := range keys {
				t[i] = fmt.Sprint(keys[i])
			}
			return t, nil
		}).Build()
	_ = fetcher.MSet(ctx, []int{123, 456}, []string{"123", "456"})
	if !maps.Equal(mapCache.Map, map[string]string{"test:123": "123", "test:456": "456"}) {
		t.Fatalf("set mapCache failed")
	}
}

func TestSingle_RefreshAndDel(t *testing.T) {
	mapCache := NewMapCache[string]()
	fetcher := New[int, string](mapCache).
		WithLoadFunc(func(ctx context.Context, key int) (string, error) {
			return fmt.Sprint(key), nil
		}).
		Build()

	_ = fetcher.Refresh(ctx, 123, 456)
	if !maps.Equal(mapCache.Map, map[string]string{"123": "123", "456": "456"}) {
		t.Fatalf("set cache failed")
	}
	_ = fetcher.Del(ctx, 123, 456)
	if !maps.Equal(mapCache.Map, map[string]string{}) {
		t.Fatalf("del cache failed")
	}
}

type MapCache[V any] struct {
	Map  map[string]V
	Fail bool
}

func NewMapCache[V any]() *MapCache[V] {
	return &MapCache[V]{
		Map: make(map[string]V),
	}
}

func (t *MapCache[V]) newEntry(key string, value V) cache.Entry[V] {
	return cache.NewEntry(key, value, 0)
}

func (t *MapCache[V]) Get(ctx context.Context, key string) (cache.Entry[V], error) {
	if t.Fail {
		return nil, errors.New("error")
	}
	v, ok := t.Map[key]
	if !ok {
		return nil, errors.New("nil")
	}
	return t.newEntry(key, v), nil
}

func (t *MapCache[V]) MGet(ctx context.Context, keys []string) ([]cache.Entry[V], error) {
	if t.Fail {
		return nil, errors.New("error")
	}
	res := make([]cache.Entry[V], len(keys))
	for i, key := range keys {
		if v, ok := t.Map[key]; ok {
			res[i] = t.newEntry(key, v)
		}
	}
	return res, nil
}

func (t *MapCache[V]) Set(ctx context.Context, entries ...cache.Entry[V]) error {
	if t.Fail {
		return errors.New("error")
	}
	for _, entry := range entries {
		t.Map[entry.Key()] = entry.Value()
	}
	return nil
}
func (t *MapCache[V]) Del(ctx context.Context, keys ...string) error {
	if t.Fail {
		return errors.New("error")
	}
	for _, key := range keys {
		delete(t.Map, key)
	}
	return nil
}
