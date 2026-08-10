package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack"
)

type redisCache struct {
	client *redis.Client
}

func (c *redisCache) Set(ctx context.Context, key string, v interface{}) error {
	return c.SetWithExpiration(ctx, key, v, 0)
}

func (c *redisCache) SetWithExpiration(ctx context.Context, key string, v interface{}, expiration time.Duration) error {
	switch v.(type) {
	case string, int, int32, int64, float64, bool, []byte:
		return c.client.Set(ctx, key, v, expiration).Err()
	default:
		data, err := msgpack.Marshal(v)
		if err != nil {
			return err
		}
		return c.client.Set(ctx, key, data, expiration).Err()
	}
}

func (c *redisCache) Get(ctx context.Context, key string, v interface{}) error {
	switch ptr := v.(type) {
	case *string:
		val, err := getFromRedis[string](ctx, c.client, key)
		if err != nil {
			return err
		}
		*ptr = val
	case *int:
		val, err := getFromRedis[int](ctx, c.client, key)
		if err != nil {
			return err
		}
		*ptr = val
	case *int32:
		val, err := getFromRedis[int](ctx, c.client, key)
		if err != nil {
			return err
		}
		*ptr = int32(val)
	case *int64:
		val, err := getFromRedis[int64](ctx, c.client, key)
		if err != nil {
			return err
		}
		*ptr = val
	case *float64:
		val, err := getFromRedis[float64](ctx, c.client, key)
		if err != nil {
			return err
		}
		*ptr = val
	case *bool:
		val, err := getFromRedis[bool](ctx, c.client, key)
		if err != nil {
			return err
		}
		*ptr = val
	default:
		data, err := c.client.Get(ctx, key).Bytes()
		if err != nil {
			return err
		}
		return msgpack.Unmarshal(data, v)
	}
	return nil
}

func getFromRedis[T any](ctx context.Context, client *redis.Client, key string) (T, error) {
	var zero T
	cmd := client.Get(ctx, key)
	if err := cmd.Err(); err != nil {
		return zero, err
	}

	switch any(zero).(type) {
	case string:
		return any(cmd.Val()).(T), nil
	case int:
		val, err := cmd.Int()
		return any(val).(T), err
	case int64:
		val, err := cmd.Int64()
		return any(val).(T), err
	case float64:
		val, err := cmd.Float64()
		return any(val).(T), err
	case bool:
		val, err := cmd.Bool()
		return any(val).(T), err
	default:
		data, err := cmd.Bytes()
		if err != nil {
			return zero, err
		}
		var result T
		if err := msgpack.Unmarshal(data, &result); err != nil {
			return zero, err
		}
		return result, nil
	}
}

func (c *redisCache) GetInt64(ctx context.Context, key string) (int64, error) {
	return c.client.Get(ctx, key).Int64()
}

func (c *redisCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *redisCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

func (r *redisCache) HSet(ctx context.Context, key, field string, value interface{}) error {
	data, err := msgpack.Marshal(value)
	if err != nil {
		return nil
	}

	_, err = r.client.HSet(ctx, key, data).Result()
	return err
}

func (c *redisCache) HGet(ctx context.Context, key, field string, value interface{}) error {
	data, err := c.client.HGet(ctx, key, field).Bytes()
	if err != nil {
		return err
	}

	return msgpack.Unmarshal(data, value)
}

func (c *redisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

func (c *redisCache) HLen(ctx context.Context, key string) (int64, error) {
	return c.client.HLen(ctx, key).Result()
}

func (c *redisCache) SetBit(ctx context.Context, key string, offset int64, value int) error {
	return c.client.SetBit(ctx, key, offset, value).Err()
}

func (c *redisCache) GetBit(ctx context.Context, key string, offset int64) (bool, error) {
	res, err := c.client.GetBit(ctx, key, offset).Result()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

func (c *redisCache) BitCount(ctx context.Context, key string) (int64, error) {
	return c.client.BitCount(ctx, key, &redis.BitCount{
		Start: 0,
		End:   -1,
	}).Result()
}

func (c *redisCache) LLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

func (c *redisCache) LPop(ctx context.Context, key string, toValue interface{}) error {
	data, err := c.client.LPop(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return msgpack.Unmarshal(data, toValue)
}

func (c *redisCache) LPush(ctx context.Context, key string, value interface{}) error {
	data, err := msgpack.Marshal(value)
	if err != nil {
		return nil
	}

	return c.client.LPush(ctx, key, data).Err()
}

func (c *redisCache) LSet(ctx context.Context, key string, index int64, value interface{}) error {
	data, err := msgpack.Marshal(value)
	if err != nil {
		return nil
	}
	return c.client.LSet(ctx, key, index, data).Err()
}

func (c *redisCache) LGet(ctx context.Context, key string, index int64, toValue interface{}) error {
	data, err := c.client.LIndex(ctx, key, index).Bytes()
	if err != nil {
		return nil
	}
	return msgpack.Unmarshal(data, toValue)
}

func (c *redisCache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func NewRedisCache(client *redis.Client) Cache {
	return &redisCache{
		client: client,
	}
}
