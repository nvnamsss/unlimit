package utility

import (
	"net/http"

	"github.com/voidforge-studios/unlimit/errors"
	"github.com/voidforge-studios/unlimit/logger"
)

// HandleError handles API errors and returns status code and response body
func HandleError(err error) (int, map[string]any) {
	appErr := toError(err)
	statusCode := mapErrorToStatus(appErr)

	// Optionally log the error (can be removed or customized as needed)
	if statusCode >= 500 {
		logger.Errorf("Error handling request: %+v", appErr)
	} else {
		logger.Warnf("Client error: %+v", appErr)
	}

	body := map[string]any{
		"error": map[string]any{
			"code":    appErr.Code.Code(),
			"message": appErr.Message,
		},
	}
	return statusCode, body
}

// toError converts any error to *errors.Error
func toError(err error) *errors.Error {
	if e, ok := err.(*errors.Error); ok {
		return e
	}

	v := errors.New(errors.NewErrorCode(500), err.Error())
	return v.(*errors.Error)
}

// mapErrorToStatus maps Error code to HTTP status code
func mapErrorToStatus(err *errors.Error) int {
	switch err.Code.Status() {
	case http.StatusNotFound:
		return http.StatusNotFound
	case http.StatusConflict:
		return http.StatusConflict
	case http.StatusBadRequest:
		return http.StatusBadRequest
	case http.StatusForbidden:
		return http.StatusForbidden
	case http.StatusUnauthorized:
		return http.StatusUnauthorized
	case http.StatusPaymentRequired:
		return http.StatusPaymentRequired
	case http.StatusServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
