FROM golang:1.26.3-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/

RUN cd api-gateway && go mod download
COPY grpc ./grpc

COPY api-gateway ./api-gateway/

RUN cd api-gateway && CGO_ENABLED=0 GOOS=linux go build -o /bin/api-gateway ./cmd/server/main.go

FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /bin/api-gateway /bin/api-gateway

EXPOSE 8000

CMD ["/bin/api-gateway"]
