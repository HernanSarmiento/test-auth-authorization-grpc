package interceptors

import (
	"context"
	"crypto/ecdsa"

	"github.com/HernanSarmiento/test-auth-authorization-grpc/common/auth"
	"google.golang.org/grpc"
)

func AuthInterceptor(pubKey *ecdsa.PublicKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		claims, err := auth.VerifyToken(ctx, pubKey)
		if err != nil {
			return nil, err
		}

		newctx := context.WithValue(ctx, auth.UserIDKey, claims.ID)
		newCtx := context.WithValue(newctx, auth.RoleKey, auth.AuthorNameKey)

		return handler(newCtx, req)
	}
}
