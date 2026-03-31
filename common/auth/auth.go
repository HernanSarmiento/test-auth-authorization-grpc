package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	UserIDKey     contextKey = "user_id"
	AuthorNameKey contextKey = "author_name"
	RoleKey       contextKey = "user_role"
)

type MyCustomsClaims struct {
	UserID     string `json:"user_id"`
	AuthorName string `json:"author_name"`
	Role       string `json:"role"`
	*jwt.RegisteredClaims
}

func LoadPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read private key: %w", err)
	}
	key, err := jwt.ParseECPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("Error occur while parsing private key %w", err)
	}
	return key, nil
}

func LoadPublicKey(path string) (*ecdsa.PublicKey, error) {
	// 1. Leer el archivo físico (.pem)
	asn1Data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error leyendo llave pública: %w", err)
	}

	// 2. Decodificar el bloque PEM
	block, _ := pem.Decode(asn1Data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("formato de llave pública inválido")
	}
	// 3. Parsear el contenido PKIX (estándar para llaves públicas)
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parseando llave pública: %w", err)
	}
	// 4. Asegurarnos de que sea una llave ECDSA (la que elegimos para el proyecto)
	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("la llave no es de tipo ECDSA")
	}

	return ecdsaPub, nil
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

	tokenString := strings.TrimPrefix(values[0], "Bearer ")
	claims := &MyCustomsClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Error: unexpected method %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		fmt.Printf("JWT Error: %v\n", err)
		return nil, err
	}
	if err != nil || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "Error: invalid token")
	}

	return claims, nil
}
