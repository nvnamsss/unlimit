package errors

// ErrorCode contains HTTP status, module and detail code.
// It is used to represent structured error codes in the format 4xxyyzz,
// where xx is the HTTP status, yy is the module, and zz is the detail code.
// Example usage:
//
//	code := GetCode(404, 2, 5)
//	errCode := NewErrorCode(code)
//	fmt.Println(errCode.Status())     // Output: 404
//	fmt.Println(errCode.Module())     // Output: 2
//	fmt.Println(errCode.DetailCode()) // Output: 5
//
// This struct is useful for categorizing and identifying errors in a consistent way.
type ErrorCode struct {
	status     int
	module     int
	detailCode int
}

// Code returns the integer with format 4xxyyzz.
// This is useful for logging or transmitting error codes in a compact form.
func (errCode ErrorCode) Code() int {
	return errCode.status*10000 + errCode.module*100 + errCode.detailCode
}

// Status returns HTTP status code.
// Useful for mapping error codes to HTTP responses.
func (errCode ErrorCode) Status() int {
	return errCode.status
}

// Module returns module error code.
// Useful for identifying which module produced the error.
func (errCode ErrorCode) Module() int {
	return errCode.module
}

// DetailCode returns detail error code.
// Useful for distinguishing specific error cases within a module.
func (errCode ErrorCode) DetailCode() int {
	return errCode.detailCode
}

// NewErrorCode constructs an ErrorCode from an integer code.
// If the code is only 3 digits (HTTP status code), it is treated as a status code and expanded.
// Example usage:
//
//	errCode := NewErrorCode(404) // status only, becomes 4040000
//	errCode := NewErrorCode(4040205) // full code with module 2, detail 5
func NewErrorCode(code int) ErrorCode {
	// If code is a standard HTTP status (< 1000), treat it as status only
	if code < 1000 {
		return ErrorCode{
			status:     code,
			module:     0,
			detailCode: 0,
		}
	}

	return ErrorCode{
		status:     code / 10000,
		module:     (code / 100) % 100,
		detailCode: code % 100,
	}
}

// GetCode builds an error code integer from status, module, and detailCode.
// Example usage:
//
//	code := GetCode(404, 2, 5) // Output: 4040205
//
// This function is useful for generating structured error codes.
func GetCode(status, module, detailCode int) int {
	return status*10000 + module*100 + detailCode
}
