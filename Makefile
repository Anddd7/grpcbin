NEXT_VERSION:=$(shell semtag final -o)

dependency:
	sudo apt-get update -y
	sudo apt-get install -y protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto:
	@protoc \
		--go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pb/service.proto

test:
	go test -v ./...

build: proto
	go build -o ./bin/grpcbin .

install: build
	mv ./bin/grpcbin ~/bin

release:
	echo "increasing version to $(NEXT_VERSION)"
	@sed -i 's/"version": "v[0-9]*\.[0-9]*\.[0-9]*"/"version": "$(NEXT_VERSION)"/' main.go
	@git add main.go
	@git commit -m "Auto Release - $(NEXT_VERSION)"
	@git tag $(NEXT_VERSION)
	echo "pushing to origin"
	@git push origin main
	@git push origin $(NEXT_VERSION)