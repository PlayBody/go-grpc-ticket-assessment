default:
	go test -v ./...
	go build

protoc:
	protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. proto/train.proto

build:
	go test -v ./...
	go build -ldflags="-s -w"

test:
	go test -v ./...

run:
	go test -v ./...
	go run github.com/playbody/train-ticket-service
