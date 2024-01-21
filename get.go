package anycache

import (
	"context"
	"sync/atomic"
)

func (a *anyCache[K, V]) get(ctx context.Context, key K) (V, error) {
	switch a.strategy {
	case StrategySourceFirst:
		return a.getSourceFirst(ctx, key)
	case StrategyCacheOnly:
		return a.getCache(ctx, key)
	case StrategyCacheFirst:
		fallthrough
	default:
		return a.getCacheFirst(ctx, key)
	}
}

func (a *anyCache[K, V]) getCacheFirst(ctx context.Context, key K) (V, error) {

	value, err := a.getCache(ctx, key)
	if err == nil {
		return value, nil
	}
	value, err = a.getSource(ctx, key)
	if err != nil {
		return value, err
	}
	return value, nil
}

func (a *anyCache[K, V]) getSourceFirst(ctx context.Context, key K) (V, error) {
	value, err := a.getSource(ctx, key)
	if err == nil {
		return value, nil
	}
	return a.getCache(ctx, key)
}

func (a *anyCache[K, V]) getCache(ctx context.Context, key K) (V, error) {
	value, err := a.cache.Get(ctx, a.buildKey(key))
	if err != nil {
		return a.emptyValue, err
	}
	atomic.AddInt64(&a.hit, 1)
	return value.Value(), nil
}

func (a *anyCache[K, V]) getSource(ctx context.Context, key K) (V, error) {
	atomic.AddInt64(&a.source, 1)
	value, err := a.loader.Load(ctx, key)
	if err != nil {
		return value, err
	}
	_ = a.cache.Set(ctx, a.newEntry(a.buildKey(key), value))
	return value, nil
}
