package middleware

import (
	"context"
	"demo/ms_user/internal/domain"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Permission represents a required permission
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
	"/user.v1.UserService/CreateUser":            PermissionWriteUser, // Allow authenticated users to create their own profile
	"/user.v1.UserService/CheckUserStatus":       PermissionReadUser,
}

// AuthorizationInterceptor provides authorization checks
type AuthorizationInterceptor struct {
	userRepo domain.UserRepository
}

// NewAuthorizationInterceptor creates a new authorization interceptor
func NewAuthorizationInterceptor(userRepo domain.UserRepository) *AuthorizationInterceptor {
	return &AuthorizationInterceptor{userRepo: userRepo}
}

// Unary returns a gRPC unary server interceptor for authorization
func (a *AuthorizationInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Get user ID from context (set by authentication middleware)
		userID, ok := ctx.Value("user_id").(string)
		if !ok || userID == "" {
			return nil, status.Errorf(codes.Unauthenticated, "user ID not found in context")
		}

		// Get required permission for this method
		requiredPermission, exists := MethodPermissions[info.FullMethod]
		if !exists {
			// If no permission is required, allow the request
			return handler(ctx, req)
		}

		// Get user from repository to check role
		user, err := a.userRepo.GetByClerkID(ctx, userID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
		}

		// Check if user has required permission
		if !a.hasPermission(Role(user.Role), requiredPermission) {
			return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions for %s", info.FullMethod)
		}

		// Add user role to context for further use
		ctx = context.WithValue(ctx, "user_role", user.Role)
		ctx = context.WithValue(ctx, "user", user)

		return handler(ctx, req)
	}
}

// hasPermission checks if a role has a specific permission
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

// IsAdmin checks if the current user is an admin
func IsAdmin(ctx context.Context) bool {
	role, ok := ctx.Value("user_role").(string)
	return ok && strings.ToLower(role) == "admin"
}

// GetCurrentUser retrieves the current user from context
func GetCurrentUser(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value("user").(*domain.User)
	return user, ok
}
