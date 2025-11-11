package main

import (
	"bck/auth/authpb"
	"bck/user/userpb"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"user-service/internal/database"
	handler "user-service/internal/handlers"
	"user-service/internal/kafka"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	// get env vars
	httpPort := os.Getenv("HTTP_PORT")
	port := os.Getenv("USER_GRPC_PORT")
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")
	authAddr := os.Getenv("AUTH_SERVICE_ADDR")
	topic := os.Getenv("KAFKA_TOPIC")
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092"
	}
	brokers := strings.Split(brokersEnv, ",")

	// connect db
	log.Println("Connecting mongodb...")
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI).SetConnectTimeout(10 * time.Second))
	if err != nil {
		log.Fatalf("[Error]: Failed to connect mongodb: %v", err)
	}
	repo := database.NewMongoRepo(client, dbName)
	authConn, err := grpc.NewClient("localhost:"+authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[Error]: Failed to connect to AuthService: %v", err)
	}
	defer authConn.Close()

	// init services
	authClient := authpb.NewAuthServiceClient(authConn)
	userProducer := kafka.NewUserProducer(brokers, topic)
	userHandler := handler.NewUserHandler(repo, authClient, userProducer)
	// Set up HTTP handlers
	http.HandleFunc("/profile", userHandler.Profile)
	http.HandleFunc("/register", userHandler.Register)
	http.HandleFunc("/login", userHandler.Login)
	http.HandleFunc("/logout", userHandler.Logout)

	go func() {
		log.Printf("User HTTP service listening on port %s", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			log.Fatalf("[Error]: Failed to serve HTTP: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("[Error]: Failed to listen on port %s: %v", port, err)
	}

	// start grpc server
	server := grpc.NewServer()
	userpb.RegisterUserServiceServer(server, userHandler)
	reflection.Register(server)

	// graceful shutdown btw
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("[Shutting down]: user service gracefully")
		log.Println("[Shutting down]: UserProducer")
		if err := userProducer.Close(); err != nil {
			log.Printf("[Error]: closing UserProducer: %v", err)
		}
		log.Println("[Shutting down]: GRPC server")
		server.GracefulStop()
		log.Println("[Shutting down]: mongodb")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("[Error]: disconnecting from mongodb: %v", err)
		}
	}()

	log.Printf("User grpc service listening on port %s", port)
	log.Printf("mongodb: %s, Database: %s", mongoURI, dbName)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("[Error]: Failed to serve: %v", err)
	}
}
