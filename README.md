# Grpcbin

Similar with httpbin(<https://httpbin.org/>) to test the grpc connection.

## Usage

```sh
# install binary
grpcbin -h

# start server
grpcbin serve
# send grpc call
grpcbin unary --message hello --response-headers=responder=anddd7 --headers=caller=anddd7


# pull docker image
docker pull ghcr.io/anddd7/grpcbin

# start server
docker run -d -p 50051:50051 ghcr.io/anddd7/grpcbin
# send grpc call
docker run -it grpcbin ./grpcbin unary --message hello --response-headers=responder=anddd7 --headers=caller=anddd7 --host <server_container_ip>
```

## TODO

- [ ] grpcs support
