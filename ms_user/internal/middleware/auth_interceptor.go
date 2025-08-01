package middleware

import (
	"context"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/client"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// PublicMetadata holds public information from the Clerk JWT.
type PublicMetadata struct {
	Role string `json:"role,omitempty"`
}

// AuthInterceptor handles JWT validation and extracts user information.
type AuthInterceptor struct {
	clerkClient *client.Client
	jwksClient  *jwks.Client
}

func NewAuthInterceptor(clerkClient *client.Client) *AuthInterceptor {
	return &AuthInterceptor{clerkClient: clerkClient}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		claims, err := jwt.Verify(ctx, &jwt.VerifyParams{Token: token})
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "token verification failed: %v", err)
		}

		// Inject user ID into context
		ctx = context.WithValue(ctx, "user_id", claims.Subject)

		return handler(ctx, req)
	}
}

// injectRoleFromCustomClaims parses custom claims and injects the user role into the context.
// It is designed to fail silently to avoid blocking authentication if metadata is malformed.
func (i *AuthInterceptor) injectRoleFromCustomClaims(ctx context.Context, customClaims interface{}) context.Context {
	customBytes, err := json.Marshal(customClaims)
	if err != nil {
		return ctx // Cannot process, return original context.
	}

	var claimsMap map[string]interface{}
	if err := json.Unmarshal(customBytes, &claimsMap); err != nil {
		return ctx
	}

	metadataData, ok := claimsMap["metadata"]
	if !ok {
		return ctx
	}

	metadataBytes, err := json.Marshal(metadataData)
	if err != nil {
		return ctx
	}

	var publicMetadata PublicMetadata
	if err := json.Unmarshal(metadataBytes, &publicMetadata); err != nil {
		return ctx
	}

	if publicMetadata.Role != "" {
		ctx = context.WithValue(ctx, UserRoleKey, publicMetadata.Role)
	}

	return ctx
}
