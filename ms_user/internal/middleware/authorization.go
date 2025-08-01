package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Permission represents a required permission for an action.
type Permission string

const (
	PermissionReadUser   Permission = "user:read"
	PermissionWriteUser  Permission = "user:write"
	PermissionDeleteUser Permission = "user:delete"
	PermissionAdminUser  Permission = "user:admin"
)

// Role represents user roles
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// RolePermissions maps roles to their allowed permissions
var RolePermissions = map[Role][]Permission{
	RoleUser: {
		PermissionReadUser,
		PermissionWriteUser, // Users can modify their own data
	},
	RoleAdmin: {
		PermissionReadUser,
		PermissionWriteUser,
		PermissionDeleteUser,
		PermissionAdminUser,
	},
}

// MethodPermissions maps gRPC methods to required permissions
var MethodPermissions = map[string]Permission{
	"/user.v1.UserService/GetUser":               PermissionReadUser,
	"/user.v1.UserService/UpdateUser":            PermissionWriteUser,
	"/user.v1.UserService/UpdateUserPreferences": PermissionWriteUser,
	"/user.v1.UserService/DeleteUser":            PermissionDeleteUser,
	"/user.v1.UserService/CheckUserStatus":       PermissionReadUser,
	"/user.v1.UserService/ActivateUser":          PermissionWriteUser,
}

// ExemptMethods lists methods that bypass role validation but still require authentication.
var ExemptMethods = map[string]bool{
	"/user.v1.UserService/CreateUser": true, // Users need to create profile before getting roles.
}

// AuthorizationInterceptor provides RBAC (Role-Based Access Control) checks.
type AuthorizationInterceptor struct{}

// NewAuthorizationInterceptor creates a new AuthorizationInterceptor.
func NewAuthorizationInterceptor() *AuthorizationInterceptor {
	return &AuthorizationInterceptor{}
}

// Unary returns a gRPC unary server interceptor for authorization.
func (a *AuthorizationInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if method is exempt from role validation
		if ExemptMethods[info.FullMethod] {
			// Still require authentication (user ID must be present)
			if _, ok := ctx.Value(UserIDKey).(string); !ok {
				return nil, status.Errorf(codes.Unauthenticated, "user ID not found in context")
			}
			return handler(ctx, req)
		}

		requiredPermission, requiresAuth := MethodPermissions[info.FullMethod]
		if !requiresAuth {
			return handler(ctx, req)
		}

		// User ID must be present for any authenticated endpoint.
		if _, ok := ctx.Value(UserIDKey).(string); !ok {
			return nil, status.Errorf(codes.Unauthenticated, "user ID not found in context")
		}

		// The user role must be present in the JWT claims.
		userRole, ok := ctx.Value(UserRoleKey).(string)
		if !ok || userRole == "" {
			return nil, status.Errorf(codes.PermissionDenied, "user role not found in JWT claims")
		}

		if !a.hasPermission(Role(userRole), requiredPermission) {
			return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions for %s", info.FullMethod)
		}

		return handler(ctx, req)
	}
}

// hasPermission checks if a role has a specific permission.
func (a *AuthorizationInterceptor) hasPermission(role Role, permission Permission) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}
