package middleware

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	domain "demo/ms_knowledge/internal/domain"
	userv1 "demo/ms_user/api/proto/v1"
)

// AuthInterceptor validates Clerk JWTs on inbound gRPC requests and resolves user identities.
type AuthInterceptor struct {
	jwksClient *jwks.Client
	userClient userv1.UserServiceClient
}

// NewAuthInterceptor constructs a new AuthInterceptor using Clerk JWKS and User service client.
// Both clerkSecretKey and userClient are required for proper authentication flow.
func NewAuthInterceptor(clerkSecretKey string, userClient userv1.UserServiceClient) *AuthInterceptor {
	if clerkSecretKey == "" {
		panic("clerkSecretKey is required for AuthInterceptor")
	}
	if userClient == nil {
		panic("userClient is required for AuthInterceptor")
	}

	jwksClient := jwks.NewClient(&clerk.ClientConfig{
		BackendConfig: clerk.BackendConfig{Key: &clerkSecretKey},
	})
	return &AuthInterceptor{
		jwksClient: jwksClient,
		userClient: userClient,
	}
}

// Unary returns a gRPC unary interceptor that validates the Authorization Bearer token
// and resolves the Clerk user ID to internal user ID via User service.
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		claims, err := jwt.Verify(ctx, &jwt.VerifyParams{
			Token:      token,
			JWKSClient: i.jwksClient,
			Leeway:     30 * time.Second,
		})
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "token verification failed: %v", err)
		}

		// Extract Clerk user ID from claims.Subject
		clerkUserID := claims.Subject
		if clerkUserID == "" {
			return nil, status.Errorf(codes.Unauthenticated, "user ID not found in token")
		}

		// Inject Clerk user ID into context for debugging/logging
		ctx = context.WithValue(ctx, domain.ClerkUserIDKey, clerkUserID)

		// Extract role from custom claims and inject into context
		var role string
		if claims.Custom != nil {
			raw, _ := json.Marshal(claims.Custom)
			var m map[string]interface{}
			_ = json.Unmarshal(raw, &m)
			if md, ok := m["metadata"]; ok {
				rawMD, _ := json.Marshal(md)
				var mm map[string]interface{}
				_ = json.Unmarshal(rawMD, &mm)
				if r, ok2 := mm["role"].(string); ok2 && r != "" {
					role = r
				}
			}
		}
		ctx = context.WithValue(ctx, domain.RoleKey, role)

		// Resolve Clerk user ID to internal user ID via User service
		internalUserID, userRole, err := i.resolveUserInfo(ctx, md)
		if err != nil {
			return nil, err // Error already formatted by resolveUserInfo
		}

		// Inject internal user ID into context
		ctx = context.WithValue(ctx, domain.OwnerIDKey, internalUserID)

		// Override role with the one from User service (server-side source of truth)
		if userRole != "" {
			ctx = context.WithValue(ctx, domain.RoleKey, userRole)
		}

		return handler(ctx, req)
	}
}

// resolveUserInfo calls the User service to resolve Clerk user ID to internal user ID and role.
func (i *AuthInterceptor) resolveUserInfo(ctx context.Context, md metadata.MD) (string, string, error) {
	// User service client should always be available due to constructor validation
	if i.userClient == nil {
		return "", "", status.Errorf(codes.Internal, "user service client not configured")
	}

	// Create outgoing context with the same metadata (including authorization header)
	outCtx := metadata.NewOutgoingContext(ctx, md)

	// Get Clerk user ID from context
	clerkUserID, ok := ctx.Value(domain.ClerkUserIDKey).(string)
	if !ok || clerkUserID == "" {
		return "", "", status.Errorf(codes.Internal, "clerk user ID not found in context")
	}

	// Call User service to get user details using Clerk user ID
	uReq := &userv1.GetUserRequest{
		Identifier: &userv1.GetUserRequest_ClerkUserId{
			ClerkUserId: clerkUserID,
		},
	}
	uResp, err := i.userClient.GetUser(outCtx, uReq)
	if err != nil {
		// Check if it's a gRPC error and handle appropriately
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.Unauthenticated:
				return "", "", status.Errorf(codes.Unauthenticated, "user authentication failed: %v", err)
			case codes.NotFound:
				return "", "", status.Errorf(codes.Unauthenticated, "user not found in system")
			case codes.Unavailable:
				return "", "", status.Errorf(codes.Internal, "user service unavailable: %v", err)
			default:
				return "", "", status.Errorf(codes.Internal, "user service error: %v", err)
			}
		}
		return "", "", status.Errorf(codes.Internal, "user service call failed: %v", err)
	}

	// Validate response
	if uResp == nil || uResp.GetUser() == nil {
		return "", "", status.Errorf(codes.Internal, "invalid response from user service")
	}

	user := uResp.GetUser()
	if user.GetId() == "" {
		return "", "", status.Errorf(codes.Internal, "user service returned empty user ID")
	}

	return user.GetId(), user.GetRole(), nil
}
