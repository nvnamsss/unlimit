package utility

import (
	"testing"
)

func TestConvert_DirectTypeMatch(t *testing.T) {
	val := 42
	result := Convert[int](val)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestConvert_NilValue(t *testing.T) {
	result := Convert[int](nil)
	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}
}

func TestConvert_Float64ToInt64(t *testing.T) {
	val := 42.7
	result := Convert[int64](val)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestConvert_Int64ToFloat64(t *testing.T) {
	val := int64(42)
	result := Convert[float64](val)
	if result != 42.0 {
		t.Errorf("expected 42.0, got %f", result)
	}
}

func TestConvert_IntToUint(t *testing.T) {
	val := 42
	result := Convert[uint](val)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestConvert_Float32ToInt(t *testing.T) {
	val := float32(99.9)
	result := Convert[int](val)
	if result != 99 {
		t.Errorf("expected 99, got %d", result)
	}
}

func TestConvert_StringToInt(t *testing.T) {
	val := "123"
	result := Convert[int](val)
	if result != 123 {
		t.Errorf("expected 123, got %d", result)
	}
}

func TestConvert_StringToFloat(t *testing.T) {
	val := "123.45"
	result := Convert[float64](val)
	if result != 123.45 {
		t.Errorf("expected 123.45, got %f", result)
	}
}

func TestConvert_IntToString(t *testing.T) {
	val := 123
	result := Convert[string](val)
	if result != "123" {
		t.Errorf("expected '123', got '%s'", result)
	}
}

func TestConvert_FloatToString(t *testing.T) {
	val := 123.45
	result := Convert[string](val)
	if result != "123.45" {
		t.Errorf("expected '123.45', got '%s'", result)
	}
}

func TestConvert_StringToBool(t *testing.T) {
	val := "true"
	result := Convert[bool](val)
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestConvert_BoolToString(t *testing.T) {
	val := true
	result := Convert[string](val)
	if result != "true" {
		t.Errorf("expected 'true', got '%s'", result)
	}
}

func TestConvert_PointerToValue(t *testing.T) {
	val := 42
	ptr := &val
	result := Convert[int](ptr)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestConvert_ValueToPointer(t *testing.T) {
	val := 42
	result := Convert[*int](val)
	if result == nil || *result != 42 {
		t.Errorf("expected pointer to 42, got %v", result)
	}
}

func TestConvert_NilPointer(t *testing.T) {
	var ptr *int
	result := Convert[int](ptr)
	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}
}

func TestConvert_InvalidStringToInt(t *testing.T) {
	val := "invalid"
	result := Convert[int](val)
	if result != 0 {
		t.Errorf("expected 0 for invalid conversion, got %d", result)
	}
}

func TestConvert_UintToInt(t *testing.T) {
	val := uint(42)
	result := Convert[int](val)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestConvert_Int8ToInt64(t *testing.T) {
	val := int8(42)
	result := Convert[int64](val)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestConvert_Float64ToInt32(t *testing.T) {
	val := 42.9
	result := Convert[int32](val)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

// Additional comprehensive tests

func TestConvert_NegativeIntToUint(t *testing.T) {
	val := -42
	result := Convert[uint](val)
	// This will wrap around due to data loss
	if result == 0 {
		t.Errorf("expected non-zero for negative to uint conversion")
	}
}

func TestConvert_LargeFloatToInt(t *testing.T) {
	val := 9999999999.99
	result := Convert[int64](val)
	if result != 9999999999 {
		t.Errorf("expected 9999999999, got %d", result)
	}
}

func TestConvert_Int16ToInt32(t *testing.T) {
	val := int16(1234)
	result := Convert[int32](val)
	if result != 1234 {
		t.Errorf("expected 1234, got %d", result)
	}
}

func TestConvert_Uint8ToUint64(t *testing.T) {
	val := uint8(255)
	result := Convert[uint64](val)
	if result != 255 {
		t.Errorf("expected 255, got %d", result)
	}
}

func TestConvert_Float32ToFloat64(t *testing.T) {
	val := float32(3.14)
	result := Convert[float64](val)
	if result < 3.13 || result > 3.15 {
		t.Errorf("expected ~3.14, got %f", result)
	}
}

func TestConvert_Float64ToFloat32(t *testing.T) {
	val := 3.14159
	result := Convert[float32](val)
	if result < 3.14 || result > 3.15 {
		t.Errorf("expected ~3.14, got %f", result)
	}
}

func TestConvert_Uint64ToInt64(t *testing.T) {
	val := uint64(12345)
	result := Convert[int64](val)
	if result != 12345 {
		t.Errorf("expected 12345, got %d", result)
	}
}

func TestConvert_Int32ToUint32(t *testing.T) {
	val := int32(9999)
	result := Convert[uint32](val)
	if result != 9999 {
		t.Errorf("expected 9999, got %d", result)
	}
}

func TestConvert_EmptyStringToInt(t *testing.T) {
	val := ""
	result := Convert[int](val)
	if result != 0 {
		t.Errorf("expected 0 for empty string, got %d", result)
	}
}

func TestConvert_WhitespaceStringToInt(t *testing.T) {
	val := "  "
	result := Convert[int](val)
	if result != 0 {
		t.Errorf("expected 0 for whitespace string, got %d", result)
	}
}

func TestConvert_StringToUint(t *testing.T) {
	val := "789"
	result := Convert[uint](val)
	if result != 789 {
		t.Errorf("expected 789, got %d", result)
	}
}

func TestConvert_StringToFloat32(t *testing.T) {
	val := "3.14"
	result := Convert[float32](val)
	if result < 3.13 || result > 3.15 {
		t.Errorf("expected ~3.14, got %f", result)
	}
}

func TestConvert_UintToString(t *testing.T) {
	val := uint(456)
	result := Convert[string](val)
	if result != "456" {
		t.Errorf("expected '456', got '%s'", result)
	}
}

func TestConvert_Float32ToString(t *testing.T) {
	val := float32(9.5)
	result := Convert[string](val)
	if result != "9.5" {
		t.Errorf("expected '9.5', got '%s'", result)
	}
}

func TestConvert_BoolFalseToString(t *testing.T) {
	val := false
	result := Convert[string](val)
	if result != "false" {
		t.Errorf("expected 'false', got '%s'", result)
	}
}

func TestConvert_StringToBoolFalse(t *testing.T) {
	val := "false"
	result := Convert[bool](val)
	if result != false {
		t.Errorf("expected false, got %v", result)
	}
}

func TestConvert_StringToBool_One(t *testing.T) {
	val := "1"
	result := Convert[bool](val)
	if result != true {
		t.Errorf("expected true for '1', got %v", result)
	}
}

func TestConvert_StringToBool_Zero(t *testing.T) {
	val := "0"
	result := Convert[bool](val)
	if result != false {
		t.Errorf("expected false for '0', got %v", result)
	}
}

func TestConvert_InvalidStringToBool(t *testing.T) {
	val := "invalid"
	result := Convert[bool](val)
	if result != false {
		t.Errorf("expected false for invalid string, got %v", result)
	}
}

func TestConvert_PointerToPointer(t *testing.T) {
	val := 99
	ptr := &val
	result := Convert[*int](ptr)
	if result == nil || *result != 99 {
		t.Errorf("expected pointer to 99, got %v", result)
	}
}

func TestConvert_NilPointerToPointer(t *testing.T) {
	var ptr *int
	result := Convert[*int](ptr)
	if result != nil {
		t.Errorf("expected nil pointer, got %v", result)
	}
}

func TestConvert_FloatPointerToValue(t *testing.T) {
	val := 3.14
	ptr := &val
	result := Convert[float64](ptr)
	if result != 3.14 {
		t.Errorf("expected 3.14, got %f", result)
	}
}

func TestConvert_IntPointerToStringPointer(t *testing.T) {
	val := 42
	ptr := &val
	result := Convert[*string](ptr)
	if result == nil || *result != "42" {
		t.Errorf("expected pointer to '42', got %v", result)
	}
}

func TestConvert_AssignableTypes(t *testing.T) {
	type MyInt int
	val := MyInt(100)
	result := Convert[int](val)
	if result != 100 {
		t.Errorf("expected 100, got %d", result)
	}
}

func TestConvert_CustomTypeToCustomType(t *testing.T) {
	type MyInt int
	type OtherInt int
	val := MyInt(42)
	result := Convert[OtherInt](val)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestConvert_Int8ToUint8(t *testing.T) {
	val := int8(127)
	result := Convert[uint8](val)
	if result != 127 {
		t.Errorf("expected 127, got %d", result)
	}
}

func TestConvert_Uint16ToInt8(t *testing.T) {
	val := uint16(50)
	result := Convert[int8](val)
	if result != 50 {
		t.Errorf("expected 50, got %d", result)
	}
}

func TestConvert_NegativeFloatToInt(t *testing.T) {
	val := -42.7
	result := Convert[int](val)
	if result != -42 {
		t.Errorf("expected -42, got %d", result)
	}
}

func TestConvert_NegativeStringToInt(t *testing.T) {
	val := "-999"
	result := Convert[int](val)
	if result != -999 {
		t.Errorf("expected -999, got %d", result)
	}
}

func TestConvert_StringWithPlusToInt(t *testing.T) {
	val := "+123"
	result := Convert[int](val)
	if result != 123 {
		t.Errorf("expected 123, got %d", result)
	}
}

func TestConvert_ScientificNotationString(t *testing.T) {
	val := "1.23e2"
	result := Convert[float64](val)
	if result != 123.0 {
		t.Errorf("expected 123.0, got %f", result)
	}
}

func TestConvert_NegativeIntToString(t *testing.T) {
	val := -456
	result := Convert[string](val)
	if result != "-456" {
		t.Errorf("expected '-456', got '%s'", result)
	}
}

func TestConvert_NegativeFloatToString(t *testing.T) {
	val := -3.14
	result := Convert[string](val)
	if result != "-3.14" {
		t.Errorf("expected '-3.14', got '%s'", result)
	}
}

func TestConvert_ZeroValue(t *testing.T) {
	val := 0
	result := Convert[string](val)
	if result != "0" {
		t.Errorf("expected '0', got '%s'", result)
	}
}

func TestConvert_UnsupportedConversion(t *testing.T) {
	type MyStruct struct {
		Value int
	}
	val := MyStruct{Value: 42}
	result := Convert[string](val)
	if result != "" {
		t.Errorf("expected empty string for unsupported conversion, got '%s'", result)
	}
}

func TestConvert_StructToInt(t *testing.T) {
	type MyStruct struct {
		Value int
	}
	val := MyStruct{Value: 42}
	result := Convert[int](val)
	if result != 0 {
		t.Errorf("expected 0 for unsupported conversion, got %d", result)
	}
}

func TestConvert_SliceToString(t *testing.T) {
	val := []int{1, 2, 3}
	result := Convert[string](val)
	if result != "" {
		t.Errorf("expected empty string for slice conversion, got '%s'", result)
	}
}

func TestConvert_MapToInt(t *testing.T) {
	val := map[string]int{"key": 42}
	result := Convert[int](val)
	if result != 0 {
		t.Errorf("expected 0 for map conversion, got %d", result)
	}
}

func TestConvert_StringToUint64(t *testing.T) {
	val := "18446744073709551615"
	result := Convert[uint64](val)
	if result != 18446744073709551615 {
		t.Errorf("expected max uint64, got %d", result)
	}
}

func TestConvert_InvalidStringToFloat(t *testing.T) {
	val := "not-a-number"
	result := Convert[float64](val)
	if result != 0.0 {
		t.Errorf("expected 0.0 for invalid string, got %f", result)
	}
}

func TestConvert_InvalidStringToUint(t *testing.T) {
	val := "invalid"
	result := Convert[uint](val)
	if result != 0 {
		t.Errorf("expected 0 for invalid string, got %d", result)
	}
}

func TestConvert_StringToInt8(t *testing.T) {
	val := "127"
	result := Convert[int8](val)
	if result != 127 {
		t.Errorf("expected 127, got %d", result)
	}
}

func TestConvert_StringToUint8(t *testing.T) {
	val := "255"
	result := Convert[uint8](val)
	if result != 255 {
		t.Errorf("expected 255, got %d", result)
	}
}

func TestConvert_StringToInt16(t *testing.T) {
	val := "32767"
	result := Convert[int16](val)
	if result != 32767 {
		t.Errorf("expected 32767, got %d", result)
	}
}

func TestConvert_StringToInt32(t *testing.T) {
	val := "2147483647"
	result := Convert[int32](val)
	if result != 2147483647 {
		t.Errorf("expected 2147483647, got %d", result)
	}
}

func TestConvert_StringToInt64(t *testing.T) {
	val := "9223372036854775807"
	result := Convert[int64](val)
	if result != 9223372036854775807 {
		t.Errorf("expected max int64, got %d", result)
	}
}

func TestConvert_Int64ToString(t *testing.T) {
	val := int64(9223372036854775807)
	result := Convert[string](val)
	if result != "9223372036854775807" {
		t.Errorf("expected '9223372036854775807', got '%s'", result)
	}
}

func TestConvert_Uint64ToString(t *testing.T) {
	val := uint64(18446744073709551615)
	result := Convert[string](val)
	if result != "18446744073709551615" {
		t.Errorf("expected '18446744073709551615', got '%s'", result)
	}
}

func TestConvert_Int8ToString(t *testing.T) {
	val := int8(-128)
	result := Convert[string](val)
	if result != "-128" {
		t.Errorf("expected '-128', got '%s'", result)
	}
}

func TestConvert_Uint8ToString(t *testing.T) {
	val := uint8(255)
	result := Convert[string](val)
	if result != "255" {
		t.Errorf("expected '255', got '%s'", result)
	}
}

func TestConvert_PointerToString(t *testing.T) {
	val := 789
	ptr := &val
	result := Convert[string](ptr)
	if result != "789" {
		t.Errorf("expected '789', got '%s'", result)
	}
}

func TestConvert_NilPointerToString(t *testing.T) {
	var ptr *int
	result := Convert[string](ptr)
	if result != "" {
		t.Errorf("expected empty string for nil pointer, got '%s'", result)
	}
}

func TestConvert_BoolPointerToValue(t *testing.T) {
	val := true
	ptr := &val
	result := Convert[bool](ptr)
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

// Edge cases for 100% coverage

func TestConvert_InvalidResultType(t *testing.T) {
	// Test case where convertValue returns a value but with wrong type
	// This should trigger the result.Type() != targetType check
	type MyCustomType struct {
		Value int
	}
	val := MyCustomType{Value: 42}
	result := Convert[int](val)
	if result != 0 {
		t.Errorf("expected 0 for invalid conversion, got %d", result)
	}
}

func TestConvert_PointerToInvalidType(t *testing.T) {
	// Test convertToPointer where result is not valid
	type MyStruct struct {
		Value int
	}
	val := MyStruct{Value: 42}
	result := Convert[*int](val)
	if result != nil {
		t.Errorf("expected nil for invalid pointer conversion, got %v", result)
	}
}

func TestConvert_ChannelToString(t *testing.T) {
	// Test convertToString with unsupported type (channel)
	ch := make(chan int)
	result := Convert[string](ch)
	if result != "" {
		t.Errorf("expected empty string for channel conversion, got '%s'", result)
	}
}

func TestConvert_FuncToString(t *testing.T) {
	// Test convertToString with unsupported type (function)
	fn := func() int { return 42 }
	result := Convert[string](fn)
	if result != "" {
		t.Errorf("expected empty string for function conversion, got '%s'", result)
	}
}

func TestConvert_InvalidPointerConversion(t *testing.T) {
	// Test pointer conversion where the element type can't be converted
	type MyStruct struct {
		Value int
	}
	val := &MyStruct{Value: 42}
	result := Convert[*string](val)
	if result != nil {
		t.Errorf("expected nil for invalid pointer conversion, got %v", result)
	}
}

func TestConvert_AssignableInterface(t *testing.T) {
	// Test the AssignableTo path with interface
	type Stringer interface {
		String() string
	}

	type MyString string

	// Create a value that implements Stringer
	val := MyString("test")

	// Try to convert - this will test the path but won't actually
	// hit AssignableTo since MyString doesn't implement Stringer
	result := Convert[string](val)
	if result != "test" {
		t.Errorf("expected 'test', got '%s'", result)
	}
}

func TestConvert_AssignableToSameBaseType(t *testing.T) {
	// The AssignableTo line is primarily for interface conversions
	// but since we can't use interface{} as a type parameter easily,
	// we test with type that goes through ConvertibleTo instead
	type StringAlias string
	var val StringAlias = "hello"

	result := Convert[string](val)
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestConvert_NilPointerToPointerDifferentPath(t *testing.T) {
	// Test nil pointer to pointer of different type
	// This hits line 56-58 in convertToPointer
	var nilPtr *int = nil
	result := Convert[*string](nilPtr)
	if result != nil {
		t.Errorf("expected nil for nil pointer conversion, got %v", result)
	}
}
