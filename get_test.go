package anycache

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"testing"
)

func TestGetCacheFirst(t *testing.T) {
	mapCache := NewMapCache[string]()
	fail := false
	fetcher := New[int, string](mapCache).
		WithStrategy(StrategyCacheFirst).
		WithBatchLoadFunc(func(ctx context.Context, keys []int) ([]string, error) {
			if fail {
				return nil, errors.New("error")
			}
			t := make([]string, len(keys))
			for i, k := range keys {
				if k == 78 {
					return nil, errors.New("error")
				}
				t[i] = fmt.Sprint(k)
			}
			return t, nil
		}).Build()
	// source failed
	fail = true
	v, err := fetcher.Get(ctx, 123)
	if err == nil {
		t.Fatalf("unexpected success")
	}
	fail = false
	// source
	v, _ = fetcher.Get(ctx, 123)
	if v != "123" || fetcher.(*anyCache[int, string]).source != 2 || !maps.Equal(mapCache.Map, map[string]string{"123": "123"}) {
		t.Fatalf("unexpected value: %s, source: %d, mapCache: %v", v, fetcher.(*anyCache[int, string]).source, mapCache.Map)
	}
	// cache
	v, _ = fetcher.Get(ctx, 123)
	if v != "123" || fetcher.(*anyCache[int, string]).source != 2 {
		t.Fatalf("unexpected value: %s, source: %d", v, fetcher.(*anyCache[int, string]).source)
	}
}

func TestGetCacheOnly(t *testing.T) {
	mapCache := NewMapCache[string]()
	fetcher := New[int, string](mapCache).
		WithStrategy(StrategyCacheOnly).
		WithBatchLoadFunc(func(ctx context.Context, keys []int) ([]string, error) {
			t := make([]string, len(keys))
			for i, k := range keys {
				if k == 78 {
					return nil, errors.New("error")
				}
				t[i] = fmt.Sprint(k)
			}
			return t, nil
		}).Build()
	// cache nil
	v, err := fetcher.Get(ctx, 123)
	if err == nil || fetcher.(*anyCache[int, string]).source != 0 {
		t.Fatalf("unexpected err: %s, source: %d", err, fetcher.(*anyCache[int, string]).source)
	}
	// cache exist
	_ = fetcher.Refresh(ctx, 123)
	v, err = fetcher.Get(ctx, 123)
	if err != nil || v != "123" {
		t.Fatalf("unexpected err: %s, value: %s", err, v)
	}
}

func TestGetSourceFirst(t *testing.T) {
	mapCache := NewMapCache[string]()
	fail := false
	fetcher := New[int, string](mapCache).
		WithStrategy(StrategySourceFirst).
		WithBatchLoadFunc(func(ctx context.Context, keys []int) ([]string, error) {
			if fail {
				return nil, errors.New("")
			}
			t := make([]string, len(keys))
			for i, k := range keys {
				if k == 78 {
					return nil, errors.New("error")
				}
				t[i] = fmt.Sprint(k)
			}
			return t, nil
		}).Build()
	// source
	v, _ := fetcher.Get(ctx, 123)
	if v != "123" || fetcher.(*anyCache[int, string]).source != 1 || !maps.Equal(mapCache.Map, map[string]string{"123": "123"}) {
		t.Fatalf("unexpected value: %s, source: %d, mapCache: %v", v, fetcher.(*anyCache[int, string]).source, mapCache.Map)
	}

	v, _ = fetcher.Get(ctx, 123)
	if v != "123" || fetcher.(*anyCache[int, string]).source != 2 {
		t.Fatalf("unexpected value: %s, source: %d", v, fetcher.(*anyCache[int, string]).source)
	}
	// source failed then cache
	fail = true
	v, err := fetcher.Get(ctx, 123)
	if err != nil && v != "123" || fetcher.(*anyCache[int, string]).source != 3 {
		t.Fatalf("unexpected err: %s, value: %s, source: %d", err, v, fetcher.(*anyCache[int, string]).source)
	}
}
