FROM golang:1.26.3-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY user-service/go.mod user-service/go.sum ./user-service/

RUN cd user-service && go mod download
COPY grpc ./grpc

COPY user-service ./user-service/

RUN cd user-service && CGO_ENABLED=0 GOOS=linux go build -o /bin/user-service ./cmd/server/main.go

FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /bin/user-service /bin/user-service

EXPOSE 8000

CMD ["/bin/user-service"]
