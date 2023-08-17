gen:
	protoc --proto_path=proto proto/*.proto --go_out=.
	protoc --proto_path=proto proto/*.proto --go-grpc_out=.

clean:
	rm -rf pb/*.go

run:
	go run main.go