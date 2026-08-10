package gen

import (
	"reflect"
	"testing"
	"time"
)

func TestReflectDataGenerator_Interface(t *testing.T) {
	// Test that ReflectDataGenerator properly implements DataGenerator interface
	var dataGen DataGenerator[Person]
	dataGen = NewReflectDataGenerator[Person](Person{})

	if dataGen == nil {
		t.Fatal("NewReflectDataGenerator should return a valid DataGenerator")
	}

	person := dataGen.Generate()

	// Verify the struct has been populated
	if person.ID == "" {
		t.Error("ID was not generated")
	}

	if person.Name == "" {
		t.Error("Name was not generated")
	}
}

func TestReflectDataGenerator_Options(t *testing.T) {
	// Test options configuration
	gen := NewReflectDataGenerator[Person](
		Person{},
		WithSliceLength[Person](5),
		WithMapEntries[Person](3),
		WithMaxDepth[Person](-1),
	)

	person := gen.Generate()

	// Verify slice length option
	if len(person.Tags) != 5 {
		t.Errorf("Expected 5 tags, got %d", len(person.Tags))
	}

	// Verify map entries option
	if len(person.Metadata) != 3 {
		t.Errorf("Expected 3 metadata entries, got %d", len(person.Metadata))
	}
}

func TestReflectDataGenerator_CustomGenerator(t *testing.T) {
	// Test custom generator for a specific type
	customEmail := "test@example.com"
	customName := "Custom Name"

	// We need to create a special generator for string fields since
	// the custom generator above would affect all strings
	customGen := NewReflectDataGenerator[Person](
		Person{},
		WithCustomGenerator[Person](
			reflect.TypeOf(Person{}),
			func() interface{} {
				return Person{
					Name:  customName,
					Email: customEmail,
				}
			},
		),
	)

	person := customGen.Generate()

	// Verify custom generation
	if person.Name != customName {
		t.Errorf("Expected custom name %s, got %s", customName, person.Name)
	}

	if person.Email != customEmail {
		t.Errorf("Expected custom email %s, got %s", customEmail, person.Email)
	}
}

func TestReflectDataGenerator_IDGeneration(t *testing.T) {
	// Test ID generation using custom ID generator
	prefixedIDGen := NewUUIDGenerator("test-")
	gen := NewReflectDataGenerator[Person](
		Person{},
		WithIDGenerator[Person](prefixedIDGen),
	)

	person := gen.Generate()

	// Verify ID has the prefix
	if !reflect.DeepEqual(person.ID[:5], "test-") {
		t.Errorf("Expected ID to start with 'test-', got %s", person.ID)
	}
}

