package cache

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}) error
	SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, toValue interface{}) error
	GetInt64(ctx context.Context, key string) (int64, error)

	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)

	HSet(ctx context.Context, key, field string, value interface{}) error
	HGet(ctx context.Context, key, field string, toValue interface{}) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HLen(ctx context.Context, key string) (int64, error)

	SetBit(ctx context.Context, key string, offset int64, value int) error
	GetBit(ctx context.Context, key string, offset int64) (bool, error)
	BitCount(ctx context.Context, key string) (int64, error)

	LLen(ctx context.Context, key string) (int64, error)
	LPop(ctx context.Context, key string, toValue interface{}) error
	LPush(ctx context.Context, key string, value interface{}) error
	LSet(ctx context.Context, key string, index int64, value interface{}) error
	LGet(ctx context.Context, key string, index int64, toValue interface{}) error
	Del(ctx context.Context, key string) error
}

type TypedCache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V) bool
	Delete(key K) bool
	Len() int
}

// TypedCacheWithExpiration extends TypedCache with time-to-live support.
// It provides methods for setting cache entries with expiration times.
// Example usage:
//
//	cache := NewTTLCache[string, *User](1000)
//	cache.SetWithTTL("user:123", user, 5*time.Minute)
//	value, ok := cache.Get("user:123")
//
// This interface is useful for caching data that should automatically
// expire after a specified duration.
type TypedCacheWithExpiration[K comparable, V any] interface {
	TypedCache[K, V]
	// SetWithTTL stores a value with a time-to-live duration.
	// The entry will be automatically removed after the TTL expires.
	// Returns true if an eviction occurred due to capacity limits.
	SetWithTTL(key K, value V, ttl time.Duration) bool
}
