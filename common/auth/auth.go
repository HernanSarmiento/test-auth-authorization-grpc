package auth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MyCustomsClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	*jwt.RegisteredClaims
}

func VerifyToken(ctx context.Context, pubKey *ecdsa.PublicKey) (*MyCustomsClaims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Error: no metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "Error: missing authorization header")
	}

	tokenString := strings.TrimPrefix(values[0], "bearer ")
	claims := &MyCustomsClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Error: unexpected method %v", t.Header["alg"])
		}
		return pubKey, nil
	})

	if err != nil || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "Error: invalid token")
	}

	return claims, nil
}
