// Package utility provides helper functions for context management in Gin and standard Go contexts.
// This package includes utilities for storing and retrieving user information, roles, and UUIDs
// from both Gin contexts and standard Go contexts.
package utility

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ContextKey string

const (
	// UserIDKey is the context key for storing user ID values.
	// Used to identify the current authenticated user in the system.
	UserIDKey ContextKey = "user_id"
	// RefUSerIDKey is the context key for storing reference user ID values.
	// Used to track related or referenced users in the context.
	RefUSerIDKey ContextKey = "ref_user_id"
	// RoleKey is the context key for storing user role information.
	// Used to store the role/permission level of the current user.
	RoleKey ContextKey = "role"
	// UUIDKey is the context key for storing UUID values.
	// Used to store unique identifiers in the context.
	UUIDKey        ContextKey = "uuid"
	FingerprintKey ContextKey = "fingerprint"
)

func GetValueFromContext[T any](ctx context.Context, key ContextKey) (T, bool) {
	value, ok := ctx.Value(key).(T)
	return value, ok
}

func SetValueToContext(ctx context.Context, key ContextKey, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

func GetValueFromGinContext[T any](ctx *gin.Context, key ContextKey) (T, bool) {
	value, ok := ctx.Get(string(key))
	if ok {
		return value.(T), ok
	}

	return GetValueFromContext[T](ctx.Request.Context(), key)
}

func SetValueToGinContext(ctx *gin.Context, key ContextKey, value interface{}) {
	ctx.Set(string(key), value)
	SetValueToContext(ctx.Request.Context(), key, value)
}

// GetUserIDFromGinContext retrieves the user ID from a Gin context.
// It first checks the Gin context storage, then falls back to the request context.
// Returns the user ID and a boolean indicating if the value was found.
// Example usage:
//
//	func handler(c *gin.Context) {
//		userID, exists := GetUserIDFromGinContext(c)
//		if exists {
//			fmt.Printf("User ID: %d", userID)
//		}
//	}
//
// This function is useful for retrieving user identification in Gin handlers.
func GetUserIDFromGinContext(ctx *gin.Context) (int64, bool) {
	userID, ok := ctx.Get(string(UserIDKey))
	if ok {
		return userID.(int64), ok
	}

	return GetUserIDFromContext(ctx.Request.Context())
}

// GetUserIDFromContext retrieves the user ID from a standard Go context.
// Returns the user ID and a boolean indicating if the value was found.
// Example usage:
//
//	ctx := context.Background()
//	ctx = SetUserIDToContext(ctx, 123)
//	userID, exists := GetUserIDFromContext(ctx)
//	fmt.Printf("User ID: %d, exists: %t", userID, exists) // Output: User ID: 123, exists: true
//
// This function is useful for retrieving user identification from standard contexts.
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

// SetUserIDToContext stores a user ID in a standard Go context.
// Returns a new context with the user ID value stored.
// Example usage:
//
//	ctx := context.Background()
//	newCtx := SetUserIDToContext(ctx, 123)
//	userID, _ := GetUserIDFromContext(newCtx)
//	fmt.Printf("Stored user ID: %d", userID) // Output: Stored user ID: 123
//
// This function is useful for passing user identification through context chains.
func SetUserIDToContext(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// SetUserIDToGinContext stores a user ID in a Gin context.
// The value is stored in the Gin context's key-value store.
// Example usage:
//
//	func middleware(c *gin.Context) {
//		SetUserIDToGinContext(c, 123)
//		c.Next()
//	}
//
// This function is useful for storing user identification in Gin middleware.
func SetUserIDToGinContext(ctx *gin.Context, userID int64) {
	ctx.Set(string(UserIDKey), userID)
	SetUserIDToContext(ctx.Request.Context(), userID)
}

// SetUserRefIDToGinContext stores a reference user ID in a Gin context.
// The value is stored in the Gin context's key-value store using the RefUSerIDKey.
// Example usage:
//
//	func middleware(c *gin.Context) {
//		SetUserRefIDToGinContext(c, "ref-user-abc123")
//		c.Next()
//	}
//
// This function is useful for storing reference user identification in Gin middleware
// when tracking related or referenced users in the context.
func SetUserRefIDToGinContext(ctx *gin.Context, userID string) {
	ctx.Set(string(RefUSerIDKey), userID)
	SetUserRefIDToContext(ctx.Request.Context(), userID)
}

// SetUserRefIDToContext stores a reference user ID in a standard Go context.
// Returns a new context with the reference user ID value stored.
// Example usage:
//
//	ctx := context.Background()
//	newCtx := SetUserRefIDToContext(ctx, "ref-user-abc123")
//	refUserID, _ := GetUserRefIDFromContext(newCtx)
//	fmt.Printf("Stored ref user ID: %s", refUserID) // Output: Stored ref user ID: ref-user-abc123
//
// This function is useful for passing reference user identification through context chains.
func SetUserRefIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, RefUSerIDKey, userID)
}

