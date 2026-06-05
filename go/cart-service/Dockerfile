FROM golang:1.26.3-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY cart-service/go.mod cart-service/go.sum ./cart-service/

RUN cd cart-service && go mod download
COPY grpc ./grpc

COPY cart-service ./cart-service/

RUN cd cart-service && CGO_ENABLED=0 GOOS=linux go build -o /bin/cart-service ./cmd/server/main.go

FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /bin/cart-service /bin/cart-service

EXPOSE 8000

CMD ["/bin/cart-service"]
