package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/alecthomas/kong"
)

type Globals struct {
	Debug  bool   `short:"d" help:"enable debug mode"`
	Server string `default:"localhost" help:"server address"`
	Port   int    `default:"50051" help:"server port"`
	Host   string `default:"localhost" help:"server host, used for load balancing"`
}

type CLI struct {
	Globals

	Serve                  ServeCmd                  `cmd:"" help:"start a gRPC server"`
	Unary                  UnaryCmd                  `cmd:"" help:"perform a unary call"`
	ServerStreaming        ServerStreamingCmd        `cmd:"" aliases:"srvstr" help:"perform a server streaming call"`
	ClientStreaming        ClientStreamingCmd        `cmd:"" aliases:"clistr" help:"perform a client streaming call"`
	BidirectionalStreaming BidirectionalStreamingCmd `cmd:"" aliases:"bistr" help:"perform a bidirectional streaming call"`
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("grpcbin"),
		kong.Description("A gRPC server and client for testing"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": "v1.0.1",
		},
	)

	if cli.Globals.Debug {
		opts := slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &opts)))
	}

	startTime := time.Now()

	err := ctx.Run(&cli.Globals)
	if err != nil {
		slog.Error("failed", "err", err)
	}

	slog.Info("Execution", "time", time.Since(startTime))
}
