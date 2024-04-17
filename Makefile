dependency:
	sudo apt-get update -y
	sudo apt-get install -y protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto:
	@protoc \
		--go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		service.proto

test:
	go test -v ./...

build: proto
	go build -o ./bin/grpcbin .

install: build
	mv ./bin/grpcbin ~/bin