package main

import (
	"bck/auth/authpb"
	"bck/product/productpb"
	"log"
	"net"
	"net/http"
	"os"
	"product-service/internal/database"
	"product-service/internal/handlers"
	"product-service/internal/kafka"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}
	grpcport := os.Getenv("PRODUCT_GRPC_PORT")
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")
	authAddr := os.Getenv("AUTH_SERVICE_ADDR")
	httpPort := os.Getenv("HTTP_PORT")
	topic := os.Getenv("KAFKA_TOPIC")
	brokersENV := os.Getenv("KAFKA_BROKERS")
	brokers := strings.Split(brokersENV, ",")

	// mongodb connection
	log.Println("connecting mongodb")
	cl, err := mongo.Connect(options.Client().ApplyURI(mongoURI).SetConnectTimeout(5 * time.Second))
	if err != nil {
		log.Fatalf("couldnt connect to mongodb: %v", err)
	}
	repo := database.NewMongoRepo(cl, dbName)
	authConn, err := grpc.NewClient("localhost:"+authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("couldnt connect to authservice grpc: %v", err)
	}
	defer authConn.Close()

	authClient := authpb.NewAuthServiceClient(authConn)
	productProducer := kafka.NewProductProducer(brokers, topic)
	productHandler := handlers.NewProductHandler(repo, authClient, productProducer)

	// grpc
	server := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	productpb.RegisterProductServiceServer(server, productHandler)
	reflection.Register(server)

	lis, err := net.Listen("tcp", ":"+grpcport)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcport, err)
	}

	// http handlers
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	http.HandleFunc("/product", productHandler.CreateProductHTTP)
	http.HandleFunc("/products/get", productHandler.GetAllProductsHTTP)
	http.HandleFunc("/products/search", productHandler.SearchProductHTTP)
	http.HandleFunc("/products/update", productHandler.UpdateProductHTTP)
	http.HandleFunc("/products/delete", productHandler.DeleteProductHTTP)

	go func() {
		log.Printf("Product HTTP service listneing on port %s ", httpPort)
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			log.Fatalf("Couldnt serve the Prouct HTTP service: %v", err)
		}
	}()
	log.Printf("ProductService gRPC server listening on port %s", grpcport)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
