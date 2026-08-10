package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/voidforge-studios/unlimit/stability"
)

func main() {
	// v1()
	// v2()
	v3_direct()
	// v3_batch()
	v4_batch()
}

func v1() {
	srv, err := miniredis.Run()
	if err != nil {
		panic(fmt.Sprintf("start miniredis failed: %v", err))
	}
	defer srv.Close()

	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	defer func() {
		_ = client.Close()
	}()

	// run allow commands

	var (
		times int = 1e5
	)

	// rate limiter
	rl, err := stability.NewRedisRateLimiter(client, stability.RedisRateLimiterConfig{
		Limit: float64(times * 1000),
		Burst: times * 1000,
	})
	if err != nil {
		panic(err)
	}

	log.Printf("Start rate limiter")
	now := time.Now()
	wg := sync.WaitGroup{}
	for i := 0; i < int(times); i++ {
		wg.Add(1)
		go func(idx int) {
			key := strconv.Itoa(idx)
			if !rl.Allow(context.Background(), key) {
				// log.Printf("oh no")
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	log.Printf("Rate limiter done, elapsed: %v, stats=%+v", time.Since(now), rl.Stats())
}

func v2() {
	srv, err := miniredis.Run()
	if err != nil {
		panic(fmt.Sprintf("start miniredis failed: %v", err))
	}
	defer srv.Close()

	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	defer func() {
		_ = client.Close()
	}()

	// run allow commands

	var (
		times int = 1e5
	)

	// rate limiter v2
	rl, err := stability.NewRedisRateLimiterV2(client, stability.RedisRateLimiterV2Config{
		Limit:            float64(times * 1000),
		Burst:            times * 1000,
		PipelineEnabled:  true,
		PipelineMaxWait:  10 * time.Microsecond,
		PipelineMaxBatch: 1000,
		PipelineShards:   64,
	})
	if err != nil {
		panic(err)
	}

	log.Printf("Start rate limiter")
	now := time.Now()
	wg := sync.WaitGroup{}
	for i := 0; i < int(times); i++ {
		wg.Add(1)
		go func(idx int) {
			key := strconv.Itoa(idx)
			rl.Allow(context.Background(), key)
			wg.Done()
		}(i)
	}

	wg.Wait()
	log.Printf("Rate limiter done, elapsed: %v, stats=%+v", time.Since(now), rl.Stats())
}

func v3_direct() {
	srv, err := miniredis.Run()
	if err != nil {
		panic(fmt.Sprintf("start miniredis failed: %v", err))
	}
	defer srv.Close()

	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	defer func() {
		_ = client.Close()
	}()

	// run allow commands

	var (
		times int = 1e5
	)

	// rate limiter v3
	rl, err := stability.NewRedisRateLimiterV3(client, stability.RedisRateLimiterV3Config{
		Limit: float64(times * 1000),
		Burst: times * 1000,
	})
	if err != nil {
		panic(err)
	}

	// initialize the keys by refilling tokens
	for i := 0; i < int(times); i++ {
		key := strconv.Itoa(i)
		if !rl.Allow(context.Background(), key) {
		}
	}

	log.Printf("Start rate limiter")
	now := time.Now()
	wg := sync.WaitGroup{}
	for i := 0; i < int(times); i++ {
		wg.Add(1)
		go func(idx int) {
			key := strconv.Itoa(idx)
			rl.Allow(context.Background(), key)
			wg.Done()
		}(i)
	}

	wg.Wait()
	log.Printf("Rate limiter done, elapsed: %v, stats=%+v", time.Since(now), rl.Stats())
}

func v3_batch() {
	srv, err := miniredis.Run()
	if err != nil {
		panic(fmt.Sprintf("start miniredis failed: %v", err))
	}
	defer srv.Close()

	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	defer func() {
		_ = client.Close()
	}()

	// run allow commands

	var (
		times int = 1e5
	)

	// rate limiter v3
	rl, err := stability.NewRedisRateLimiterBatchV3(client, stability.RedisRateLimiterBatchV3Config{
		Limit:            float64(times * 1000),
		Burst:            times * 1000,
		PipelineMaxWait:  10 * time.Microsecond,
		PipelineMaxBatch: 1000,
		PipelineShards:   64,
	})
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	// initialize the keys by refilling tokens
	for i := 0; i < int(times); i++ {
		key := strconv.Itoa(i)
		log.Printf("%v", i)
		wg.Add(1)
		go func() {
			rl.Allow(context.Background(), key)
			wg.Done()
		}()
	}

	wg.Wait()
	time.Sleep(10 * time.Second)
	log.Printf("Start rate limiter")
	now := time.Now()
	for i := 0; i < int(times); i++ {
		wg.Add(1)
		go func(idx int) {
			key := strconv.Itoa(idx)
			rl.Allow(context.Background(), key)
			wg.Done()
		}(i)
	}

	wg.Wait()
	log.Printf("Rate limiter done, elapsed: %v, stats=%+v", time.Since(now), rl.Stats())
}

func v4_batch() {
	srv, err := miniredis.Run()
	if err != nil {
		panic(fmt.Sprintf("start miniredis failed: %v", err))
	}
	defer srv.Close()

	client := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	defer func() {
		_ = client.Close()
	}()

	// run allow commands

	var (
		times int = 1e5
	)

	// rate limiter v4
	rl, err := stability.NewRedisRateLimiterBatchV4(client, stability.RedisRateLimiterBatchV4Config{
		Limit:            float64(times * 1000),
		Burst:            times * 1000,
		OperationTimeout: 10 * time.Second,
		PipelineMaxWait:  10 * time.Microsecond,
		PipelineMaxBatch: 100,
		PipelineShards:   512,
	})
	if err != nil {
		panic(err)
	}

	// initialize the keys by refilling tokens
	for i := 0; i < int(times); i++ {
		key := strconv.Itoa(i)
		if err := rl.InitializeTokenBucket(context.Background(), key, times); err != nil {
			panic(err)
		}
	}

	log.Printf("Start rate limiter")
	now := time.Now()
	wg := sync.WaitGroup{}
	for i := 0; i < int(times); i++ {
		wg.Add(1)
		go func(idx int) {
			key := strconv.Itoa(idx)
			rl.Allow(context.Background(), key)
			wg.Done()
		}(i)
	}

	wg.Wait()
	log.Printf("Rate limiter done, elapsed: %v, stats=%+v", time.Since(now), rl.Stats())
}
