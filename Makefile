gen:
	protoc --proto_path=proto proto/*.proto --go_out=.
	protoc --proto_path=proto proto/*.proto --go-grpc_out=.

clean:
	rm -rf pb/*.go

server:
	go run cmd/server/main.go --port 8090

client:
	go run cmd/client/main.go  --address localhost:8090

format:
	gofmt -s -w .










.PHONY: gen clean server client