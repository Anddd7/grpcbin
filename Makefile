NAME					:= grpcbin
DIST					:= ./dist
NEXT_VERSION	:= $(shell semtag final -o)

dep:
	sudo apt-get update -y
	sudo apt-get install -y protobuf-compiler
	go install \
		github.com/golangci/golangci-lint/cmd/golangci-lint \
		mvdan.cc/gofumpt \
		github.com/bufbuild/buf/cmd/buf \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		google.golang.org/protobuf/cmd/protoc-gen-go \

proto:
	@buf generate

build: proto
	go build -o $(DIST)/$(NAME) ./

clean:
	rm -rf $(DIST)

fmt:
	go fmt ./...
	gofumpt -l -w .
	go vet ./...

test: proto
	go test -v ./...

lint:
	golangci-lint run -v

cover:
	go test -coverprofile coverage.out ./...

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
