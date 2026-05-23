# ZooKeeper — A Distributed Coordination Service in Go

A learning project that builds a ZooKeeper-like distributed coordination service from scratch.

## What is ZooKeeper?

ZooKeeper is a centralized service for distributed coordination. Think of it as a tiny, reliable filesystem that lives across multiple servers. Distributed systems use it to answer questions like:

- Which server is the current leader?
- What's the database connection string?
- Which services are alive right now?
- Who holds the distributed lock?

## Project Roadmap

| Phase | Status | Description |
|-------|--------|-------------|
| 1 - Single Node | Done | Znode tree, WAL, snapshots, gRPC API |
| 2 - Cluster & Leader Election | Planned | Raft consensus, leader-follower replication |
| 3 - Log Shipping & Node Join | Planned | Snapshot transfer, new node catch-up |
| 4 - Sessions & Watches | Planned | Ephemeral nodes, watch notifications |
| 5 - Distributed Primitives | Planned | Locks, leader election, service discovery |
| 6 - Deployment | Planned | Docker, Kubernetes StatefulSet |

## Quick Start

Start the node:

```bash
go run ./cmd/zknode --port 2181 --data-dir ./data
```

Use the CLI:

```bash
go run ./cmd/zkcli --server localhost:2181 create /app "hello"
go run ./cmd/zkcli --server localhost:2181 get /app
go run ./cmd/zkcli --server localhost:2181 set /app "world"
go run ./cmd/zkcli --server localhost:2181 ls /
go run ./cmd/zkcli --server localhost:2181 delete /app
```

Run tests:

```bash
go test ./... -v
```

## Project Structure

```
cmd/
  zknode/main.go           server binary
  zkcli/main.go            CLI client

internal/
  znode/                   in-memory data tree
    znode.go               ZNode struct (data + children)
    tree.go                DataTree (Create, Get, Set, Delete, GetChildren, snapshot methods)
    tree_test.go           14 tests

  wal/                     write-ahead log
    wal.go                 Entry struct, WAL (Open, Append, ReadAll)
    wal_test.go            3 tests

  snapshot/                point-in-time tree dump
    snapshot.go            Snapshot struct, Save, Load
    snapshot_test.go       2 tests

  store/                   coordinator (WAL + tree + snapshot)
    store.go               Store (recovery, Create, Get, Set, Delete, TakeSnapshot)
    store_test.go          4 tests

  server/                  gRPC server
    server.go              thin bridge: gRPC request -> Store -> gRPC response

api/proto/
  zk.proto                 gRPC service definition
  zkpb/                    generated Go code
```

## Learning Documentation

1. [Data Model](docs/01-data-model.md) - What are znodes and why are they a tree
2. [Write-Ahead Log](docs/02-write-ahead-log.md) - How changes survive crashes
3. [Snapshots](docs/03-snapshots.md) - How to recover fast
4. [Store](docs/04-store.md) - How WAL, tree, and snapshot work together
5. [gRPC Server](docs/05-grpc-server.md) - How clients talk to the node over the network
