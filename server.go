package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	pb "github.com/Anddd7/grpcbin/pb"
	"google.golang.org/grpc/credentials"
	grpc_health_pb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type ServeCmd struct{}

func loadServerTlsConfig(caCertFile, certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server key pair: %v", err)
	}

	cacertPool, err := loadCACertPool(caCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    cacertPool,
	}, nil
}

func loadCACertPool(caCertFile string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		return nil, err
	}
	ok := certPool.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, err
	}
	return certPool, nil
}

func (cmd *ServeCmd) Run(globals *Globals) error {
	port := fmt.Sprintf(":%d", globals.Port)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		slog.Error("failed to listen", "err", err)
		return err
	}

	var s *grpc.Server

	if globals.TlsCert != "" && globals.TlsKey != "" {
		tlsConfig, err := loadServerTlsConfig(globals.TlsCa, globals.TlsCert, globals.TlsKey)
		if err != nil {
			slog.Error("failed to load server TLS config", "err", err)
			return err
		}
		slog.Info("TLS enabled")

		s = grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	} else {
		s = grpc.NewServer()
	}

	reflection.Register(s)

	pb.RegisterGrpcbinServiceServer(s, &server{})
	grpc_health_pb.RegisterHealthServer(s, &server{})

	slog.Info("Server listening on", "addr", port)
	if err := s.Serve(lis); err != nil {
		slog.Error("failed to serve", "err", err)
		return err
	}

	return nil
}

type server struct {
	pb.UnimplementedGrpcbinServiceServer
	grpc_health_pb.UnimplementedHealthServer
}

func (s *server) Unary(ctx context.Context, req *pb.UnaryRequest) (*pb.UnaryResponse, error) {
	reqAttrs := req.RequestAttributes
	if reqAttrs == nil {
		reqAttrs = &pb.RequestAttributes{}
	}
	respAttrs, err := getResponseAttributes(ctx)
	if err != nil {
		return nil, err
	}

	if reqAttrs.HttpCode > 0 {
		slog.Debug("Returning HTTP status", "code", reqAttrs.HttpCode)
		if reqAttrs.HttpCode > 400 {
			return nil, status.Error(codes.Code(req.RequestAttributes.HttpCode), "HTTP status code returned")
		}
	}

	if reqAttrs.Delay > 0 {
		slog.Debug("Sleeping", "delay", reqAttrs.Delay)
		time.Sleep(time.Duration(req.RequestAttributes.Delay) * time.Second)
	}

	if reqAttrs.ResponseHeaders != nil {
		for key, value := range reqAttrs.ResponseHeaders {
			_ = grpc.SendHeader(ctx, metadata.Pairs(key, value))
		}
	}

	return &pb.UnaryResponse{
		ResponseAttributes: respAttrs,
		Result:             req.Data,
	}, nil
}

func getResponseAttributes(ctx context.Context) (*pb.ResponseAttributes, error) {
	p, _ := peer.FromContext(ctx)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get metadata")
	}

	headers := make(map[string]string)
	for key, values := range md {
		headers[key] = strings.Join(values, ",")
	}

	return &pb.ResponseAttributes{
		RequesterIp:        p.Addr.String(),
		RequesterHost:      headers["host"],
		RequesterUserAgent: headers["user-agent"],
		RequestHeaders:     headers,
	}, nil
}

func (s *server) ServerStreaming(req *pb.ServerStreamingRequest, stream pb.GrpcbinService_ServerStreamingServer) error {
	ctx := stream.Context()
	reqAttrs := req.RequestAttributes
	if reqAttrs == nil {
		reqAttrs = &pb.RequestAttributes{}
	}
	respAttrs, err := getResponseAttributes(ctx)
	if err != nil {
		return err
	}

	if reqAttrs.HttpCode > 0 {
		slog.Debug("Returning HTTP status", "code", reqAttrs.HttpCode)
		if reqAttrs.HttpCode > 400 {
			return status.Error(codes.Code(req.RequestAttributes.HttpCode), "HTTP status code returned")
		}
	}

	if reqAttrs.ResponseHeaders != nil {
		for key, value := range reqAttrs.ResponseHeaders {
			_ = grpc.SendHeader(ctx, metadata.Pairs(key, value))
		}
	}

	for i := 0; i < int(max(req.Count, 1)); i++ {
		if reqAttrs.Delay > 0 {
			slog.Debug("Sleeping", "delay", reqAttrs.Delay)
			time.Sleep(time.Duration(reqAttrs.Delay) * time.Second)
		}

		result := fmt.Sprintf("stream_%d %s", i+1, req.Data)
		response := &pb.ServerStreamingResponse{
			ResponseAttributes: respAttrs,
			Result:             result,
		}
		if err := stream.Send(response); err != nil {
			slog.Error("Failed to send response", "err", err)
			return err
		}
	}

	log.Println("Server streaming completed")

	return nil
}

