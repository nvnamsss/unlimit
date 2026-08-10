package efflux

import (
	"context"

	"github.com/voidforge-studios/unlimit/algo"
)

// ShardingWorkerPool routes tasks to a deterministic worker shard by task ID hash.
type ShardingWorkerPool[T ShardedTask] struct {
	hasher  algo.ConsistentHasher
	workers []Worker[T]
}

// NewShardingWorkerPool creates a sharded worker pool where each shard owns an independent worker pool.
func NewShardingWorkerPool[T ShardedTask](
	workers []Worker[T],
) *ShardingWorkerPool[T] {
	numShards := len(workers)

	pool := &ShardingWorkerPool[T]{
		workers: workers,
		hasher:  algo.ConsistentHasher{Buckets: numShards},
	}

	return pool
}

// GetShard returns the deterministic shard index for a task ID.
func (swp *ShardingWorkerPool[T]) GetWorker(task T) Worker[T] {
	taskID := task.GetID()
	if taskID == "" {
		return swp.workers[0]
	}
	index := swp.hasher.GetBucket(taskID)
	return swp.workers[index]
}

func (swp *ShardingWorkerPool[T]) Submit(ctx context.Context, task T) error {
	worker := swp.GetWorker(task)
	worker.Process(ctx, task)
	return nil
}

type ShardingWorkerPoolWR[T ShardedTask, R any] struct {
	hasher  algo.ConsistentHasher
	workers []WorkerWR[T, R]
}

func NewShardingWorkerPoolWR[T ShardedTask, R any](
	workers []WorkerWR[T, R],
) *ShardingWorkerPoolWR[T, R] {
	numShards := len(workers)

	pool := &ShardingWorkerPoolWR[T, R]{
		workers: workers,
		hasher:  algo.ConsistentHasher{Buckets: numShards},
	}

	return pool
}

func (swp *ShardingWorkerPoolWR[T, R]) GetWorker(task T) WorkerWR[T, R] {
	taskID := task.GetID()
	if taskID == "" {
		return swp.workers[0]
	}
	index := swp.hasher.GetBucket(taskID)
	return swp.workers[index]
}

func (swp *ShardingWorkerPoolWR[T, R]) Submit(ctx context.Context, task T) (R, error) {
	worker := swp.GetWorker(task)
	return worker.Process(ctx, task)
}
