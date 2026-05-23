package main

// zkcli is a command-line client that talks to a running zknode via gRPC.
//
// Usage:
//   go run ./cmd/zkcli --server localhost:2181 create /app "hello"
//   go run ./cmd/zkcli --server localhost:2181 get /app
//   go run ./cmd/zkcli --server localhost:2181 set /app "world"
//   go run ./cmd/zkcli --server localhost:2181 delete /app
//   go run ./cmd/zkcli --server localhost:2181 ls /

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/syamsularifin/zookeeper/api/proto/zkpb"
)

func main() {
	if len(os.Args) < 4 {
		printUsage()
		os.Exit(1)
	}

	if os.Args[1] != "--server" {
		printUsage()
		os.Exit(1)
	}

	serverAddr := os.Args[2]
	command := os.Args[3]
	args := os.Args[4:]

	// Connect to the server.
	// insecure.NewCredentials() means no TLS — fine for local development.
	// In production, you'd use TLS certificates.
	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Create the gRPC client from the generated code.
	client := zkpb.NewZooKeeperClient(conn)

	// 5 second timeout for all operations.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch command {
	case "create":
		cmdCreate(ctx, client, args)
	case "get":
		cmdGet(ctx, client, args)
	case "set":
		cmdSet(ctx, client, args)
	case "delete":
		cmdDelete(ctx, client, args)
	case "ls":
		cmdLs(ctx, client, args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func cmdCreate(ctx context.Context, c zkpb.ZooKeeperClient, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: create <path> [data]")
		os.Exit(1)
	}

	req := &zkpb.CreateRequest{Path: args[0]}
	if len(args) >= 2 {
		req.Data = []byte(args[1])
	}

	resp, err := c.Create(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("created %s\n", resp.Path)
}

func cmdGet(ctx context.Context, c zkpb.ZooKeeperClient, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: get <path>")
		os.Exit(1)
	}

	resp, err := c.Get(ctx, &zkpb.GetRequest{Path: args[0]})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(resp.Data))
}

func cmdSet(ctx context.Context, c zkpb.ZooKeeperClient, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: set <path> <data>")
		os.Exit(1)
	}

	_, err := c.Set(ctx, &zkpb.SetRequest{Path: args[0], Data: []byte(args[1])})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("updated")
}

func cmdDelete(ctx context.Context, c zkpb.ZooKeeperClient, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: delete <path>")
		os.Exit(1)
	}

	_, err := c.Delete(ctx, &zkpb.DeleteRequest{Path: args[0]})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("deleted")
}

func cmdLs(ctx context.Context, c zkpb.ZooKeeperClient, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: ls <path>")
		os.Exit(1)
	}

	resp, err := c.GetChildren(ctx, &zkpb.GetChildrenRequest{Path: args[0]})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(resp.Children) == 0 {
		fmt.Println("(no children)")
		return
	}
	for _, child := range resp.Children {
		fmt.Println(child)
	}
}

func printUsage() {
	fmt.Println("usage: zkcli --server <addr> <command> [args]")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  create <path> [data]    create a znode")
	fmt.Println("  get    <path>           read a znode's data")
	fmt.Println("  set    <path> <data>    update a znode's data")
	fmt.Println("  delete <path>           delete a znode")
	fmt.Println("  ls     <path>           list children")
}
