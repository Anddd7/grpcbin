FROM golang:1.22.3-alpine3.19 AS builder

WORKDIR /app

RUN apk update && apk add protobuf

RUN GO111MODULE=on && \
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN GRPC_HEALTH_PROBE_VERSION=v0.4.13 && \
  wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
  chmod +x /bin/grpc_health_probe

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  pb/service.proto

RUN go build -o grpcbin .

FROM alpine:3.19

COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
COPY --from=builder /app/grpcbin /grpcbin

COPY certs /certs

EXPOSE 50051

ENTRYPOINT ["/grpcbin"]
CMD ["serve"]
