# Grpcbin

Similar with httpbin(<https://httpbin.org/>) to test the grpc connection.

## Usage

Binary

```sh
# install binary
curl -sSL https://github.com/Anddd7/grpcbin/releases/download/v0.0.1/grpcbin -o grpcbin
chmod +x grpcbin

# start server
grpcbin serve

# send grpc call
grpcbin unary --message hello --response-headers=responder=anddd7 --headers=caller=anddd7
```

Docker

```sh
# pull docker image
docker pull ghcr.io/anddd7/grpcbin:latest

# start server
docker run -d -p 50051:50051 ghcr.io/anddd7/grpcbin
# send grpc call
docker run -it ghcr.io/anddd7/grpcbin ./grpcbin unary --message hello --response-headers=responder=anddd7 --headers=caller=anddd7 --host <server_container_ip>

# you can get server ip via 
docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $(docker ps -q)
```

## TODO

- [ ] streaming api
- [ ] grpcs support
