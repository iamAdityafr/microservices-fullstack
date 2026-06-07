package main

import (
	"cart-service/internal/database"
	"cart-service/internal/handlers"
	"cart-service/internal/logger"
	"fmt"
	"grpc_module/auth/authpb"
	"grpc_module/cart/cartpb"
	"grpc_module/product/productpb"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("err loading the envs ", err)
	}
	grpcport := os.Getenv("CART_GRPC_PORT")
	authgrpcport := os.Getenv("AUTH_GRPC_PORT")
	productgrpcport := os.Getenv("PRODUCT_GRPC_PORT")
	dbname := os.Getenv("DB_NAME")
	logDev := os.Getenv("LOG_DEV")
	mongoUri := os.Getenv("MONGO_URI")
	httpPort := os.Getenv("HTTP_PORT")
	// kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}
	// kafkaTopic := os.Getenv("KAFKA_TOPIC")

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

	// db connection
	log.Println("connect to mongodb")
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoUri).SetTimeout(5 * time.Second))
	if err != nil {
		log.Fatalf("couldnt connect to the mongodb: %v", err)
	}
	repo := database.NewMongoRepo(mongoClient, dbname)

	// grpc
	authConn, err := grpc.NewClient("localhost:"+authgrpcport, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("couldnt connect to authservice grpc: %v", err)
	}
	defer authConn.Close()

	productConn, err := grpc.NewClient("localhost:"+productgrpcport, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("couldnt connect to authservice grpc: %v", err)
	}
	defer authConn.Close()
	productClient := productpb.NewProductServiceClient(productConn)
	authClient := authpb.NewAuthServiceClient(authConn)
	carthandler := handlers.NewCartHandler(repo, logger, productClient, authClient)
	// cartProducer := kafka.NewCartProducer(kafkaBrokers, kafkaTopic)
	// cartConsumer := kafka.NewCartConsumer(kafkaBrokers, kafkaTopic, "cart-service-group", repo)

	server := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	cartpb.RegisterCartServiceServer(server, carthandler)
	reflection.Register(server)

	// http handlers
	http.Handle("/", carthandler.Routes())

	go func() {
		log.Printf("Product HTTP service listneing on port %s ", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			log.Fatalf("Couldnt serve the Prouct HTTP service: %v", err)
		}
	}()
	lis, err := net.Listen("tcp", ":"+grpcport)
	if err != nil {
		log.Fatalf("Failed to listen to port %s: %v", grpcport, err)
	}
	log.Println("prouctservice grpc server listening on port: ", grpcport)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Couldnt serve grpc cart service: %v", err)
	}
}
