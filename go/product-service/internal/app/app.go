package app

import (
	"context"
	"fmt"
	"grpc_module/user/userpb"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"product-service/internal/config"
	"product-service/internal/database"
	handler "product-service/internal/handlers"
	"product-service/internal/kafka"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Application struct {
	cfg      config.Config
	logger   *zap.Logger
	grpcConn *grpc.ClientConn
	server   *http.Server
	dbclient *mongo.Client
	shutdown []func(context.Context) error
}

func New(cfg *config.Config, logger *zap.Logger) (*Application, error) {
	app := &Application{cfg: *cfg, logger: logger}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// mongo client
	clientOpts := options.Client().
		ApplyURI(cfg.Database.DSN()).
		SetMaxPoolSize(uint64(cfg.Database.MaxConns)).
		SetMinPoolSize(uint64(cfg.Database.MinConns)).
		SetMaxConnIdleTime(time.Duration(cfg.Database.MaxConnIdleTime))

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	// Ping db
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("db unreachable: %w", err)
	}
	app.dbclient = client
	logger.Info("Connected to MongoDb")
	repo := database.NewMongoRepo(client, cfg.Database.Dbname)

	// grpc client
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", app.cfg.Grpc.AuthService.Host, app.cfg.Grpc.AuthService.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc dial err: %w", err)
	}
	app.grpcConn = conn
	app.shutdown = append(app.shutdown, func(_ context.Context) error {
		return conn.Close()
	})
	grpcClient := userpb.NewUserServiceClient(conn)
	logger.Info("Connected to grpc auth-service server", zap.String("state", conn.GetState().String()))

	// kafka producer
	brokers := []string{fmt.Sprintf("%s:%s", app.cfg.Kafka.Addr, app.cfg.Kafka.Addr.Port)}
	userProducer := kafka.NewUserProducer(brokers, app.cfg.Kafka.Topic)

	// http handler
	mux := http.NewServeMux()
	httpHandler := handler.NewUserHandler(repo, logger, grpcClient, userProducer)
	srv := &http.Server{Addr: fmt.Sprintf(":%s", cfg.App.Port), Handler: handlers.Routes()}
	app.server = srv
	return app, nil
}
func (app *Application) Run(ctx context.Context) error {
	port := fmt.Sprintf(":%d", app.cfg.App.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	serverErr := make(chan error, 1)

	go func() {
		app.logger.Info("user-service listening", zap.String("port", port))
		if err := app.server.ListenAndServe(); err != nil {
			serverErr <- err
		}
	}()
	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		app.logger.Info("shutdown signal recv", zap.String("signal", sig.String()))
	case <-ctx.Done():
		app.logger.Info("context cancelled, shutting down")
	}
	return app.Stop(ctx)
}
func (app *Application) Stop(ctx context.Context) error {
	app.logger.Info("Shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := app.server.Shutdown(shutdownCtx); err != nil {
		app.logger.Error("http server shutdown err", zap.Error(err))
	}

	for i := len(app.shutdown) - 1; i >= 0; i-- {
		if err := app.shutdown[i](shutdownCtx); err != nil {
			app.logger.Error("shutdown error", zap.Error(err))
		}
	}
	app.logger.Info("shutdown complete")
	return nil
}
