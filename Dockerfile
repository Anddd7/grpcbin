FROM golang:1.22.3-alpine3.19 AS builder

WORKDIR /app

RUN apk update && apk add protobuf

RUN GRPC_HEALTH_PROBE_VERSION=v0.4.13 && \
  wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
  chmod +x /bin/grpc_health_probe

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN GO111MODULE=on && go install \
  github.com/bufbuild/buf/cmd/buf \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
  google.golang.org/grpc/cmd/protoc-gen-go-grpc \
  google.golang.org/protobuf/cmd/protoc-gen-go

COPY . .

RUN buf generate

RUN go build -o grpcbin .

FROM alpine:3.20

COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
COPY --from=builder /app/grpcbin /grpcbin

COPY certs /certs

EXPOSE 50051

ENTRYPOINT ["/grpcbin"]
CMD ["serve"]
