# 05 - gRPC Server

## The Problem

The Store works, but only from Go code in the same process. ZooKeeper is a network service — clients connect from different machines, possibly written in different languages. We need a way to call Create/Get/Set/Delete over the network.

## The Solution: gRPC + Protocol Buffers

gRPC is a framework for remote procedure calls. It lets a client call a function on a server as if it were a local function call, but the data travels over the network.

Protocol Buffers (protobuf) is the serialization format. We define our API in a `.proto` file, and a code generator (`protoc`) creates:

1. Go structs for requests and responses (`CreateRequest`, `GetResponse`, etc.)
2. A server interface we must implement
3. A client that can connect to our server

## The Proto File

```protobuf
service ZooKeeper {
  rpc Create(CreateRequest) returns (CreateResponse);
  rpc Get(GetRequest) returns (GetResponse);
  rpc Set(SetRequest) returns (SetResponse);
  rpc Delete(DeleteRequest) returns (DeleteResponse);
  rpc GetChildren(GetChildrenRequest) returns (GetChildrenResponse);
}
```

This defines five RPCs. Each takes a request message and returns a response message. From this, protoc generates ~500 lines of Go code that handles serialization, networking, and connection management.

## The Server - A Thin Bridge

The server's only job is to translate between gRPC and Store:

```
gRPC CreateRequest  →  store.Create()  →  gRPC CreateResponse
gRPC GetRequest     →  store.Get()     →  gRPC GetResponse
```

Each method is ~5 lines:

```go
func (s *Server) Get(ctx context.Context, req *zkpb.GetRequest) (*zkpb.GetResponse, error) {
    data, err := s.store.Get(req.Path)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "%v", err)
    }
    return &zkpb.GetResponse{Data: data}, nil
}
```

No business logic. The Store does the real work. This separation means:

- Store can be tested without a network
- Server could be swapped for HTTP or WebSocket without changing Store
- Each layer has one job

## Error Codes

gRPC has standard error codes. We map our errors to appropriate codes:

| Store error | gRPC code | Meaning |
|-------------|-----------|---------|
| "already exists" | AlreadyExists | Create on existing path |
| "not found" | NotFound | Get/Set on missing path |
| "has children" | FailedPrecondition | Delete on non-leaf node |

These codes let clients handle errors programmatically without parsing error strings.

## The Full Request Flow

```
Client (zkcli)                         Server (zknode)
     │                                      │
     │  go run zkcli create /app "hello"    │
     │                                      │
     │  1. Build CreateRequest{             │
     │       Path: "/app",                  │
     │       Data: "hello"                  │
     │     }                                │
     │                                      │
     │  2. Serialize to protobuf bytes      │
     │                                      │
     │  3. Send over TCP ──────────────→    │
     │                                      │
     │                              4. Deserialize request
     │                              5. server.Create() called
     │                              6. store.Create("/app", "hello")
     │                                   ├── WAL.Append()
     │                                   └── tree.Create()
     │                              7. Build CreateResponse{Path: "/app"}
     │                              8. Serialize to protobuf bytes
     │                                      │
     │  ←───────────────── 9. Send over TCP │
     │                                      │
     │  10. Deserialize response            │
     │  11. Print "created /app"            │
```

## The Main Binary (zknode)

```go
func main() {
    store := store.New(walPath, snapPath)   // recover from disk
    server := server.New(store, port)        // create gRPC server
    server.Start()                           // listen forever

    // on Ctrl+C:
    store.Close()                            // take final snapshot
}
```

Three steps: create store (recovery happens here), create server, start listening. On shutdown, close the store (takes a final snapshot).

## The CLI Client (zkcli)

```
zkcli --server localhost:2181 create /app "hello"

  1. Connect to localhost:2181 via gRPC
  2. Parse command: "create" → call client.Create()
  3. Build CreateRequest, send it
  4. Receive CreateResponse, print result
```

The client is stateless. It connects, makes one call, prints the result, and exits. In later phases, we'll add persistent sessions with heartbeats.

## Files

- `api/proto/zk.proto` - service definition
- `api/proto/zkpb/` - generated Go code (do not edit)
- `internal/server/server.go` - gRPC server implementation
- `cmd/zknode/main.go` - server binary
- `cmd/zkcli/main.go` - CLI client
