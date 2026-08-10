package utility

// GetPointer returns a pointer to the given value of any type
func GetPointer[T any](v T) *T {
	return &v
}

// GetPointerString returns a pointer to the given string value
func GetPointerString(v string) *string {
	return &v
}

// GetPointerInt returns a pointer to the given int value
func GetPointerInt(v int) *int {
	return &v
}

// GetPointerInt32 returns a pointer to the given int32 value
func GetPointerInt32(v int32) *int32 {
	return &v
}

// GetPointerInt64 returns a pointer to the given int64 value
func GetPointerInt64(v int64) *int64 {
	return &v
}

// GetPointerBool returns a pointer to the given bool value
func GetPointerBool(v bool) *bool {
	return &v
}

// GetPointerFloat32 returns a pointer to the given float32 value
func GetPointerFloat32(v float32) *float32 {
	return &v
}

// GetPointerFloat64 returns a pointer to the given float64 value
func GetPointerFloat64(v float64) *float64 {
	return &v
}

// GetPointerByte returns a pointer to the given byte value
func GetPointerByte(v byte) *byte {
	return &v
}

// GetPointerRune returns a pointer to the given rune value
func GetPointerRune(v rune) *rune {
	return &v
}
