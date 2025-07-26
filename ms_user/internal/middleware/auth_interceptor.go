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

type AuthInterceptor struct {
	clerkClient *client.Client
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
