module api-gateway

go 1.25.1

require (
	github.com/joho/godotenv v1.5.1
	google.golang.org/grpc v1.76.0
	bck v0.0.0-00010101000000-000000000000
)

require (
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace bck => ../grpc
