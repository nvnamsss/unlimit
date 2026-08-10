package errors

import (
	"database/sql"
	stderr "errors"
	"io"
	"testing"
)

func TestError_New(t *testing.T) {
	errorCode := NewErrorCode(GetCode(404, 0, 0))
	err := New(errorCode, "not found")
	appErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("Expected *Error, got %T", err)
	}

	if appErr.Code != errorCode {
		t.Errorf("Expected ErrorCode %v, got %v", errorCode, appErr.Code)
	}

	if appErr.Message != "not found" {
		t.Errorf("Expected message 'not found', got '%s'", appErr.Message)
	}
}

func TestError_Newf(t *testing.T) {
	code := NewErrorCode(GetCode(400, 0, 0))
	err := Newf(code, "bad request: %s", "invalid")
	appErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("Expected *Error, got %T", err)
	}
	if appErr.Code != code {
		t.Errorf("Expected ErrorCode %v, got %v", code, appErr.Code)
	}
	expectedMsg := "bad request: invalid"
	if appErr.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, appErr.Message)
	}
}

func TestError_ErrorMethod(t *testing.T) {
	appErr := &Error{Message: "msg"}
	if appErr.Error() != "msg" {
		t.Errorf("Expected Error() to return 'msg', got '%s'", appErr.Error())
	}
	orig := stderr.New("original")
	appErr.OriginalError = orig
	if appErr.Error() != "original" {
		t.Errorf("Expected Error() to return original error message, got '%s'", appErr.Error())
	}
}

func TestError_Wrap(t *testing.T) {
	appErr := &Error{Message: "wrap", Code: NewErrorCode(500)}
	wrapped := appErr.Wrap(stderr.New("wrapped error"))
	if wrapped.OriginalError.Error() != "wrapped error" {
		t.Errorf("Expected wrapped error message, got '%s'", wrapped.OriginalError.Error())
	}
	if wrapped.Code != appErr.Code {
		t.Errorf("Expected ErrorCode %v, got %v", appErr.Code, wrapped.Code)
	}
	if wrapped.Message != appErr.Message {
		t.Errorf("Expected message '%s', got '%s'", appErr.Message, wrapped.Message)
	}
}

func TestError_WrapNil(t *testing.T) {
	appErr := &Error{Message: "wrapnil"}
	wrapped := appErr.Wrap(nil)
	if wrapped != appErr {
		t.Errorf("Expected Wrap(nil) to return original error")
	}
}

func TestError_Metadata(t *testing.T) {
	meta := map[string]interface{}{"foo": "bar"}
	appErr := &Error{Message: "meta", Metadata: meta}
	if appErr.Metadata["foo"] != "bar" {
		t.Errorf("Expected metadata 'bar', got '%v'", appErr.Metadata["foo"])
	}
}

// TestError_Unwrap tests that Unwrap() correctly returns the original error
func TestError_Unwrap(t *testing.T) {
	originalErr := stderr.New("database connection failed")
	appErr := &Error{
		Message:       "failed to query user",
		Code:          NewErrorCode(500),
		OriginalError: originalErr,
	}

	unwrapped := appErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Expected Unwrap() to return original error, got %v", unwrapped)
	}
}

// TestError_UnwrapNil tests that Unwrap() returns nil when there's no original error
func TestError_UnwrapNil(t *testing.T) {
	appErr := &Error{
		Message: "validation failed",
		Code:    NewErrorCode(400),
	}

	unwrapped := appErr.Unwrap()
	if unwrapped != nil {
		t.Errorf("Expected Unwrap() to return nil, got %v", unwrapped)
	}
}

