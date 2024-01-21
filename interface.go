package anycache

import (
	"context"
)

type Fetcher[K any, V any] interface {
	Get(ctx context.Context, key K) (V, error)
	MGet(ctx context.Context, keys []K) ([]V, error)
	Set(ctx context.Context, key K, value V) error
	MSet(ctx context.Context, keys []K, values []V) error
	Del(ctx context.Context, keys ...K) error
	Refresh(ctx context.Context, keys ...K) error
}

type Loader[K any, V any] interface {
	Load(ctx context.Context, key K) (V, error)
}

type loadFunc[K any, V any] func(ctx context.Context, key K) (V, error)

func (l loadFunc[K, V]) Load(ctx context.Context, key K) (V, error) {
	return l(ctx, key)
}

type BatchLoader[K any, V any] interface {
	BatchLoad(ctx context.Context, keys []K) ([]V, error)
}

type batchLoadFunc[K any, V any] func(ctx context.Context, keys []K) ([]V, error)

func (b batchLoadFunc[K, V]) BatchLoad(ctx context.Context, keys []K) ([]V, error) {
	return b(ctx, keys)
}
