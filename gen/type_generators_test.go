package gen

import (
	"reflect"
	"testing"
	"unicode"
)

func TestBasePrimitiveGenerator_New(t *testing.T) {
	generator := NewBasePrimitiveGenerator()
	if generator == nil {
		t.Error("NewBasePrimitiveGenerator() returned nil")
	}

}

func TestIntGenerator_New(t *testing.T) {
	// Test for different integer types
	tests := []struct {
		name string
		want interface{}
	}{
		{"int", NewIntGenerator[int]()},
		{"int8", NewIntGenerator[int8]()},
		{"int16", NewIntGenerator[int16]()},
		{"int32", NewIntGenerator[int32]()},
		{"int64", NewIntGenerator[int64]()},
		{"uint", NewIntGenerator[uint]()},
		{"uint8", NewIntGenerator[uint8]()},
		{"uint16", NewIntGenerator[uint16]()},
		{"uint32", NewIntGenerator[uint32]()},
		{"uint64", NewIntGenerator[uint64]()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want == nil {
				t.Errorf("NewIntGenerator[%s]() returned nil", tt.name)
			}
		})
	}
}

func TestIntGenerator_Generate(t *testing.T) {
	// Test generation for different integer types
	testCases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "int",
			testFunc: func(t *testing.T) {
				generator := NewIntGenerator[int]()
				val := generator.Generate()
				if reflect.TypeOf(val).Kind() != reflect.Int {
					t.Errorf("Expected type int, got %T", val)
				}
			},
		},
		{
			name: "int8",
			testFunc: func(t *testing.T) {
				generator := NewIntGenerator[int8]()
				val := generator.Generate()
				if reflect.TypeOf(val).Kind() != reflect.Int8 {
					t.Errorf("Expected type int8, got %T", val)
				}
			},
		},
		{
			name: "int16",
			testFunc: func(t *testing.T) {
				generator := NewIntGenerator[int16]()
				val := generator.Generate()
				if reflect.TypeOf(val).Kind() != reflect.Int16 {
					t.Errorf("Expected type int16, got %T", val)
				}
			},
		},
		{
			name: "int32",
			testFunc: func(t *testing.T) {
				generator := NewIntGenerator[int32]()
				val := generator.Generate()
				if reflect.TypeOf(val).Kind() != reflect.Int32 {
					t.Errorf("Expected type int32, got %T", val)
				}
			},
		},
		{
			name: "int64",
			testFunc: func(t *testing.T) {
				generator := NewIntGenerator[int64]()
				val := generator.Generate()
				if reflect.TypeOf(val).Kind() != reflect.Int64 {
					t.Errorf("Expected type int64, got %T", val)
				}
			},
		},
		{
			name: "uint",
			testFunc: func(t *testing.T) {
				generator := NewIntGenerator[uint]()
				val := generator.Generate()
				if reflect.TypeOf(val).Kind() != reflect.Uint {
					t.Errorf("Expected type uint, got %T", val)
				}
			},
		},
		{
			name: "uint8",
			testFunc: func(t *testing.T) {
				generator := NewIntGenerator[uint8]()
				val := generator.Generate()
				if reflect.TypeOf(val).Kind() != reflect.Uint8 {
					t.Errorf("Expected type uint8, got %T", val)
				}
			},
		},
		{
			name: "uint16",
			testFunc: func(t *testing.T) {
				generator := NewIntGenerator[uint16]()
				val := generator.Generate()
				if reflect.TypeOf(val).Kind() != reflect.Uint16 {
					t.Errorf("Expected type uint16, got %T", val)
				}
			},
		},
		{
			name: "uint32",
			testFunc: func(t *testing.T) {
				generator := NewIntGenerator[uint32]()
				val := generator.Generate()
				if reflect.TypeOf(val).Kind() != reflect.Uint32 {
					t.Errorf("Expected type uint32, got %T", val)
				}
			},
		},
		{
			name: "uint64",
			testFunc: func(t *testing.T) {
				generator := NewIntGenerator[uint64]()
				val := generator.Generate()
				if reflect.TypeOf(val).Kind() != reflect.Uint64 {
					t.Errorf("Expected type uint64, got %T", val)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.testFunc)
	}
}

func TestIntGenerator_Distribution(t *testing.T) {
	// Test distribution of generated values for signed types
	generator := NewIntGenerator[int]()

	positive := 0
	negative := 0
	zero := 0

	// Generate enough samples to test distribution
	for i := 0; i < 1000; i++ {
		val := generator.Generate()
		if val > 0 {
			positive++
		} else if val < 0 {
			negative++
		} else {
			zero++
		}
	}

	// Verify we get a mix of positive and negative values
	if positive == 0 || negative == 0 {
		t.Error("IntGenerator should generate both positive and negative values")
	}
}

