package errors

import "net/http"

// Auth module error codes
const (
	ModuleAuth   = 1 // Auth module identifier
	ModuleCommon = 0 // Common module identifier
)

var (
	ErrorCodeInvalidRequest = NewErrorCode(GetCode(http.StatusBadRequest, ModuleCommon, 1))
	ErrorCodeUnauthorized   = NewErrorCode(GetCode(http.StatusUnauthorized, ModuleCommon, 1))
	ErrorCodeInternalServer = NewErrorCode(GetCode(http.StatusInternalServerError, ModuleCommon, 1))
	ErrorCodeNoResponse     = NewErrorCode(GetCode(http.StatusInternalServerError, ModuleCommon, 2))
	ErrorCodeNotFound       = NewErrorCode(GetCode(http.StatusNotFound, ModuleCommon, 1))

	// Auth error codes
	ErrorCodeMissingAuthHeader       = NewErrorCode(GetCode(http.StatusUnauthorized, ModuleAuth, 1))
	ErrorCodeInvalidAuthFormat       = NewErrorCode(GetCode(http.StatusUnauthorized, ModuleAuth, 2))
	ErrorCodeInvalidToken            = NewErrorCode(GetCode(http.StatusUnauthorized, ModuleAuth, 3))
	ErrorCodeTokenExpired            = NewErrorCode(GetCode(http.StatusUnauthorized, ModuleAuth, 4))
	ErrorCodeInsufficientPermissions = NewErrorCode(GetCode(http.StatusForbidden, ModuleAuth, 5))
	ErrorCodeAuthUnauthorized        = NewErrorCode(GetCode(http.StatusUnauthorized, ModuleAuth, 6))
)
