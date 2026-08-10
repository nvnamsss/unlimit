package gen

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

// Default generation settings
const (
	defaultStringLength = 8
	defaultSliceLength  = 3
	defaultMapEntries   = 2
)

// TypeGenerator is a strategy interface for generating values of specific types
type TypeGenerator interface {
	// Generate generates a value for the given type
	Generate(t reflect.Type) interface{}
}

// ReflectDataGenerator generates data for any type using reflection
type ReflectDataGenerator[T any] struct {
	idGen            IDGenerator
	stringLength     int
	sliceLength      int
	mapEntries       int
	customGenerators map[reflect.Type]func() interface{}
	intGenerator     DataGenerator[int]
	int8Generator    DataGenerator[int8]
	int16Generator   DataGenerator[int16]
	int32Generator   DataGenerator[int32]
	int64Generator   DataGenerator[int64]
	uintGenerator    DataGenerator[uint]
	uint8Generator   DataGenerator[uint8]
	uint16Generator  DataGenerator[uint16]
	uint32Generator  DataGenerator[uint32]
	uint64Generator  DataGenerator[uint64]

	float32Generator DataGenerator[float32]
	float64Generator DataGenerator[float64]

	stringGenerator DataGenerator[string]
	boolGenerator   DataGenerator[bool]

	// typeGenerators   map[reflect.Kind]func() interface{}
	rand         *rand.Rand
	maxDepth     int
	currentDepth int
	typeInstance T // Used for reflection to determine type
}

// ReflectDataGeneratorOption is a function that configures the generator
type ReflectDataGeneratorOption[T any] func(*ReflectDataGenerator[T])

// WithSliceLength sets the length of generated slices
func WithSliceLength[T any](length int) ReflectDataGeneratorOption[T] {
	return func(g *ReflectDataGenerator[T]) {
		if length >= 0 {
			g.sliceLength = length
		}
	}
}

// WithMapEntries sets the number of entries in generated maps
func WithMapEntries[T any](entries int) ReflectDataGeneratorOption[T] {
	return func(g *ReflectDataGenerator[T]) {
		if entries >= 0 {
			g.mapEntries = entries
		}
	}
}

// WithMaxDepth sets the maximum nesting depth for struct fields
func WithMaxDepth[T any](depth int) ReflectDataGeneratorOption[T] {
	return func(g *ReflectDataGenerator[T]) {
		if depth < 0 {
			depth = math.MaxInt32
		}

		if depth > 0 {
			g.maxDepth = depth
		}
	}
}

// WithCustomGenerator registers a custom generator for a specific type
func WithCustomGenerator[T any](typ reflect.Type, generator func() interface{}) ReflectDataGeneratorOption[T] {
	return func(g *ReflectDataGenerator[T]) {
		g.customGenerators[typ] = generator
	}
}

// WithIDGenerator sets the ID generator to use for ID fields
func WithIDGenerator[T any](idGen IDGenerator) ReflectDataGeneratorOption[T] {
	return func(g *ReflectDataGenerator[T]) {
		if idGen != nil {
			g.idGen = idGen
		}
	}
}

// NewReflectDataGenerator creates a new ReflectDataGenerator with the given options
func NewReflectDataGenerator[T any](typeInstance T, options ...ReflectDataGeneratorOption[T]) DataGenerator[T] {
	generator := &ReflectDataGenerator[T]{
		idGen:            NewUUIDGenerator(""),
		stringLength:     defaultStringLength,
		sliceLength:      defaultSliceLength,
		mapEntries:       defaultMapEntries,
		customGenerators: make(map[reflect.Type]func() interface{}),
		rand:             rand.New(rand.NewSource(time.Now().UnixNano())),
		maxDepth:         5,
		currentDepth:     0,
		typeInstance:     typeInstance,
	}

	// Register default type generators
	generator.registerDefaultGenerators()

	// Apply all options
	for _, option := range options {
		option(generator)
	}

	return generator
}

