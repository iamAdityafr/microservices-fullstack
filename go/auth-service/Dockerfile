FROM golang:1.26.3-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY auth-service/go.mod auth-service/go.sum ./auth-service/

RUN cd auth-service && go mod download
COPY grpc ./grpc

COPY auth-service ./auth-service/

RUN cd auth-service && CGO_ENABLED=0 GOOS=linux go build -o /bin/auth-service ./cmd/server/main.go

FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /bin/auth-service /bin/auth-service

EXPOSE 8000

CMD ["/bin/auth-service"]
