package cache

import (
	"context"
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/vmihailenco/msgpack"
)

var (
	ErrMethodNotSupported = errors.New("This method is not supported")
	ErrKeyNotFound        = errors.New("Key not found")
	ErrInvalidItem        = errors.New("Item is invalid")
)

type memoryCache struct {
	mc *cache.Cache
}

func (c *memoryCache) Set(ctx context.Context, key string, v interface{}) error {
	return c.SetWithExpiration(ctx, key, v, 0)
}

func (c *memoryCache) SetWithExpiration(ctx context.Context, key string, v interface{}, expiration time.Duration) error {
	data, err := msgpack.Marshal(v)
	if err != nil {
		return nil
	}
	c.mc.Set(key, data, expiration)
	return nil
}

func (c *memoryCache) Get(ctx context.Context, key string, v interface{}) error {
	item, ok := c.mc.Get(key)
	if !ok {
		return ErrKeyNotFound
	}
	data, ok := item.([]byte)
	if !ok {
		return ErrInvalidItem
	}
	return msgpack.Unmarshal(data, v)
}

func (c *memoryCache) GetInt64(ctx context.Context, key string) (int64, error) {
	item, ok := c.mc.Get(key)
	if !ok {
		return 0, ErrKeyNotFound
	}
	var rs int64
	data, ok := item.([]byte)
	if !ok {
		return 0, ErrInvalidItem
	}

	if err := msgpack.Unmarshal(data, &rs); err != nil {
		return 0, err
	}
	return rs, nil
}

func (c *memoryCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.mc.IncrementInt64(key, 1)
}

func (c *memoryCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.mc.IncrementInt64(key, value)
}

func (r *memoryCache) HSet(ctx context.Context, key, field string, value interface{}) error {
	return ErrMethodNotSupported
}

func (c *memoryCache) HGet(ctx context.Context, key, field string, value interface{}) error {
	return ErrMethodNotSupported
}

func (c *memoryCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return nil, ErrMethodNotSupported
}

func (c *memoryCache) HLen(ctx context.Context, key string) (int64, error) {
	return 0, ErrMethodNotSupported
}

func (c *memoryCache) SetBit(ctx context.Context, key string, offset int64, value int) error {
	return ErrMethodNotSupported
}

func (c *memoryCache) GetBit(ctx context.Context, key string, offset int64) (bool, error) {
	return false, ErrMethodNotSupported
}

func (c *memoryCache) BitCount(ctx context.Context, key string) (int64, error) {
	return 0, ErrMethodNotSupported
}

func (c *memoryCache) LLen(ctx context.Context, key string) (int64, error) {
	return 0, ErrMethodNotSupported
}

func (c *memoryCache) LPop(ctx context.Context, key string, toValue interface{}) error {
	return ErrMethodNotSupported
}

func (c *memoryCache) LPush(ctx context.Context, key string, value interface{}) error {
	return ErrMethodNotSupported
}

func (c *memoryCache) LSet(ctx context.Context, key string, index int64, value interface{}) error {
	return ErrMethodNotSupported
}

func (c *memoryCache) LGet(ctx context.Context, key string, index int64, toValue interface{}) error {
	return ErrMethodNotSupported
}

func (c *memoryCache) Del(ctx context.Context, key string) error {
	c.mc.Delete(key)
	return nil
}

func NewMemoryCache(expiration time.Duration, cleanUpInterval time.Duration) Cache {
	return &memoryCache{
		mc: cache.New(expiration, cleanUpInterval),
	}
}
