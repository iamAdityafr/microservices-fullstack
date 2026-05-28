package main

import (
	"auth-service/internal/handler"
	"auth-service/internal/logger"
	"auth-service/internal/service"
	"auth-service/utils"
	"fmt"
	"grpc_module/auth/authpb"
	"net"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("error loading the envs: %v", err)
		return
	}
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		fmt.Print("Field not found from .env")
		return
	}
	logDev := os.Getenv("LOG_DEV")
	utils.InitJWT()

	// init logger
	logMode, err := strconv.ParseBool(logDev)
	if err != nil {
		fmt.Println("err parsing bool for logDev")
		return
	}
	logger, err := logger.InitLogger(logMode)
	if err != nil {
		fmt.Println("err in init logger")
		return
	}

	// grpc setup
	authService, err := service.NewAuthService()
	if err != nil {
		logger.Error("failed to create auth service", zap.Error(err))
		return
	}
	authHandler := handler.NewAuthHandler(logger, authService)
	server := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
	)

	authpb.RegisterAuthServiceServer(server, authHandler)
	reflection.Register(server)
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("failed to listen on port", zap.String("port", port), zap.Error(err))
		return
	}
	logger.Info("AuthService RPC server is listening", zap.String("port", port))
	if err := server.Serve(lis); err != nil {
		logger.Error("err in servicing grpc server", zap.Error(err))
		return
	}
}
