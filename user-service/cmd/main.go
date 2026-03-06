package main

import (
	"fmt"
	"log"
	"net"

	proto "github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/user"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/user-service/internal/config"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/user-service/internal/db"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/user-service/internal/handlers"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/user-service/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Coudn't load env var %s", err)
	}
	db := db.InitDB(cfg)
	addr := ":50051"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("couldn't listen on port %s error: %v", addr, err)
	}

	userRepo := repository.NewPostgresRepo(db)

	grpcServer := grpc.NewServer()
	UserService := handlers.NewUserHandler(userRepo)
	proto.RegisterUserServiceServer(grpcServer, UserService)
	reflection.Register(grpcServer)

	fmt.Printf("My app is running on port %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve grpc server at %s err: %s", addr, err)
	}
}
