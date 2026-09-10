package utility_test

import (
	"fmt"

	"github.com/nvnamsss/unlimit/utility"
)

func ExampleConvert_float64ToInt() {
	val := 42.9
	result := utility.Convert[int](val)
	fmt.Println(result)
	// Output: 42
}

func ExampleConvert_intToString() {
	val := 123
	result := utility.Convert[string](val)
	fmt.Println(result)
	// Output: 123
}

func ExampleConvert_stringToInt() {
	val := "456"
	result := utility.Convert[int](val)
	fmt.Println(result)
	// Output: 456
}

func ExampleConvert_pointerConversion() {
	val := 42
	ptr := utility.Convert[*int](val)
	fmt.Println(*ptr)
	// Output: 42
}

func ExampleConvert_numericTypes() {
	// float64 to int64 with data loss (truncation)
	f := 99.99
	i := utility.Convert[int64](f)
	fmt.Println(i)

	// int32 to uint64
	i32 := int32(100)
	u64 := utility.Convert[uint64](i32)
	fmt.Println(u64)

	// Output:
	// 99
	// 100
}
