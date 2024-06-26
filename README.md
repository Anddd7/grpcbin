# Grpcbin

Similar with httpbin(<https://httpbin.org/>) to test the grpc connection.

## Quick Start

Binary

```sh
# install binary
export VERSION="1.1.0"
export PLATFORM="linux_amd64"

curl -sSL https://github.com/Anddd7/grpcbin/releases/download/v${VERSION}/grpcbin_v${VERSION}_${PLATFORM}.tar.gz -o grpcbin.tar.gz
tar -xzf grpcbin.tar.gz grpcbin

# start server
./grpcbin serve

# send grpc call
./grpcbin unary --message hello
```

Docker

```sh
# pull docker image
docker pull ghcr.io/anddd7/grpcbin:latest

# start server
docker run -d -p 50051:50051 ghcr.io/anddd7/grpcbin
# send grpc call to server
docker run -it ghcr.io/anddd7/grpcbin unary --message hello --server <server_container_ip>

# you can get server ip via 
docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $(docker ps -q)
```

### TLS Support

> from v1.1.0, it support tls connection. \
> you can use `--tls-cert` and `--tls-key` to define the cert and key file for both server and client.

Built-in self-signed certificate (see `certs`)

```sh
# start server with tls cert
docker run -d -p 50051:50051 ghcr.io/anddd7/grpcbin serve --tls-cert="/certs/server.crt" --tls-key="/certs/server.key"
# send grpc call to server
docker run -it ghcr.io/anddd7/grpcbin unary --message hello --server <server_container_ip> --tls-cert="/certs/server.crt"
```

Use your own certificate

```sh
# start server with tls cert
docker run -d -p 50051:50051 -v /path/to/your/cert:/certs ghcr.io/anddd7/grpcbin serve --tls-cert="/certs/server.crt" --tls-key="/certs/server.key"
# send grpc call to server
docker run -it -v /path/to/your/cert:/certs ghcr.io/anddd7/grpcbin unary --message hello --server <server_container_ip> --tls-cert="/certs/server.crt"
```

## Usage

**RequestAttributes** is used to contol how server handle the request, including:

- http_code: returns specific http status code, only <400 returns body
- delay: delay time for server to response
- response_headers: add headers to response

**ResponseAttributes** is used to inspect the request metadata, including:

- requester_ip
- requester_host
- requester_user_agent
- request_headers

### Unary

Send a data and get a result

```sh
grpcbin unary --message hello

# delay 5s
grpcbin unary --message hello --delay 5

# add custom headers in both request and response
grpcbin unary --message hello --response-headers=responder=anddd7 --headers=caller=anddd7
```

### Server Streaming

Send a data and get a stream of results

```sh
grpcbin server-streaming --message hello

# get 5 messages
grpcbin server-streaming --message hello --count 5

# delay 2s for each message
grpcbin server-streaming --message hello --count 5 --delay 2
```

### Client Streaming

Send a stream of data and get a result

```sh
grpcbin client-streaming --message hello

# send 5 messages
grpcbin client-streaming --message hello --count 5

# delay 2s for server response
grpcbin client-streaming --message hello --count 5 --delay 2
```

### Bidirectional Streaming

Send a stream of data and get a stream of results

```sh
grpcbin bidirectional-streaming --message hello

# send and get 5 messages
grpcbin bidirectional-streaming --message hello --count 5

# delay 2s for each round
grpcbin bidirectional-streaming --message hello --count 5 --delay 2
```

## Tips

You you want to use it in kubernetes, check example and tutorial [here](https://github.com/Anddd7/rubber-duck/tree/main/grpcbin).