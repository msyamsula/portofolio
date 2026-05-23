package main

// zknode is the binary that starts the ZooKeeper node.
//
// Usage:
//   go run ./cmd/zknode --port 2181 --data-dir ./data
//
// What happens when you run this:
//   1. Create the data directory (for WAL + snapshot files)
//   2. Create a Store (loads snapshot + replays WAL if they exist)
//   3. Start the gRPC server on the given port
//   4. Wait for Ctrl+C
//   5. On shutdown: take final snapshot + close WAL

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/syamsularifin/zookeeper/internal/server"
	"github.com/syamsularifin/zookeeper/internal/store"
)

func main() {
	port := flag.Int("port", 2181, "gRPC listen port")
	dataDir := flag.String("data-dir", "./data", "directory for WAL and snapshot files")
	flag.Parse()

	// Ensure data directory exists
	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create data dir: %v\n", err)
		os.Exit(1)
	}

	walPath := filepath.Join(*dataDir, "wal.log")
	snapPath := filepath.Join(*dataDir, "snapshot.json")

	// Create the Store — this recovers from existing snapshot + WAL
	s, err := store.New(walPath, snapPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create store: %v\n", err)
		os.Exit(1)
	}

	// Handle Ctrl+C (SIGINT) and container stop (SIGTERM).
	// When the signal arrives, we close the store (takes final snapshot)
	// and exit cleanly.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		fmt.Printf("\nreceived %s, shutting down...\n", sig)
		s.Close()
		os.Exit(0)
	}()

	// Start the gRPC server — this blocks forever
	srv := server.New(s, *port)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "server failed: %v\n", err)
		os.Exit(1)
	}
}