func TestFloatGenerator_New(t *testing.T) {
	// Test for float32 and float64
	tests := []struct {
		name string
		want interface{}
	}{
		{"float32", NewFloatGenerator[float32]()},
		{"float64", NewFloatGenerator[float64]()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want == nil {
				t.Errorf("NewFloatGenerator[%s]() returned nil", tt.name)
			}
		})
	}
}

func TestFloatGenerator_Generate(t *testing.T) {
	// Test generation for float32 and float64
	t.Run("float32", func(t *testing.T) {
		generator := NewFloatGenerator[float32]()
		val := generator.Generate()
		if reflect.TypeOf(val).Kind() != reflect.Float32 {
			t.Errorf("Expected type float32, got %T", val)
		}
	})

	t.Run("float64", func(t *testing.T) {
		generator := NewFloatGenerator[float64]()
		val := generator.Generate()
		if reflect.TypeOf(val).Kind() != reflect.Float64 {
			t.Errorf("Expected type float64, got %T", val)
		}
	})
}

func TestFloatGenerator_Distribution(t *testing.T) {
	// Test distribution of generated values
	generator := NewFloatGenerator[float64]()

	positive := 0
	negative := 0
	zero := 0

	// Generate enough samples to test distribution
	for i := 0; i < 1000; i++ {
		val := generator.Generate()
		if val > 0 {
			positive++
		} else if val < 0 {
			negative++
		} else {
			zero++
		}
	}

	// Verify we get a mix of positive and negative values
	if positive == 0 || negative == 0 {
		t.Error("FloatGenerator should generate both positive and negative values")
	}
}

func TestStringGenerator_New(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   bool
	}{
		{"default length", 0, true},
		{"custom length", 16, true},
		{"negative length", -5, true}, // Should use default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewStringGenerator(tt.length)
			if (generator != nil) != tt.want {
				t.Errorf("NewStringGenerator(%v) returned %v, want %v", tt.length, generator != nil, tt.want)
			}
		})
	}
}

func TestStringGenerator_Generate(t *testing.T) {
	t.Run("default length", func(t *testing.T) {
		generator := NewStringGenerator(0)
		val := generator.Generate()
		if len(val) != 8 {
			t.Errorf("Expected string of length 8, got %d", len(val))
		}
		if reflect.TypeOf(val).Kind() != reflect.String {
			t.Errorf("Expected type string, got %T", val)
		}
	})

	t.Run("custom length", func(t *testing.T) {
		length := 16
		generator := NewStringGenerator(length)
		val := generator.Generate()
		if len(val) != length {
			t.Errorf("Expected string of length %d, got %d", length, len(val))
		}
	})
}

func TestStringGenerator_CharacterSet(t *testing.T) {
	generator := NewStringGenerator(100) // Long enough to likely include all character types
	val := generator.Generate()

	hasLower := false
	hasUpper := false
	hasDigit := false

	for _, r := range val {
		if unicode.IsLower(r) {
			hasLower = true
		} else if unicode.IsUpper(r) {
			hasUpper = true
		} else if unicode.IsDigit(r) {
			hasDigit = true
		}
	}

	if !hasLower || !hasUpper || !hasDigit {
		t.Error("StringGenerator should generate strings with lowercase, uppercase and digits")
	}
}

func TestMultipleGenerations(t *testing.T) {
	// Test that consecutive generations produce different values

	t.Run("IntGenerator_MultipleGenerate", func(t *testing.T) {
		generator := NewIntGenerator[int]()
		v1 := generator.Generate()
		v2 := generator.Generate()

		// With enough samples, we should eventually get different values
		different := false
		for i := 0; i < 10 && !different; i++ {
			v3 := generator.Generate()
			if v3 != v1 && v3 != v2 {
				different = true
			}
		}

		if !different {
			t.Error("IntGenerator should generate different values across calls")
		}
	})

	t.Run("FloatGenerator_MultipleGenerate", func(t *testing.T) {
		generator := NewFloatGenerator[float64]()
		v1 := generator.Generate()
		v2 := generator.Generate()

		// With floating-point, there's an extremely high probability of different values
		if v1 == v2 {
			t.Error("FloatGenerator should generate different values across calls")
		}
	})

	t.Run("StringGenerator_MultipleGenerate", func(t *testing.T) {
		generator := NewStringGenerator(8)
		v1 := generator.Generate()
		v2 := generator.Generate()

		// With 8-character strings, there's an extremely high probability of different values
		if v1 == v2 {
			t.Error("StringGenerator should generate different values across calls")
		}
	})
}
