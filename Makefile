NAME					:= grpcbin
DIST					:= ./dist
NEXT_VERSION	:= $(shell semtag final -o)

dep:
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	sudo apt-get update -y
	sudo apt-get install -y protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto:
	@protoc \
		--go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pb/service.proto

build: proto
	go build -o $(DIST)/$(NAME) ./

clean:
	rm -rf $(DIST)

fmt:
	go fmt ./...
	gofumpt -l -w .
	go vet ./...

test: proto
	go test -v

lint:
	golangci-lint run -v

cover:
	go test -coverprofile coverage.out

coverweb: cover
	go tool cover -html=coverage.out

check: fmt lint cover

install: build
	mkdir -p ~/bin
	mv $(DIST)/$(NAME) ~/bin/$(NAME) 

uninstall:
	rm ~/bin/$(NAME)

cliversion:
	echo "increasing version to $(NEXT_VERSION)"
	@sed -i 's/"version": "v[0-9]*\.[0-9]*\.[0-9]*"/"version": "$(NEXT_VERSION)"/' main.go
	@git add main.go
	@git commit -m "Auto Release - $(NEXT_VERSION)"

release: cliversion
	@git tag $(NEXT_VERSION)
	echo "pushing to origin"
	@git push origin main
	@git push origin $(NEXT_VERSION)