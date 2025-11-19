package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	JWTSecret              string
	UserServiceURL         string
	ProductServiceURL      string
	OrderServiceURL        string
	CartServiceURL         string
	PaymentServiceURL      string
	NotificationServiceURL string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No env file found")
	}

	return &Config{
		JWTSecret:              os.Getenv("JWT_SECRET"),
		Port:                   os.Getenv("API_GATEWAY_PORT"),
		UserServiceURL:         os.Getenv("USER_SERVICE_URL"),
		ProductServiceURL:      os.Getenv("PRODUCT_SERVICE_URL"),
		OrderServiceURL:        os.Getenv("ORDER_SERVICE_URL"),
		CartServiceURL:         os.Getenv("CART_SERVICE_URL"),
		PaymentServiceURL:      os.Getenv("PAYMENT_SERVICE_URL"),
		NotificationServiceURL: os.Getenv("NOTIFICATION_SERVICE_URL"),
	}
}
