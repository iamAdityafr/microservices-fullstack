FROM golang:1.26.3-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY payment-service/go.mod payment-service/go.sum ./payment-service/

RUN cd payment-service && go mod download
COPY grpc ./grpc

COPY payment-service ./payment-service/

RUN cd payment-service && CGO_ENABLED=0 GOOS=linux go build -o /bin/payment-service ./cmd/server/main.go

FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /bin/payment-service /bin/payment-service

EXPOSE 8000

CMD ["/bin/payment-service"]
