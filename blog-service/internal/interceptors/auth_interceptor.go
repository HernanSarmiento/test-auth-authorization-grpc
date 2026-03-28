package interceptors

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/HernanSarmiento/test-auth-authorization-grpc/common/auth"
	"google.golang.org/grpc"
)

var publicMethods = map[string]bool{
	"/blog-service/GetAllPosts": true,
	"/blog-service/GetPost":     true,
}

func AuthInterceptor(pubKey *ecdsa.PublicKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		fmt.Printf("[Interceptor]Trying to access to: %s\n", info.FullMethod)

		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		claims, err := auth.VerifyToken(ctx, pubKey)
		if err != nil {
			return nil, err
		}

		fmt.Printf("[Interceptor]: User %s (%s) authorized for %s\n",
			claims.AuthorName, claims.ID, info.FullMethod,
		)

		ctx = context.WithValue(ctx, auth.UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, auth.AuthorNameKey, claims.AuthorName)
		ctx = context.WithValue(ctx, auth.RoleKey, claims.Role)

		resp, err := handler(ctx, req)
		if err != nil {
			fmt.Printf("[Interceptor]: Method %s executed by %s FAILED: %s\n", info.FullMethod, claims.AuthorName, err)
		} else {
			fmt.Printf("[Interceptor]: Method %s executed Successfully by %s SUCCESS", info.FullMethod, claims.AuthorName)
		}

		return resp, err
	}
}
