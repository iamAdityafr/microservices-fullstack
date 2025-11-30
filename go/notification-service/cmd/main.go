package main

import (
	"context"
	"log"
	"notification-service/kafka"
	"notification-service/service"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf("err loading .env")
	}
	mailGunKey := os.Getenv("MAILGUN_API_KEY")
	mainGunDomain := os.Getenv("MAILGUN_DOMAIN")
	brokerENV := os.Getenv("KAFKA_BROKERS")
	brokers := strings.Split(brokerENV, ",")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	kafkaTopis := strings.Split(kafkaTopic, ",")

	mailer := service.NewMailGunMailer(mailGunKey, mainGunDomain)
	notificationConsumer := kafka.NewNotificationConsumer(brokers, kafkaTopis, "mail-service-group", mailer)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("[Shutting down]: notification service")
		cancel()
	}()

	if err := notificationConsumer.Consume(ctx); err != nil {
		log.Printf("Err consuming messages: %v", err)
	}

	if err := notificationConsumer.Close(); err != nil {
		log.Printf("Err closing consumer: %v", err)
	}

	log.Println("Notification service stopped")
}
