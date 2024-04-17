# docker image for golang cli tool
# linux, go 1.22.0

FROM golang:1.22.2-alpine3.19 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o grpcbin .

FROM alpine:3.19

COPY --from=builder /app/grpcbin /grpcbin

EXPOSE 50051

CMD ["grpcbin", "serve"]