// TestError_ErrorsIs tests that errors.Is() works correctly with our Error type
func TestError_ErrorsIs(t *testing.T) {
	// Test with same error code
	err1 := New(ErrorCodeNotFound, "user not found")
	err2 := New(ErrorCodeNotFound, "resource not found")

	if !stderr.Is(err1, err2) {
		t.Error("Expected errors.Is() to return true for errors with same code")
	}

	// Test with different error codes
	err3 := New(ErrorCodeInvalidRequest, "bad input")
	if stderr.Is(err1, err3) {
		t.Error("Expected errors.Is() to return false for errors with different codes")
	}

	// Test with wrapped errors
	originalErr := io.EOF
	wrappedErr := Wrap(originalErr, ErrorCodeInternalServer, "read failed")

	if !stderr.Is(wrappedErr, originalErr) {
		t.Error("Expected errors.Is() to find wrapped io.EOF")
	}

	// Test with sql.ErrNoRows
	dbErr := Wrap(sql.ErrNoRows, ErrorCodeNotFound, "user not found")
	if !stderr.Is(dbErr, sql.ErrNoRows) {
		t.Error("Expected errors.Is() to find wrapped sql.ErrNoRows")
	}
}

// TestError_ErrorsAs tests that errors.As() works correctly with our Error type
func TestError_ErrorsAs(t *testing.T) {
	// Test unwrapping to our Error type
	err := New(ErrorCodeUnauthorized, "unauthorized access")
	var appErr *Error
	if !stderr.As(err, &appErr) {
		t.Error("Expected errors.As() to extract *Error")
	}
	if appErr.Code.Status() != 401 {
		t.Errorf("Expected status 401, got %d", appErr.Code.Status())
	}

	// Test with wrapped error - extract our Error type
	wrappedErr := Wrap(io.EOF, ErrorCodeInternalServer, "read failed")
	var extractedErr *Error
	if !stderr.As(wrappedErr, &extractedErr) {
		t.Error("Expected errors.As() to extract *Error from wrapped error")
	}
	if extractedErr.Code.Status() != 500 {
		t.Errorf("Expected status 500, got %d", extractedErr.Code.Status())
	}

	// Test extracting the original error through the chain
	originalErr := &customError{msg: "custom"}
	chainedErr := Wrap(originalErr, ErrorCodeInternalServer, "operation failed")
	var custom *customError
	if !stderr.As(chainedErr, &custom) {
		t.Error("Expected errors.As() to extract customError through chain")
	}
	if custom.msg != "custom" {
		t.Errorf("Expected custom message 'custom', got '%s'", custom.msg)
	}
}

// TestError_ErrorsIsWithInterface tests errors.Is() when working through error interface
func TestError_ErrorsIsWithInterface(t *testing.T) {
	// Create errors as error interface
	var err1 error = New(ErrorCodeNotFound, "not found")
	var err2 error = New(ErrorCodeNotFound, "different message")

	// Should match based on error code
	if !stderr.Is(err1, err2) {
		t.Error("Expected errors.Is() to work with error interface and same code")
	}

	// Test with predefined errors
	var userErr error = ErrNotFound
	var customErr error = New(ErrorCodeNotFound, "custom not found")

	if !stderr.Is(userErr, customErr) {
		t.Error("Expected errors.Is() to match predefined error with custom error of same code")
	}
}

// TestError_ErrorsAsWithInterface tests errors.As() when working through error interface
func TestError_ErrorsAsWithInterface(t *testing.T) {
	// Wrap a standard library error
	var err error = Wrap(sql.ErrNoRows, ErrorCodeNotFound, "user not found")

	// Extract our Error type
	var appErr *Error
	if !stderr.As(err, &appErr) {
		t.Fatal("Expected errors.As() to extract *Error from error interface")
	}

	if appErr.Code.Status() != 404 {
		t.Errorf("Expected status 404, got %d", appErr.Code.Status())
	}
	if appErr.Message != "user not found" {
		t.Errorf("Expected message 'user not found', got '%s'", appErr.Message)
	}

	// Verify we can still find the original error
	if !stderr.Is(err, sql.ErrNoRows) {
		t.Error("Expected to find sql.ErrNoRows in error chain")
	}
}

// customError is a helper type for testing errors.As()
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}
