module user-service

go 1.25.3

require (
	bck v0.0.0-00010101000000-000000000000
	github.com/joho/godotenv v1.5.1
	github.com/segmentio/kafka-go v0.4.49
	go.mongodb.org/mongo-driver v1.17.6
	go.mongodb.org/mongo-driver/v2 v2.4.0
	golang.org/x/crypto v0.43.0
	google.golang.org/grpc v1.76.0
)

require (
	github.com/golang/snappy v1.0.0 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/net v0.45.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace bck => ../grpc
