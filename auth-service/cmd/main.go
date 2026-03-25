package main

import (
	"log"
	"net"

	authpb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/auth"
	userpb "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/user"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/auth-service/internal/config"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/auth-service/internal/handlers"
	repo "github.com/HernanSarmiento/test-auth-authorization-grpc/auth-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error couldn't load env vars %v", err)
	}

	privKey, err := handlers.LoadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		log.Fatalf("Error while loading private key %v", err)
	}

	pubKey, err := handlers.LoadPublicKey(cfg.PublicKeyPath)
	if err != nil {
		log.Fatalf("Error while loading public key %v", err)
	}
	addr := ":" + cfg.AUTH_SERVER_PORT

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Error while running up auth-service port %v", err)
	}

	userAddr := "localhost:50051"

	userConn, _ := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer userConn.Close()
	userClient := userpb.NewUserServiceClient(userConn)

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	redisRepo := repo.RedisBlackListRepo{Client: rdb}

	authService := handlers.NewAuthService(userClient, privKey, redisRepo, pubKey)

	grpcServer := grpc.NewServer()
	authpb.RegisterAuthServiceServer(grpcServer, authService)
	reflection.Register(grpcServer)

	log.Printf("Auth Service is running on %s 🚀", addr)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error while serving auth service %v", err)
	}
}
