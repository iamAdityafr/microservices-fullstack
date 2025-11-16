package main

import (
	"bck/auth/authpb"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v78"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"payment-service/internal/database"
	"payment-service/internal/handlers"
	"payment-service/internal/kafka"
	"payment-service/internal/service"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	stripeWebhook := os.Getenv("STRIPE_WEBHOOK_SECRET")
	paymentHttpPort := os.Getenv("PAYMENT_HTTP_PORT")
	mongoURI := os.Getenv("MONGO_URI")
	dbname := os.Getenv("MONGO_DB")
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	authGrpcServicePort := os.Getenv("GRPC_Auth_Service_PORT")
	if stripeKey == "" || mongoURI == "" || kafkaTopic == "" || authGrpcServicePort == "" || paymentHttpPort == "" {
		log.Fatal("missing required env vars")
	}
	stripe.Key = stripeKey

	// connect db
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("mongo connect:", err)
	}
	defer func() { _ = mongoClient.Disconnect(ctx) }()

	repo := database.NewMongoPaymentRepo(mongoClient, dbname)
	// init services
	authConn, err := grpc.NewClient(authGrpcServicePort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("couldnt connect to authservice: ", err)

	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)
	paymentProducer := kafka.NewPaymentProducer(kafkaBrokers, kafkaTopic)
	// paymentConsumer := kafka.NewPaymentConsumer(kafkaBrokers, kafkaTopic, "payment-service-group", repo)

	stripeProvider := service.NewStripeProvider(stripeKey, stripeWebhook, kafka.NewPaymentProducer(kafkaBrokers, "payment"))

	handler := handlers.NewPaymentHandler(repo, paymentProducer, stripeProvider, authClient)

	http.HandleFunc("/payments/intent", handler.CreateIntent)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /payments/intent", handler.CreateIntent)
	mux.HandleFunc("GET /payments/", handler.GetPayment) // /payments/{orderID}
	mux.HandleFunc("POST /payments/webhook", handler.HandleWebhook)

	go func() {
		log.Println("Payment HTTP listening on :8085")
		if err := http.ListenAndServe(":"+paymentHttpPort, nil); err != nil && err != http.ErrServerClosed {
			log.Fatal("http serve:", err)
		}
	}()

	// go func() {
	// 	log.Println("Kafka consumer starting")
	// 	if err := consumer.Consume(context.Background()); err != nil {
	// 		log.Println("consumer error:", err)
	// 	}
	// }()
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("[Shutting down]: Payment service gracefully")
		log.Println("[Shutting down]: PaymentProducer")
		if err := paymentProducer.Close(ctx); err != nil {
			log.Printf("[Error]: closing PaymentProducer: %v", err)
		}
		log.Println("[Shutting down]: mongodb")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("[Error]: disconnecting from mongodb: %v", err)
		}
	}()

	log.Println("service stopped")
}
