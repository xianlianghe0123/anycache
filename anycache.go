package anycache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/xianlianghe0123/anycache/cache"
)

type (
	strategy int
)

const (
	StrategyCacheFirst = iota
	StrategySourceFirst
	StrategyCacheOnly
)

type IAnyCache[K any, V any] interface {
	WithGenKeyFunc(genKeyFunc func(t K) string) IAnyCache[K, V]
	WithStrategy(strategy strategy) IAnyCache[K, V]
	WithExpiration(expiration time.Duration) IAnyCache[K, V]
	WithNameSpace(namespace string) IAnyCache[K, V]
	WithEmptyValue(v V) IAnyCache[K, V]
	WithLoader(loader Loader[K, V]) IAnyCache[K, V]
	WithLoadFunc(loader loadFunc[K, V]) IAnyCache[K, V]
	WithBatchLoader(batchLoader BatchLoader[K, V]) IAnyCache[K, V]
	WithBatchLoadFunc(batchLoader batchLoadFunc[K, V]) IAnyCache[K, V]
	Build() Fetcher[K, V]
}

type anyCache[K any, V any] struct {
	strategy    strategy
	genKeyFunc  func(t K) string
	cache       cache.Cacher[V]
	loader      Loader[K, V]
	batchLoader BatchLoader[K, V]

	namespace  string
	emptyValue V
	expiration time.Duration

	hit    int64
	source int64
}

func New[K any, V any](cache cache.Cacher[V]) IAnyCache[K, V] {
	if cache == nil {
		panic("nil cache")
	}
	return &anyCache[K, V]{
		strategy:    StrategyCacheFirst,
		genKeyFunc:  func(k K) string { return fmt.Sprint(k) },
		cache:       cache,
		loader:      nil,
		batchLoader: nil,
		namespace:   "",
		expiration:  0,
	}
}

func (a *anyCache[K, V]) WithGenKeyFunc(genKeyFunc func(t K) string) IAnyCache[K, V] {
	a.genKeyFunc = genKeyFunc
	return a
}

func (a *anyCache[K, V]) WithStrategy(strategy strategy) IAnyCache[K, V] {
	a.strategy = strategy
	return a
}

func (a *anyCache[K, V]) WithExpiration(expiration time.Duration) IAnyCache[K, V] {
	a.expiration = expiration
	return a
}

func (a *anyCache[K, V]) WithNameSpace(namespace string) IAnyCache[K, V] {
	a.namespace = namespace
	return a
}

func (a *anyCache[K, V]) WithEmptyValue(v V) IAnyCache[K, V] {
	a.emptyValue = v
	return a
}

func (a *anyCache[K, V]) WithLoader(loader Loader[K, V]) IAnyCache[K, V] {
	if loader == nil {
		panic("loader is nil")
	}
	a.loader = loader
	return a
}

func (a *anyCache[K, V]) WithLoadFunc(loader loadFunc[K, V]) IAnyCache[K, V] {
	return a.WithLoader(loader)
}

func (a *anyCache[K, V]) WithBatchLoader(batchLoader BatchLoader[K, V]) IAnyCache[K, V] {
	if batchLoader == nil {
		panic("batchLoader is nil")
	}
	a.batchLoader = batchLoader
	return a
}

func (a *anyCache[K, V]) WithBatchLoadFunc(batchLoader batchLoadFunc[K, V]) IAnyCache[K, V] {
	return a.WithBatchLoader(batchLoader)
}

func (a *anyCache[K, V]) Build() Fetcher[K, V] {
	if a.loader == nil && a.batchLoader == nil {
		panic("no loader")
	}
	if a.loader == nil {
		a.WithLoadFunc(func(ctx context.Context, key K) (V, error) {
			values, err := a.batchLoader.BatchLoad(ctx, []K{key})
			if err != nil {
				return a.emptyValue, err
			}
			if len(values) == 0 {
				return a.emptyValue, errors.New("not found")
			}
			return values[0], nil
		})
	}
	if a.batchLoader == nil {
		a.WithBatchLoadFunc(func(ctx context.Context, keys []K) ([]V, error) {
			values := make([]V, len(keys))
			for i, key := range keys {
				value, err := a.loader.Load(ctx, key)
				if err != nil {
					values[i] = a.emptyValue
				} else {
					values[i] = value
				}
			}
			return values, nil
		})
	}
	return a
}

func (a *anyCache[K, V]) Get(ctx context.Context, key K) (V, error) {
	return a.get(ctx, key)
}

func (a *anyCache[K, V]) MGet(ctx context.Context, keys []K) ([]V, error) {
	return a.mGet(ctx, keys)
}

func (a *anyCache[K, V]) Set(ctx context.Context, key K, value V) error {
	return a.mSet(ctx, []K{key}, []V{value})
}

func (a *anyCache[K, V]) MSet(ctx context.Context, keys []K, values []V) error {
	if len(keys) != len(values) {
		return errors.New("keys and values length not equal")
	}
	if len(keys) == 0 {
		return nil
	}
	return a.mSet(ctx, keys, values)
}

func (a *anyCache[K, V]) Del(ctx context.Context, keys ...K) error {
	if len(keys) == 0 {
		return nil
	}
	return a.del(ctx, keys...)
}

func (a *anyCache[K, V]) Refresh(ctx context.Context, keys ...K) error {
	if len(keys) == 0 {
		return nil
	}
	return a.refresh(ctx, keys...)
}

func (a *anyCache[K, V]) mSet(ctx context.Context, keys []K, values []V) error {
	entries := make([]cache.Entry[V], 0, len(values))
	for i, key := range keys {
		entries = append(entries, a.newEntry(a.buildKey(key), values[i]))
	}
	return a.cache.Set(ctx, entries...)
}

func (a *anyCache[K, V]) del(ctx context.Context, keys ...K) error {
	return a.cache.Del(ctx, a.buildKeys(keys)...)
}

func (a *anyCache[K, V]) refresh(ctx context.Context, keys ...K) error {
	values, err := a.batchLoader.BatchLoad(ctx, keys)
	if err != nil {
		return err
	}
	return a.mSet(ctx, keys, values)
}

// common
func (a *anyCache[K, V]) buildKey(key K) string {
	if a.namespace == "" {
		return a.genKeyFunc(key)
	}
	return fmt.Sprintf("%s:%s", a.namespace, a.genKeyFunc(key))
}

func (a *anyCache[K, V]) buildKeys(keys []K) []string {
	cacheKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		cacheKeys = append(cacheKeys, a.buildKey(k))
	}
	return cacheKeys
}

func (a *anyCache[K, V]) newEntry(key string, value V) cache.Entry[V] {
	return cache.NewEntry(key, value, a.expiration)
}
