package anycache

import (
	"context"
	"sync/atomic"
)

func (a *anyCache[K, V]) mGet(ctx context.Context, keys []K) ([]V, error) {
	switch a.strategy {
	case StrategySourceFirst:
		return a.mGetSourceFirst(ctx, keys)
	case StrategyCacheOnly:
		_, values, err := a.mGetCache(ctx, keys)
		return values, err
	case StrategyCacheFirst:
		fallthrough
	default:
		return a.mGetCacheFirst(ctx, keys)
	}
}

func (a *anyCache[K, V]) mGetCacheFirst(ctx context.Context, keys []K) ([]V, error) {
	missKeyIndices, result, err := a.mGetCache(ctx, keys)
	if err != nil {
		result = make([]V, len(keys))
		missKeyIndices = make([]int, len(keys))
		for i := range missKeyIndices {
			missKeyIndices[i] = i
		}
	}
	if len(missKeyIndices) == 0 {
		return result, nil
	}

	missKeys := make([]K, len(missKeyIndices))
	for i := range missKeyIndices {
		missKeys[i] = keys[missKeyIndices[i]]
	}
	values, err := a.mGetSource(ctx, missKeys)
	if err != nil {
		return result, nil
	}
	for i, value := range values {
		result[missKeyIndices[i]] = value
	}
	_ = a.mSet(ctx, missKeys, values)
	return result, nil
}

func (a *anyCache[K, V]) mGetSourceFirst(ctx context.Context, keys []K) ([]V, error) {
	values, err := a.mGetSource(ctx, keys)
	if err == nil {
		return values, nil
	}
	_, values, err = a.mGetCache(ctx, keys)
	return values, err
}

func (a *anyCache[K, V]) mGetCache(ctx context.Context, keys []K) (missKeyIndices []int, values []V, err error) {
	entries, err := a.cache.MGet(ctx, a.buildKeys(keys))
	if err != nil {
		return nil, nil, err
	}
	values = make([]V, len(keys))
	for i, entry := range entries {
		if entries[i] != nil {
			values[i] = entry.Value()
		} else {
			values[i] = a.emptyValue
			missKeyIndices = append(missKeyIndices, i)
		}
	}
	atomic.AddInt64(&a.hit, int64(len(keys)-len(missKeyIndices)))
	return missKeyIndices, values, nil
}

func (a *anyCache[K, V]) mGetSource(ctx context.Context, keys []K) ([]V, error) {
	atomic.AddInt64(&a.source, int64(len(keys)))
	values, err := a.batchLoader.BatchLoad(ctx, keys)
	if err != nil {
		return nil, err
	}
	_ = a.mSet(ctx, keys, values)
	return values, nil
}
