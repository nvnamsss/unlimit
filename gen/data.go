package gen

type DataGenerator[T any] interface {
	Generate() T
}

func NewDataGenerator[T any](value T) DataGenerator[T] {
	return &dataGenerator[T]{value: value}
}

type dataGenerator[T any] struct {
	value T
}

func (d *dataGenerator[T]) Generate() T {
	return d.value
}
