package errors

import (
	"testing"
)

func TestErrorCode_NewErrorCode(t *testing.T) {
	code := GetCode(404, 2, 5)
	errCode := NewErrorCode(code)
	if errCode.Status() != 404 {
		t.Errorf("Expected status 404, got %d", errCode.Status())
	}
	if errCode.Module() != 2 {
		t.Errorf("Expected module 2, got %d", errCode.Module())
	}
	if errCode.DetailCode() != 5 {
		t.Errorf("Expected detailCode 5, got %d", errCode.DetailCode())
	}
}

func TestErrorCode_CodeMethod(t *testing.T) {
	code := GetCode(400, 1, 99)
	errCode := NewErrorCode(code)
	expected := 4000199
	if errCode.Code() != expected {
		t.Errorf("Expected code %d, got %d", expected, errCode.Code())
	}
}

func TestErrorCode_GetCode(t *testing.T) {
	status, module, detail := 500, 10, 42
	code := GetCode(status, module, detail)
	expected := 5001042
	if code != expected {
		t.Errorf("Expected %d, got %d", expected, code)
	}
}

func TestErrorCode_ThreeDigitStatus(t *testing.T) {
	// If code is only 3 digits, NewErrorCode should treat it as status only
	code := 404
	errCode := NewErrorCode(code)
	if errCode.Status() != 404 {
		t.Errorf("Expected status 404, got %d", errCode.Status())
	}
	if errCode.Module() != 0 {
		t.Errorf("Expected module 0, got %d", errCode.Module())
	}
	if errCode.DetailCode() != 0 {
		t.Errorf("Expected detailCode 0, got %d", errCode.DetailCode())
	}
}

func TestErrorCode_BoundaryValues(t *testing.T) {
	// Test with max values for module and detailCode
	code := GetCode(200, 99, 99)
	errCode := NewErrorCode(code)
	if errCode.Module() != 99 {
		t.Errorf("Expected module 99, got %d", errCode.Module())
	}
	if errCode.DetailCode() != 99 {
		t.Errorf("Expected detailCode 99, got %d", errCode.DetailCode())
	}
}
