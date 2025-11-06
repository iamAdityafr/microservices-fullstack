package main

import (
	"auth-service/internal/handler"
	"auth-service/internal/service"
	"auth-service/utils"
	"bck/auth/authpb"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading the envs: %v", err)
	}
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		log.Fatal("Field not found from .env")
	}
	utils.InitJWT()
	// grpc setup
	authService, err := service.NewAuthService()
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}
	authHandler := handler.NewAuthHandler(authService)
	server := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
	)

	authpb.RegisterAuthServiceServer(server, authHandler)
	reflection.Register(server)
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	log.Printf("AuthService server listening on port %s", port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Error in serving grpc server: %v", err)
	}
}
