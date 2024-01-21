package anycache

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"testing"
)

func TestMGetCacheFirst(t *testing.T) {
	mapCache := NewMapCache[string]()
	fetcher := New[int, string](mapCache).
		WithStrategy(StrategyCacheFirst).
		WithLoadFunc(func(ctx context.Context, key int) (string, error) {
			return fmt.Sprint(key), nil
		}).Build()
	// source
	v, _ := fetcher.MGet(ctx, []int{123, 456})
	if !slices.Equal(v, []string{"123", "456"}) || fetcher.(*anyCache[int, string]).source != 2 ||
		!maps.Equal(mapCache.Map, map[string]string{"123": "123", "456": "456"}) {
		t.Fatalf("unexpected value: %v, source: %d, mapCache: %v", v, fetcher.(*anyCache[int, string]).source, mapCache.Map)
	}
	// cache
	v, _ = fetcher.MGet(ctx, []int{123, 456})
	if !slices.Equal(v, []string{"123", "456"}) || fetcher.(*anyCache[int, string]).source != 2 {
		t.Fatalf("unexpected value: %v, source: %d", v, fetcher.(*anyCache[int, string]).source)
	}
	// partly
	v, _ = fetcher.MGet(ctx, []int{123, 78, 456})
	if !slices.Equal(v, []string{"123", "78", "456"}) || fetcher.(*anyCache[int, string]).source != 3 {
		t.Fatalf("unexpected value: %v, source: %d", v, fetcher.(*anyCache[int, string]).source)
	}
	// cache failed
	mapCache.Fail = true
	v, _ = fetcher.MGet(ctx, []int{123, 78, 456})
	if !slices.Equal(v, []string{"123", "78", "456"}) || fetcher.(*anyCache[int, string]).source != 6 {
		t.Fatalf("unexpected value: %v, source: %d", v, fetcher.(*anyCache[int, string]).source)
	}
}

func TestMGetCacheOnly(t *testing.T) {
	mapCache := NewMapCache[string]()
	fetcher := New[int, string](mapCache).
		WithStrategy(StrategyCacheOnly).
		WithLoadFunc(func(ctx context.Context, key int) (string, error) {
			return fmt.Sprint(key), nil
		}).Build()
	// cache nil
	v, _ := fetcher.MGet(ctx, []int{123, 456})
	if !slices.Equal(v, []string{"", ""}) || fetcher.(*anyCache[int, string]).source != 0 {
		t.Fatalf("unexpected value: %v, source: %d", v, fetcher.(*anyCache[int, string]).source)
	}

	_ = fetcher.Refresh(ctx, 123, 456)
	// cache exist
	v, _ = fetcher.MGet(ctx, []int{123, 456})
	if !slices.Equal(v, []string{"123", "456"}) || fetcher.(*anyCache[int, string]).source != 0 {
		t.Fatalf("unexpected value: %v, source: %d", v, fetcher.(*anyCache[int, string]).source)
	}
	// partly
	v, _ = fetcher.MGet(ctx, []int{123, 78, 456})
	if !slices.Equal(v, []string{"123", "", "456"}) || fetcher.(*anyCache[int, string]).source != 0 {
		t.Fatalf("unexpected value: %v, source: %d", v, fetcher.(*anyCache[int, string]).source)
	}
}

func TestMGetSourceFirst(t *testing.T) {
	mapCache := NewMapCache[string]()
	fetcher := New[int, string](mapCache).
		WithStrategy(StrategySourceFirst).
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
	// source
	v, _ := fetcher.MGet(ctx, []int{123, 456})
	if !slices.Equal(v, []string{"123", "456"}) || fetcher.(*anyCache[int, string]).source != 2 ||
		!maps.Equal(mapCache.Map, map[string]string{"123": "123", "456": "456"}) {
		t.Fatalf("unexpected value: %s, source: %d, mapCache: %v", v, fetcher.(*anyCache[int, string]).source, mapCache.Map)
	}
	//
	v, _ = fetcher.MGet(ctx, []int{123, 456})
	if !slices.Equal(v, []string{"123", "456"}) || fetcher.(*anyCache[int, string]).source != 4 {
		t.Fatalf("unexpected value: %s, source: %d", v, fetcher.(*anyCache[int, string]).source)
	}

	// source failed
	v, _ = fetcher.MGet(ctx, []int{123, 78, 456})
	if !slices.Equal(v, []string{"123", "", "456"}) || fetcher.(*anyCache[int, string]).source != 7 {
		t.Fatalf("unexpected value: %s, source: %d", v, fetcher.(*anyCache[int, string]).source)
	}
}