// Generate implements the DataGenerator interface
func (g *ReflectDataGenerator[T]) Generate() T {
	// Get the type of T using reflection
	t := reflect.TypeOf(g.typeInstance)
	result := g.generate(t)

	// Convert the result to type T
	if result == nil {
		return *new(T) // Return zero value if result is nil
	}

	// Try to convert the result to T
	if typedResult, ok := result.(T); ok {
		return typedResult
	}

	// If we can't convert directly, try to use reflection
	if t.Kind() == reflect.Ptr && reflect.ValueOf(result).Kind() == reflect.Ptr {
		return reflect.ValueOf(result).Interface().(T)
	}

	// If none of the above works, return the zero value
	return *new(T)
}

// registerDefaultGenerators registers the default strategies for generating values
func (g *ReflectDataGenerator[T]) registerDefaultGenerators() {
	g.boolGenerator = NewBoolGenerator()
	g.intGenerator = NewIntGenerator[int]()
	g.int8Generator = NewIntGenerator[int8]()
	g.int16Generator = NewIntGenerator[int16]()
	g.int32Generator = NewIntGenerator[int32]()
	g.int64Generator = NewIntGenerator[int64]()
	g.uintGenerator = NewIntGenerator[uint]()
	g.uint8Generator = NewIntGenerator[uint8]()
	g.uint16Generator = NewIntGenerator[uint16]()
	g.uint32Generator = NewIntGenerator[uint32]()
	g.uint64Generator = NewIntGenerator[uint64]()
	g.float32Generator = NewFloatGenerator[float32]()
	g.float64Generator = NewFloatGenerator[float64]()
	g.stringGenerator = NewStringGenerator(g.stringLength)
}

// generate creates data for the given type
func (g *ReflectDataGenerator[T]) generate(t reflect.Type) interface{} {
	// Check for custom generator
	if generator, ok := g.customGenerators[t]; ok {
		return generator()
	}

	// Check depth limit for complex types
	if g.currentDepth >= g.maxDepth {
		// Return zero value for the type when max depth is reached
		return reflect.Zero(t).Interface()
	}

	// Process based on type kind for complex types not handled by strategies
	switch t.Kind() {
	case reflect.Bool:
		return g.boolGenerator.Generate()
	case reflect.Int:
		return g.intGenerator.Generate()
	case reflect.Int8:
		return g.int8Generator.Generate()
	case reflect.Int16:
		return g.int16Generator.Generate()
	case reflect.Int32:
		return g.int32Generator.Generate()
	case reflect.Int64:
		return g.int64Generator.Generate()
	case reflect.Uint:
		return g.uintGenerator.Generate()
	case reflect.Uint8:
		return g.uint8Generator.Generate()
	case reflect.Uint16:
		return g.uint16Generator.Generate()
	case reflect.Uint32:
		return g.uint32Generator.Generate()
	case reflect.Uint64:
		return g.uint64Generator.Generate()
	case reflect.Float32:
		return g.float32Generator.Generate()
	case reflect.Float64:
		return g.float64Generator.Generate()
	case reflect.String:
		return g.stringGenerator.Generate()
	case reflect.Array:
		return g.generateArray(t)
	case reflect.Slice:
		return g.generateSlice(t)
	case reflect.Map:
		return g.generateMap(t)
	case reflect.Struct:
		return g.generateStruct(t)
	case reflect.Ptr:
		return g.generatePtr(t)
	case reflect.Interface:
		// For interfaces, we can only return nil since we don't know concrete type
		return nil
	default:
		// For other types, return zero value
		return reflect.Zero(t).Interface()
	}
}

// GenerateValue creates data for the given value by using its type
func (g *ReflectDataGenerator[T]) GenerateValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	return g.generate(reflect.TypeOf(v))
}

// GenerateStruct creates and fills a struct with random data
func (g *ReflectDataGenerator[T]) GenerateStruct(v interface{}) error {
	if v == nil {
		return fmt.Errorf("nil value provided")
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("value must be a pointer to a struct, got %T", v)
	}

	structVal := val.Elem()
	structType := structVal.Type()

	g.fillStruct(structVal, structType)
	return nil
}

