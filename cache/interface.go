package cache

import "context"

type Cacher[V any] interface {
	Get(ctx context.Context, key string) (Entry[V], error)
	MGet(ctx context.Context, keys []string) ([]Entry[V], error)
	Set(ctx context.Context, entry ...Entry[V]) error
	Del(ctx context.Context, key ...string) error
}
