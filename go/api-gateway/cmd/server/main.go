package main

import (
	"api-gateway/internal/config"
	"api-gateway/internal/middleware"
	"api-gateway/internal/proxy"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Starting [API Gateway] on port %s", cfg.Port)
	log.Printf("User Service: %s", cfg.UserServiceURL)
	log.Printf("Product Service: %s", cfg.PaymentServiceURL)
	log.Printf("Cart Service: %s", cfg.CartServiceURL)
	log.Printf("Payment Service: %s", cfg.PaymentServiceURL)
	log.Printf("Notification Service: %s", cfg.NotificationServiceURL)

	router := proxy.NewRouter(cfg)

	handler := middleware.CORSMiddleware(router)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("[Shutting down]: API Gateway gracefully")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Err in shutdown: %v", err)
		}
	}()

	log.Print("API Gateway listening on", cfg.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
	log.Print("gateway stopped !")
}
