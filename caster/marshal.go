package caster

import "encoding/json"

// MarshalCast converts a value of any type to a target type T using JSON marshaling and unmarshaling.
// It is a generic function that takes a source value of any type and attempts to convert it
// to the target type T by first marshaling to JSON, then unmarshaling to the target type.
// Example usage:
//
//	type User struct {
//		Name string `json:"name"`
//		Age  int    `json:"age"`
//	}
//	data := map[string]interface{}{"name": "Alice", "age": 30}
//	user, err := MarshalCast[User](data)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(user.Name) // Output: "Alice"
//
// This function is useful for converting between different data structures that have
// compatible JSON representations, such as converting maps to structs or between different struct types.
// Note that this conversion may lose type information or fail if the JSON representation
// is not compatible between the source and target types.
func MarshalCast[T any](src any) (T, error) {
	var result T
	bytes, err := json.Marshal(src)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(bytes, &result)
	return result, err
}
