package errors

var (
	// Authentication errors
	ErrMissingAuthHeader = New(
		ErrorCodeMissingAuthHeader,
		"missing authorization header",
	)

	ErrInvalidAuthFormat = New(
		ErrorCodeInvalidAuthFormat,
		"invalid authorization format",
	)

	ErrInvalidToken = New(
		ErrorCodeInvalidToken,
		"invalid token",
	)

	ErrTokenExpired = New(
		ErrorCodeTokenExpired,
		"token expired",
	)

	ErrInsufficientPermissions = New(
		ErrorCodeInsufficientPermissions,
		"insufficient permissions",
	)

	ErrUnauthorized = New(
		ErrorCodeAuthUnauthorized,
		"unauthorized",
	)

	// Common errors
	ErrInvalidRequest = New(
		ErrorCodeInvalidRequest,
		"invalid request",
	)

	ErrInternalServer = New(
		ErrorCodeInternalServer,
		"internal server error",
	)
	ErrNoResponse = New(
		ErrorCodeNoResponse,
		"no response",
	)
	ErrNotFound = New(
		ErrorCodeNotFound,
		"not found",
	)
)
