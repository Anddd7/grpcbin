package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	pb "github.com/Anddd7/grpcbin/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func connect(serverAddr string, host string, tlsCert string) (pb.GrpcbinServiceClient, error) {
	var err error
	var creds credentials.TransportCredentials
	if tlsCert != "" {
		creds, err = credentials.NewClientTLSFromFile(tlsCert, "")
		if err != nil {
			slog.Error("failed to load TLS certificate", "err", err)
			return nil, err
		}
	} else {
		creds = insecure.NewCredentials()
	}

	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(creds),
		grpc.WithAuthority(host),
	)
	if err != nil {
		slog.Error("failed to connect", "err", err)
		return nil, err
	}
	// defer conn.Close()
	client := pb.NewGrpcbinServiceClient(conn)

	return client, nil
}

type UnaryCmd struct {
	pb.RequestAttributes
	Message string
	Headers map[string]string
}

func (cmd *UnaryCmd) Run(globals *Globals) error {
	client, err := connect(fmt.Sprintf("%s:%d", globals.Server, globals.Port), globals.Host, globals.TlsCert)
	if err != nil {
		return err
	}

	var headers, trailers metadata.MD
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(cmd.Headers))
	response, err := client.Unary(ctx, &pb.UnaryRequest{
		RequestAttributes: &cmd.RequestAttributes,
		Data:              cmd.Message,
	},
		grpc.Header(&headers),
		grpc.Trailer(&trailers),
	)
	if err != nil {
		slog.Error("Unary RPC failed", "err", err)
		return err
	}

	printResponse(headers, trailers, response.Result, response.ResponseAttributes)

	return nil
}

func printResponse(headers, trailers metadata.MD, result string, respAttrs *pb.ResponseAttributes) {
	slog.Info(">>---->>---->>---->>---->>---->>")

	slog.Info("Request Inspection")
	slog.Info("-", "requester_ip", respAttrs.RequesterIp)
	slog.Info("-", "requester_user_agent", respAttrs.RequesterUserAgent)
	slog.Info("-", "requester_host", respAttrs.RequesterHost)

	slog.Info("Request Headers")
	for key, value := range respAttrs.RequestHeaders {
		slog.Info("-", key, value)
	}

	slog.Info("<<----<<----<<----<<----<<----<<")

	if len(headers) > 0 {
		slog.Info("Response Headers")
		for key, values := range headers {
			slog.Info("-", key, strings.Join(values, ","))
		}
	}

	if len(trailers) > 0 {
		slog.Info("Response Trailers")
		for key, values := range trailers {
			slog.Info("-", key, strings.Join(values, ","))
		}
	}

	slog.Info("Response Body")
	slog.Info("-", "response", result)
}

type ServerStreamingCmd struct {
	pb.RequestAttributes
	Message string
	Count   int32 `default:"1"`
	Headers map[string]string
}

func (cmd *ServerStreamingCmd) Run(globals *Globals) error {
	client, err := connect(fmt.Sprintf("%s:%d", globals.Server, globals.Port), globals.Host, globals.TlsCert)
	if err != nil {
		return err
	}

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(cmd.Headers))
	stream, err := client.ServerStreaming(ctx, &pb.ServerStreamingRequest{
		RequestAttributes: &cmd.RequestAttributes,
		Data:              cmd.Message,
		Count:             cmd.Count,
	})
	if err != nil {
		slog.Error("Server Streaming RPC failed", "err", err)
		return err
	}

	for {
		response, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				slog.Error("Server Streaming RPC stream closed", "err", err)
				return err
			}
			return nil
		}
		headers, _ := stream.Header()
		trailers := stream.Trailer()
		printResponse(headers, trailers, response.Result, response.ResponseAttributes)
	}
}

type ClientStreamingCmd struct {
	pb.RequestAttributes
	Message string
	Count   int32 `default:"1"`
	Headers map[string]string
}

func (cmd *ClientStreamingCmd) Run(globals *Globals) error {
	client, err := connect(fmt.Sprintf("%s:%d", globals.Server, globals.Port), globals.Host, globals.TlsCert)
	if err != nil {
		return err
	}

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(cmd.Headers))
	stream, err := client.ClientStreaming(ctx)
	if err != nil {
		slog.Error("Client Streaming RPC failed", "err", err)
		return err
	}
	for i := 0; i < int(cmd.Count); i++ {
		err := stream.Send(&pb.ClientStreamingRequest{
			RequestAttributes: &cmd.RequestAttributes,
			Data:              cmd.Message,
		})
		if err != nil {
			slog.Error("Failed to send client streaming request", "err", err)
			return err
		}
	}
	response, err := stream.CloseAndRecv()
	if err != nil {
		slog.Error("Failed to receive client streaming response", "err", err)
		return err
	}

	headers, _ := stream.Header()
	trailers := stream.Trailer()
	printResponse(headers, trailers, response.Result, response.ResponseAttributes)

	return nil
}

type BidirectionalStreamingCmd struct {
	pb.RequestAttributes
	Message string
	Count   int32 `default:"1"`
	Headers map[string]string
}

func (cmd *BidirectionalStreamingCmd) Run(globals *Globals) error {
	client, err := connect(fmt.Sprintf("%s:%d", globals.Server, globals.Port), globals.Host, globals.TlsCert)
	if err != nil {
		return err
	}

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(cmd.Headers))
	stream, err := client.BidirectionalStreaming(ctx)
	if err != nil {
		slog.Error("Bidirectional Streaming RPC failed", "err", err)
		return err
	}
	for i := 0; i < int(cmd.Count); i++ {
		err := stream.Send(&pb.BidirectionalStreamingRequest{
			RequestAttributes: &cmd.RequestAttributes,
			Data:              fmt.Sprintf("stream_%d %s", i+1, cmd.Message),
		})
		if err != nil {
			slog.Error("Failed to send bidirectional streaming request", "err", err)
			return err
		}
		response, err := stream.Recv()
		if err != nil {
			slog.Error("Failed to receive bidirectional streaming response", "err", err)
			return err
		}

		headers, _ := stream.Header()
		trailers := stream.Trailer()
		printResponse(headers, trailers, response.Result, response.ResponseAttributes)
	}

	err = stream.CloseSend()
	if err != nil {
		slog.Error("Failed to close bidirectional streaming", "err", err)
	}

	return nil
}
