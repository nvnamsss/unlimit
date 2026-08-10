package gen

import (
	"math/rand"
	"strings"
	"time"
)

// BasePrimitiveGenerator provides common functionality for primitive type generators
type BasePrimitiveGenerator struct {
	rand *rand.Rand
}

// NewBasePrimitiveGenerator creates a new base generator with seeded random source
func NewBasePrimitiveGenerator() *BasePrimitiveGenerator {
	return &BasePrimitiveGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// BoolGenerator generates random boolean values
type BoolGenerator struct {
	*BasePrimitiveGenerator
	dataGen DataGenerator[bool]
}

// NewBoolGenerator creates a new generator for boolean values
func NewBoolGenerator() DataGenerator[bool] {
	bg := &BoolGenerator{
		BasePrimitiveGenerator: NewBasePrimitiveGenerator(),
	}
	bg.dataGen = NewDataGenerator(bg.generateRandom())
	return bg
}

// Generate creates a random boolean value
func (g *BoolGenerator) Generate() bool {
	return g.dataGen.Generate()
}

// generateRandom creates a random boolean
func (g *BoolGenerator) generateRandom() bool {
	return g.rand.Intn(2) == 1
}

// IntGenerator generates random integer values
type IntGenerator[T int64 | int32 | int16 | int8 | int | uint64 | uint32 | uint16 | uint8 | uint] struct {
	*BasePrimitiveGenerator
	generateFunc func() T
}

// NewIntGenerator creates a new generator for integer values
func NewIntGenerator[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64]() DataGenerator[T] {
	var zero T
	g := &IntGenerator[T]{
		BasePrimitiveGenerator: NewBasePrimitiveGenerator(),
	}

	switch any(zero).(type) {
	case int:
		g.generateFunc = g.generateInt
	case int8:
		g.generateFunc = g.generateInt8
	case int16:
		g.generateFunc = g.generateInt16
	case int32:
		g.generateFunc = g.generateInt32
	case int64:
		g.generateFunc = g.generateInt64
	case uint:
		g.generateFunc = g.generateUint
	case uint8:
		g.generateFunc = g.generateUint8
	case uint16:
		g.generateFunc = g.generateUint16
	case uint32:
		g.generateFunc = g.generateUint32
	case uint64:
		g.generateFunc = g.generateUint64
	default:
		panic("unsupported type for IntGenerator")
	}

	return g
}

// Generate creates a random integer value of the specified type
func (g *IntGenerator[T]) Generate() T {
	return g.generateFunc()
}

func (g *IntGenerator[T]) generateInt() T {
	return any(g.rand.Intn(2001) - 1000).(T)
}

func (g *IntGenerator[T]) generateInt8() T {
	return any(int8(g.rand.Intn(256) - 128)).(T)
}

func (g *IntGenerator[T]) generateInt16() T {
	return any(int16(g.rand.Intn(65536) - 32768)).(T)
}

func (g *IntGenerator[T]) generateInt32() T {
	return any(g.rand.Int31()).(T)
}

func (g *IntGenerator[T]) generateInt64() T {
	return any(g.rand.Int63n(2001) - 1000).(T)
}

func (g *IntGenerator[T]) generateUint() T {
	return any(uint(g.rand.Intn(2001))).(T)
}

func (g *IntGenerator[T]) generateUint8() T {
	return any(uint8(g.rand.Intn(256))).(T)
}

func (g *IntGenerator[T]) generateUint16() T {
	return any(uint16(g.rand.Intn(65536))).(T)
}

func (g *IntGenerator[T]) generateUint32() T {
	return any(uint32(g.rand.Uint32())).(T)
}

func (g *IntGenerator[T]) generateUint64() T {
	return any(g.rand.Uint64() % 2001).(T)
}

// FloatGenerator generates random floating-point values
type FloatGenerator[T float32 | float64] struct {
	*BasePrimitiveGenerator
	generateFunc func() T
}

// NewFloatGenerator creates a new generator for floating-point values
func NewFloatGenerator[T float32 | float64]() DataGenerator[T] {
	var zero T
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var generate func() T

	switch any(zero).(type) {
	case float32:
		generate = func() T { return any(float32(r.Float32()*200 - 100)).(T) }
	case float64:
		generate = func() T { return any(r.Float64()*200 - 100).(T) }
	default:
		panic("unsupported type for FloatGenerator")
	}

	return &FloatGenerator[T]{
		BasePrimitiveGenerator: NewBasePrimitiveGenerator(),
		generateFunc:           generate,
	}
}

// Generate creates a random floating-point value of the specified type
func (g *FloatGenerator[T]) Generate() T {
	return g.generateFunc()
}

// StringGenerator generates random string values
type StringGenerator struct {
	*BasePrimitiveGenerator
	chars  string
	length int
	r      *rand.Rand
}

// NewStringGenerator creates a new generator for string values
// length parameter controls the length of generated strings (default 8 if <= 0)
func NewStringGenerator(length int) DataGenerator[string] {
	if length <= 0 {
		length = 8
	}

	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return &StringGenerator{
		BasePrimitiveGenerator: NewBasePrimitiveGenerator(),
		chars:                  chars,
		length:                 length,
		r:                      r,
	}
}

// Generate creates a random string value
func (g *StringGenerator) Generate() string {
	return g.generateRandom()
}

// generateRandom creates a random string with the specified length
func (g *StringGenerator) generateRandom() string {
	result := strings.Builder{}
	for i := 0; i < g.length; i++ {
		result.WriteByte(g.chars[g.r.Intn(len(g.chars))])
	}
	return result.String()
}
