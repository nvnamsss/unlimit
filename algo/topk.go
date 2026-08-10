package algo

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type TopK interface {
	Create(ctx context.Context, collection string, k int, width int, depth int, decay float64) error
	IncrBy(ctx context.Context, collection string, key string, count int64) error
	Elements(ctx context.Context, collection string) ([]string, error)
}

type RedisTopK struct {
	client *redis.Client
}

func (r *RedisTopK) Create(ctx context.Context, collection string, k, width, depth int, decay float64) error {
	return r.client.Do(ctx, "TOPK.RESERVE", collection, k, width, decay).Err()
}

func (r *RedisTopK) IncrBy(ctx context.Context, collection string, key string, count int64) error {
	return r.client.Do(ctx, "TOPK.INCRBY", collection, key, count).Err()
}

func (r *RedisTopK) Elements(ctx context.Context, collection string) ([]string, error) {
	elems, err := r.client.Do(ctx, "TOPK.LIST", collection).Result()
	if err != nil {
		return nil, err
	}
	if elems == nil {
		return nil, nil
	}
	return elems.([]string), nil
}

func NewRedisTopK(client *redis.Client) TopK {
	return &RedisTopK{
		client: client,
	}
}
