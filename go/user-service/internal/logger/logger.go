package logger

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(ctx context.Context, serviceName string) (*zap.Logger, func(context.Context) error, error) {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint("user-service:4318"), // keep hardcoded as of now
		otlploggrpc.WithInsecure(),
	}

	exp, err := otlploggrpc.New(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("otlp log exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("otel resource: %w", err)
	}

	lp := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exp)),
		log.WithResource(res),
	)

	otlpCore := otelzap.NewCore(serviceName, otelzap.WithLoggerProvider(lp))
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		zap.DebugLevel,
	)

	logger := zap.New(zapcore.NewTee(otlpCore, consoleCore), zap.AddCaller())
	return logger, lp.Shutdown, nil
}