// GetUserRefIDFromGinContext retrieves the reference user ID from a Gin context.
// It first checks the Gin context storage, then falls back to the request context.
// Returns the reference user ID string and a boolean indicating if the value was found.
// Example usage:
//
//	func handler(c *gin.Context) {
//		refUserID, exists := GetUserRefIDFromGinContext(c)
//		if exists {
//			fmt.Printf("Reference User ID: %s", refUserID)
//		}
//	}
//
// This function is useful for retrieving reference user identification in Gin handlers
// when tracking related or referenced users.
func GetUserRefIDFromGinContext(ctx *gin.Context) (string, bool) {
	userID, ok := ctx.Get(string(RefUSerIDKey))
	if ok {
		return userID.(string), ok
	}

	return GetUserRefIDFromContext(ctx.Request.Context())
}

// GetUserRefIDFromContext retrieves the reference user ID from a standard Go context.
// Returns the reference user ID string and a boolean indicating if the value was found.
// Example usage:
//
//	ctx := context.Background()
//	ctx = SetUserRefIDToContext(ctx, "ref-user-abc123")
//	refUserID, exists := GetUserRefIDFromContext(ctx)
//	fmt.Printf("Ref User ID: %s, exists: %t", refUserID, exists) // Output: Ref User ID: ref-user-abc123, exists: true
//
// This function is useful for retrieving reference user identification from standard contexts.
func GetUserRefIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(RefUSerIDKey).(string)
	return userID, ok
}

// GetRoleFromGinContext retrieves the user role from a Gin context.
// It first checks the Gin context storage, then falls back to the request context.
// Returns the role string and a boolean indicating if the value was found.
// Example usage:
//
//	func handler(c *gin.Context) {
//		role, exists := GetRoleFromGinContext(c)
//		if exists {
//			fmt.Printf("User role: %s", role)
//		}
//	}
//
// This function is useful for retrieving user roles in Gin handlers for authorization.
func GetRoleFromGinContext(ctx *gin.Context) (string, bool) {
	role, ok := ctx.Get(string(RoleKey))
	if ok {
		return role.(string), ok
	}

	return GetRoleFromContext(ctx.Request.Context())
}

// GetRoleFromContext retrieves the user role from a standard Go context.
// Returns the role string and a boolean indicating if the value was found.
// Example usage:
//
//	ctx := context.Background()
//	ctx = SetRoleToContext(ctx, "admin")
//	role, exists := GetRoleFromContext(ctx)
//	fmt.Printf("Role: %s, exists: %t", role, exists) // Output: Role: admin, exists: true
//
// This function is useful for retrieving user roles from standard contexts.
func GetRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok
}

// SetRoleToContext stores a user role in a standard Go context.
// Returns a new context with the role value stored.
// Example usage:
//
//	ctx := context.Background()
//	newCtx := SetRoleToContext(ctx, "admin")
//	role, _ := GetRoleFromContext(newCtx)
//	fmt.Printf("Stored role: %s", role) // Output: Stored role: admin
//
// This function is useful for passing user roles through context chains.
func SetRoleToContext(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, RoleKey, role)
}

// SetRoleToGinContext stores a user role in a Gin context.
// The value is stored in the Gin context's key-value store.
// Example usage:
//
//	func authMiddleware(c *gin.Context) {
//		SetRoleToGinContext(c, "admin")
//		c.Next()
//	}
//
// This function is useful for storing user roles in Gin middleware for authorization.
func SetRoleToGinContext(ctx *gin.Context, role string) {
	ctx.Set(string(RoleKey), role)
}
