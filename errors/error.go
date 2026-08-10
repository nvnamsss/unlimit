package errors

import (
	"fmt"
	"net/http"
)

// Error describes application error.
// It encapsulates a message, an optional original error, a structured error code,
// and optional metadata for additional context. This type is useful for propagating
// rich error information throughout the application, supporting error wrapping,
// custom messages, and structured error codes for logging and response handling.
//
// Example usage:
//
//	err := New(NewErrorCode(404), "resource not found")
//	fmt.Println(err.Error()) // Output: "resource not found"
//
// You can wrap an existing error:
//
//	appErr := New(NewErrorCode(500), "internal error")
//	wrapped := appErr.Wrap(errors.New("low-level failure"))
type Error struct {
	Message string `json:"message"`
	// Meta          ErrorMeta `json:"-"`
	OriginalError error                  `json:"-"`
	Code          ErrorCode              `json:"-"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Error returns the error message.
// If an OriginalError is present, its message is returned; otherwise, the Message field is used.
func (e *Error) Error() string {
	if e.OriginalError != nil {
		return e.OriginalError.Error()
	}
	return e.Message
}

// Wrap returns a new Error that wraps another error.
// The wrapped error is set as OriginalError, preserving the error code and message.
// If err is nil, returns the original Error.
func (e *Error) Wrap(err error) *Error {
	if err == nil {
		return e
	}
	return &Error{
		OriginalError: err,
		Code:          e.Code,
		Message:       e.Message,
		Metadata:      e.Metadata,
	}
}

// Unwrap returns the original wrapped error.
// This enables the error to work with errors.Is() and errors.As() from the standard library.
// Example usage:
//
//	err := errors.Wrap(sql.ErrNoRows, ErrorCodeNotFound, "user not found")
//	if errors.Is(err, sql.ErrNoRows) {
//	    // This will work correctly
//	}
//
// This method is essential for proper error chain inspection in Go 1.13+.
func (e *Error) Unwrap() error {
	return e.OriginalError
}

// Is checks if the target error matches this error based on ErrorCode.
// This allows errors with the same code to be considered equal, even if they have
// different messages or metadata.
// Example usage:
//
//	if errors.Is(err, ErrNotFound) {
//	    // Handle not found error
//	}
//
// This method enables semantic error comparison based on error codes.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}

	return e.Code.Code() == t.Code.Code()
}

// New returns an Error with the given error code and message.
// It creates a new Error instance with the specified error code and message.
// If the message is empty, it uses the HTTP status text corresponding to the error code.
// Example usage:
//
//	err := New(NewErrorCode(http.StatusNotFound), "user not found")
//	fmt.Println(err.Error()) // Output: "user not found"
//
//	err := New(NewErrorCode(404), "") // uses default HTTP status text
//	fmt.Println(err.Error()) // Output: "Not Found"
//
// This function is useful for creating standardized errors with structured error codes
// and consistent error messages throughout the application.
func New(errCode ErrorCode, msg string) error {
	if msg == "" {
		msg = http.StatusText(errCode.Status())
	}

	return &Error{
		OriginalError: nil,
		Code:          errCode,
		Message:       msg,
	}
}

// Newf returns an Error with formatted message using fmt.Sprintf.
// It creates a new Error instance with the specified error code and a formatted message.
// The message is formatted using fmt.Sprintf with the provided format string and arguments.
// Example usage:
//
//	err := Newf(NewErrorCode(400), "invalid parameter: %s", paramName)
//	fmt.Println(err.Error()) // Output: "invalid parameter: id"
//
//	err := Newf(NewErrorCode(500), "failed to process %s: %v", task, err)
//	fmt.Println(err.Error()) // Output: "failed to process task1: connection refused"
//
// This function is useful for creating errors with dynamic messages that include
// variable information while maintaining structured error codes.
func Newf(errCode ErrorCode, msg string, args ...interface{}) error {
	return &Error{
		OriginalError: nil,
		Code:          errCode,
		Message:       fmt.Sprintf(msg, args...),
	}
}

func Wrap(err error, errCode ErrorCode, msg string) error {
	if err == nil {
		return nil
	}
	return &Error{
		OriginalError: err,
		Code:          errCode,
		Message:       msg,
	}
}

func Wrapf(err error, errCode ErrorCode, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &Error{
		OriginalError: err,
		Code:          errCode,
		Message:       fmt.Sprintf(msg, args...),
	}
}
