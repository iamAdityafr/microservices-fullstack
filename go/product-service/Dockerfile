FROM golang:1.26.3-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY product-service/go.mod product-service/go.sum ./product-service/

RUN cd product-service && go mod download
COPY grpc ./grpc

COPY product-service ./product-service/

RUN cd product-service && CGO_ENABLED=0 GOOS=linux go build -o /bin/product-service ./cmd/server/main.go

FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /bin/product-service /bin/product-service

EXPOSE 8000

CMD ["/bin/product-service"]
