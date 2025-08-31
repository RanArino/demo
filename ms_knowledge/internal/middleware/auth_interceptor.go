package middleware

import (
	"context"
	"encoding/json"
	"strings"

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

// Unary returns a gRPC unary interceptor that validates the Authorization Bearer token.
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
		claims, err := jwt.Verify(ctx, &jwt.VerifyParams{Token: token, JWKSClient: i.jwksClient})
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "token verification failed: %v", err)
		}
		// Extract public metadata.role if present and inject into context
		if claims.Custom != nil {
			raw, _ := json.Marshal(claims.Custom)
			var m map[string]interface{}
			_ = json.Unmarshal(raw, &m)
			if md, ok := m["metadata"]; ok {
				rawMD, _ := json.Marshal(md)
				var mm map[string]interface{}
				_ = json.Unmarshal(rawMD, &mm)
				if role, ok2 := mm["role"].(string); ok2 && role != "" {
					ctx = context.WithValue(ctx, domain.RoleKey, role)
				}
			}
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