// Private generation methods for complex types
func (g *ReflectDataGenerator[T]) generateArray(t reflect.Type) interface{} {
	arrayLen := t.Len()
	elemType := t.Elem()

	// Create an array with the same length as the array type
	arrayVal := reflect.New(t).Elem()

	// Fill the array with random values
	g.currentDepth++
	for i := 0; i < arrayLen; i++ {
		arrayVal.Index(i).Set(reflect.ValueOf(g.generate(elemType)))
	}
	g.currentDepth--

	return arrayVal.Interface()
}

func (g *ReflectDataGenerator[T]) generateSlice(t reflect.Type) interface{} {
	elemType := t.Elem()

	// Create a slice with the specified length
	sliceVal := reflect.MakeSlice(t, g.sliceLength, g.sliceLength)

	// Fill the slice with random values
	g.currentDepth++
	for i := 0; i < g.sliceLength; i++ {
		sliceVal.Index(i).Set(reflect.ValueOf(g.generate(elemType)))
	}
	g.currentDepth--

	return sliceVal.Interface()
}

func (g *ReflectDataGenerator[T]) generateMap(t reflect.Type) interface{} {
	keyType := t.Key()
	elemType := t.Elem()

	// Create a new map
	mapVal := reflect.MakeMap(t)

	// Special handling for complex key types that may not be valid map keys
	if !isValidMapKey(keyType) {
		return mapVal.Interface()
	}

	// Add entries to the map
	g.currentDepth++
	for i := 0; i < g.mapEntries; i++ {
		retries := 10
		// deterministic key generation with retries
		// to avoid collisions in the map
		// This is a simple retry mechanism to ensure unique keys
		for retries > 0 {
			// Generate a key and check if it already exists in the map
			key := reflect.ValueOf(g.generate(keyType))
			if !key.IsValid() || !key.CanInterface() {
				continue // Skip invalid keys
			}

			// If the key already exists, retry
			if mapVal.MapIndex(key).IsValid() {
				retries--
				continue
			}

			value := reflect.ValueOf(g.generate(elemType))
			mapVal.SetMapIndex(key, value)
			break
		}
	}
	g.currentDepth--

	return mapVal.Interface()
}

func (g *ReflectDataGenerator[T]) generateStruct(t reflect.Type) interface{} {
	// Create a new instance of the struct
	structVal := reflect.New(t).Elem()

	g.fillStruct(structVal, t)

	return structVal.Interface()
}

func (g *ReflectDataGenerator[T]) fillStruct(structVal reflect.Value, t reflect.Type) {
	g.currentDepth++
	defer func() { g.currentDepth-- }()

	numFields := t.NumField()
	for i := 0; i < numFields; i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldVal := structVal.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		// Special handling for ID fields
		if isIDField(field) {
			id := g.idGen.Generate()
			fieldVal.Set(reflect.ValueOf(id))
			continue
		}

		// Generate a value for the field
		genValue := g.generate(field.Type)
		if genValue != nil {
			fieldVal.Set(reflect.ValueOf(genValue))
		}
	}
}

func (g *ReflectDataGenerator[T]) generatePtr(t reflect.Type) interface{} {
	// Get the type the pointer points to
	elemType := t.Elem()

	// Create a new pointer value
	ptrVal := reflect.New(elemType)

	// If we haven't reached max depth, generate a value for the pointer
	g.currentDepth++
	if g.currentDepth < g.maxDepth {
		ptrVal.Elem().Set(reflect.ValueOf(g.generate(elemType)))
	}
	g.currentDepth--

	return ptrVal.Interface()
}

// Helper functions

func isValidMapKey(t reflect.Type) bool {
	// Only certain types can be map keys in Go
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	default:
		return false
	}
}

func isIDField(field reflect.StructField) bool {
	name := field.Name
	return name == "ID" || name == "Id" || strings.HasSuffix(name, "ID") || strings.HasSuffix(name, "Id")
}