func TestReflectDataGenerator_PrimitiveTypes(t *testing.T) {
	// Test generating primitive types
	testCases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "int",
			testFunc: func(t *testing.T) {
				gen := NewReflectDataGenerator[int](0)
				val := gen.Generate()
				if val == 0 {
					// Low probability this is consistently 0
					for i := 0; i < 5; i++ {
						if gen.Generate() != 0 {
							return // Success
						}
					}
					t.Error("Generated int is consistently 0, likely not generating properly")
				}
			},
		},
		{
			name: "string",
			testFunc: func(t *testing.T) {
				gen := NewReflectDataGenerator[string]("")
				val := gen.Generate()
				if val == "" {
					t.Error("Generated string is empty")
				}
			},
		},
		{
			name: "bool",
			testFunc: func(t *testing.T) {
				gen := NewReflectDataGenerator[bool](false)
				_ = gen.Generate() // Cannot reliably test random boolean
			},
		},
		{
			name: "float64",
			testFunc: func(t *testing.T) {
				gen := NewReflectDataGenerator[float64](0.0)
				val := gen.Generate()
				if val == 0.0 {
					// Low probability this is consistently 0
					for i := 0; i < 5; i++ {
						if gen.Generate() != 0.0 {
							return // Success
						}
					}
					t.Error("Generated float is consistently 0, likely not generating properly")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestReflectDataGenerator_ComplexTypes(t *testing.T) {
	t.Run("Struct", func(t *testing.T) {
		gen := NewReflectDataGenerator[Person](Person{})
		person := gen.Generate()

		if person.ID == "" || person.Name == "" || person.Email == "" {
			t.Error("Struct fields were not properly generated")
		}
	})

	t.Run("Slice", func(t *testing.T) {
		gen := NewReflectDataGenerator[[]string]([]string{})
		slice := gen.Generate()

		if len(slice) != defaultSliceLength {
			t.Errorf("Expected slice length %d, got %d", defaultSliceLength, len(slice))
		}

		for _, item := range slice {
			if item == "" {
				t.Error("Generated slice contains empty string")
			}
		}
	})

	t.Run("Map", func(t *testing.T) {
		gen := NewReflectDataGenerator[map[string]int](map[string]int{})
		m := gen.Generate()

		if len(m) != defaultMapEntries {
			t.Errorf("Expected map entries %d, got %d", defaultMapEntries, len(m))
		}

		for k := range m {
			if k == "" {
				t.Error("Generated map contains empty key")
			}
		}
	})

	t.Run("Pointer", func(t *testing.T) {
		gen := NewReflectDataGenerator[*Address](nil)
		addr := gen.Generate()

		if addr == nil {
			t.Error("Generated pointer is nil")
		}

		if addr.Street == "" || addr.City == "" {
			t.Error("Pointer fields were not properly generated")
		}
	})
}

func TestReflectDataGenerator_GenerateValue(t *testing.T) {
	gen := &ReflectDataGenerator[Person]{
		stringLength: 10,
		sliceLength:  3,
		mapEntries:   2,
		maxDepth:     3,
	}
	gen.registerDefaultGenerators()

	t.Run("GenerateValue with string", func(t *testing.T) {
		val := gen.GenerateValue("")
		if str, ok := val.(string); !ok || len(str) != 10 {
			t.Errorf("Expected string of length 10, got %v", val)
		}
	})

	t.Run("GenerateValue with struct", func(t *testing.T) {
		val := gen.GenerateValue(Address{})
		if _, ok := val.(Address); !ok {
			t.Errorf("Expected Address struct, got %T", val)
		}
	})

	t.Run("GenerateValue with nil", func(t *testing.T) {
		val := gen.GenerateValue(nil)
		if val != nil {
			t.Errorf("Expected nil, got %v", val)
		}
	})
}

func TestReflectDataGenerator_GenerateStruct(t *testing.T) {
	gen := &ReflectDataGenerator[Person]{
		stringLength: 10,
		sliceLength:  3,
		mapEntries:   2,
		maxDepth:     3,
	}
	gen.registerDefaultGenerators()

	t.Run("Valid struct pointer", func(t *testing.T) {
		person := &Person{}
		err := gen.GenerateStruct(person)

		if err != nil {
			t.Errorf("GenerateStruct returned error: %v", err)
		}

		if person.Name == "" || person.Email == "" {
			t.Error("Struct fields were not properly generated")
		}
	})

	t.Run("Nil value", func(t *testing.T) {
		err := gen.GenerateStruct(nil)

		if err == nil {
			t.Error("Expected error for nil value, got nil")
		}
	})

	t.Run("Non-struct pointer", func(t *testing.T) {
		str := "test"
		err := gen.GenerateStruct(&str)

		if err == nil {
			t.Error("Expected error for non-struct pointer, got nil")
		}
	})

	t.Run("Non-pointer", func(t *testing.T) {
		person := Person{}
		err := gen.GenerateStruct(person)

		if err == nil {
			t.Error("Expected error for non-pointer, got nil")
		}
	})
}

func TestReflectDataGenerator_MaxDepth(t *testing.T) {
	// Define a recursive struct type for testing max depth
	type Node struct {
		Value int
		Next  *Node
	}

	t.Run("With default max depth", func(t *testing.T) {
		gen := NewReflectDataGenerator[Node](Node{})
		node := gen.Generate()

		// Count depth
		depth := 0
		current := &node
		for current.Next != nil {
			depth++
			current = current.Next
			if depth > 10 {
				t.Fatal("Infinite recursion detected")
			}
		}

		if depth != 5 { // Default max depth is 5
			t.Errorf("Expected depth of 5, got %d", depth)
		}
	})

	t.Run("With custom max depth", func(t *testing.T) {
		customDepth := 2
		gen := NewReflectDataGenerator[Node](Node{}, WithMaxDepth[Node](customDepth))
		node := gen.Generate()

		// Count depth
		depth := 0
		current := &node
		for current.Next != nil {
			depth++
			current = current.Next
		}

		if depth != customDepth {
			t.Errorf("Expected depth of %d, got %d", customDepth, depth)
		}
	})
}

// Use the same test types as in data_type_generators_test.go for consistency
type Person struct {
	ID        string
	Name      string
	Age       int
	Email     string
	IsActive  bool
	CreatedAt time.Time
	Address   *Address
	Tags      []string
	Metadata  map[string]string
}

type Address struct {
	Street  string
	City    string
	ZipCode string
	Country string
}