func (s *server) ClientStreaming(stream pb.GrpcbinService_ClientStreamingServer) error {
	ctx := stream.Context()
	// avoid nil pointer if no request
	reqAttrs := &pb.RequestAttributes{}
	respAttrs, err := getResponseAttributes(ctx)
	if err != nil {
		return err
	}

	var receivedData []string
	for {
		req, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			slog.Error("Error receiving data from client", "err", err)
			return err
		}

		reqAttrs = req.RequestAttributes
		if reqAttrs == nil {
			reqAttrs = &pb.RequestAttributes{}
		}
		receivedData = append(receivedData, req.Data)
	}

	if reqAttrs.HttpCode > 0 {
		slog.Debug("Returning HTTP status", "code", reqAttrs.HttpCode)
		if reqAttrs.HttpCode > 400 {
			return status.Error(codes.Code(reqAttrs.HttpCode), "HTTP status code returned")
		}
	}

	if reqAttrs.Delay > 0 {
		slog.Debug("Sleeping", "delay", reqAttrs.Delay)
		time.Sleep(time.Duration(reqAttrs.Delay) * time.Second)
	}

	if reqAttrs.ResponseHeaders != nil {
		for key, value := range reqAttrs.ResponseHeaders {
			_ = grpc.SendHeader(ctx, metadata.Pairs(key, value))
		}
	}

	response := &pb.ClientStreamingResponse{
		ResponseAttributes: respAttrs,
		Result:             strings.Join(receivedData, ", "),
	}
	if err := stream.SendAndClose(response); err != nil {
		log.Printf("Error sending response to client: %v", err)
		return err
	}

	log.Println("Client streaming completed")

	return nil
}

func (s *server) BidirectionalStreaming(stream pb.GrpcbinService_BidirectionalStreamingServer) error {
	ctx := stream.Context()
	respAttrs, err := getResponseAttributes(ctx)
	if err != nil {
		return err
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				log.Printf("Error receiving data from client: %v", err)
				return err
			}
		}

		reqAttrs := req.RequestAttributes
		if reqAttrs == nil {
			reqAttrs = &pb.RequestAttributes{}
		}

		if reqAttrs.HttpCode > 0 {
			slog.Debug("Returning HTTP status", "code", reqAttrs.HttpCode)
			if reqAttrs.HttpCode > 400 {
				return status.Error(codes.Code(reqAttrs.HttpCode), "HTTP status code returned")
			}
		}

		if reqAttrs.Delay > 0 {
			slog.Debug("Sleeping", "delay", reqAttrs.Delay)
			time.Sleep(time.Duration(reqAttrs.Delay) * time.Second)
		}

		if reqAttrs.ResponseHeaders != nil {
			for key, value := range reqAttrs.ResponseHeaders {
				_ = grpc.SendHeader(ctx, metadata.Pairs(key, value))
			}
		}

		response := &pb.BidirectionalStreamingResponse{
			ResponseAttributes: respAttrs,
			Result:             req.Data,
		}
		if err := stream.Send(response); err != nil {
			slog.Error("Failed to send response", "err", err)
			return err
		}
	}

	return nil
}

func (s *server) Check(ctx context.Context, req *grpc_health_pb.HealthCheckRequest) (*grpc_health_pb.HealthCheckResponse, error) {
	return &grpc_health_pb.HealthCheckResponse{
		Status: grpc_health_pb.HealthCheckResponse_SERVING,
	}, nil
}

func (s *server) Watch(req *grpc_health_pb.HealthCheckRequest, stream grpc_health_pb.Health_WatchServer) error {
	for {
		if err := stream.Send(&grpc_health_pb.HealthCheckResponse{Status: grpc_health_pb.HealthCheckResponse_SERVING}); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
	}
}
