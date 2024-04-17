FROM golang:1.22.2-alpine3.19 AS builder

WORKDIR /app

RUN apk update && apk add protobuf

RUN GO111MODULE=on && \
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN protoc \
  --go_out=pb --go_opt=paths=source_relative \
  --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
  service.proto

RUN go build -o grpcbin .

FROM alpine:3.19

COPY --from=builder /app/grpcbin /grpcbin

EXPOSE 50051

CMD ["/grpcbin", "serve"]
