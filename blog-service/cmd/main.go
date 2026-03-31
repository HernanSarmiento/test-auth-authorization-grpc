package main

import (
	"log"
	"net"

	"github.com/HernanSarmiento/test-auth-authorization-grpc/api/proto/gen/blog"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/blog-service/internal/config"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/blog-service/internal/db"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/blog-service/internal/handlers"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/blog-service/internal/interceptors"
	"github.com/HernanSarmiento/test-auth-authorization-grpc/blog-service/internal/repository"
	common "github.com/HernanSarmiento/test-auth-authorization-grpc/common/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error couldn't load env vars %s", err)
	}

	dbConn := db.InitDB(cfg)
	sqlDB, err := dbConn.DB()
	if err != nil {
		log.Fatalf("Error al obtener la instancia de sql.DB: %v", err)
	}
	defer sqlDB.Close()

	addr := ":" + cfg.BLOG_SERVER_PORT
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("[Error]: coudln't listen on port %s", addr)
	}

	pubKey, err := common.LoadPublicKey(cfg.PublicKeyPath)
	if err != nil {
		log.Fatalf("Error while loading public key %v", err)
	}

	blogRepo := repository.NewPostgresRepo(dbConn)
	blogHandler := handlers.NewBlogHandler(blogRepo)

	blogInterceptor := interceptors.AuthInterceptor(pubKey)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(blogInterceptor))
	blog.RegisterBlogServiceServer(grpcServer, blogHandler)
	reflection.Register(grpcServer)

	log.Printf("[Blog-service] Running on %s port 📰", addr)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[Error]: Couldn't serve [Blog-service] err: %s", err)
	}
}
