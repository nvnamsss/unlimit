package utility

import (
	"reflect"
	"strconv"
)

func Convert[T any](v any) T {
	var target T
	targetType := reflect.TypeOf(target)

	if v == nil {
		return target
	}

	sourceValue := reflect.ValueOf(v)
	sourceType := sourceValue.Type()

	// direct type match
	if sourceType == targetType {
		return v.(T)
	}

	// Note: AssignableTo check is omitted because:
	// 1. Identical types are already handled above
	// 2. Type parameters can't easily represent interface{} or channel directions
	// 3. If source is assignable to target but not identical, Convert will be used

	// handle pointer conversions
	if targetType.Kind() == reflect.Ptr {
		return convertToPointer[T](sourceValue, targetType)
	}

	if sourceType.Kind() == reflect.Ptr {
		if sourceValue.IsNil() {
			return target
		}
		sourceValue = sourceValue.Elem()
		sourceType = sourceValue.Type()
	}

	// handle same-level conversions with acceptable data loss
	result := convertValue(sourceValue, targetType)
	if result.IsValid() && result.Type() == targetType {
		return result.Interface().(T)
	}

	return target
}

func convertToPointer[T any](sourceValue reflect.Value, targetType reflect.Type) T {
	var target T
	elemType := targetType.Elem()

	if sourceValue.Kind() == reflect.Ptr {
		if sourceValue.IsNil() {
			return target
		}
		sourceValue = sourceValue.Elem()
	}

	result := convertValue(sourceValue, elemType)
	if result.IsValid() {
		ptr := reflect.New(elemType)
		ptr.Elem().Set(result)
		return ptr.Interface().(T)
	}

	return target
}

func convertValue(sourceValue reflect.Value, targetType reflect.Type) reflect.Value {
	sourceType := sourceValue.Type()
	sourceKind := sourceType.Kind()
	targetKind := targetType.Kind()

	// string conversions should be checked before ConvertibleTo
	// because reflect considers int->string convertible (as rune conversion)
	if targetKind == reflect.String && (isNumeric(sourceKind) || sourceKind == reflect.Bool) {
		return convertToString(sourceValue, targetType)
	}

	if sourceKind == reflect.String && (isNumeric(targetKind) || targetKind == reflect.Bool) {
		return convertFromString(sourceValue.String(), targetType)
	}

	// if convertible, use direct conversion
	// Note: Go's reflect package handles all numeric type conversions (int↔uint↔float)
	// with acceptable data loss, so no custom numeric conversion logic is needed
	if sourceType.ConvertibleTo(targetType) {
		return sourceValue.Convert(targetType)
	}

	return reflect.Value{}
}

func isNumeric(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func convertFromString(str string, targetType reflect.Type) reflect.Value {
	targetKind := targetType.Kind()
	result := reflect.New(targetType).Elem()

	switch targetKind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, err := strconv.ParseInt(str, 10, 64); err == nil {
			result.SetInt(val)
			return result
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val, err := strconv.ParseUint(str, 10, 64); err == nil {
			result.SetUint(val)
			return result
		}
	case reflect.Float32, reflect.Float64:
		if val, err := strconv.ParseFloat(str, 64); err == nil {
			result.SetFloat(val)
			return result
		}
	case reflect.Bool:
		if val, err := strconv.ParseBool(str); err == nil {
			result.SetBool(val)
			return result
		}
	}

	return reflect.Value{}
}

func convertToString(sourceValue reflect.Value, targetType reflect.Type) reflect.Value {
	sourceKind := sourceValue.Kind()
	var str string

	// Note: This function is only called when sourceKind is numeric or bool (checked by caller)
	switch sourceKind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(sourceValue.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str = strconv.FormatUint(sourceValue.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(sourceValue.Float(), 'f', -1, 64)
	case reflect.Bool:
		str = strconv.FormatBool(sourceValue.Bool())
	}

	result := reflect.New(targetType).Elem()
	result.SetString(str)
	return result
}
