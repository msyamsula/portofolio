# syntax=docker/dockerfile:1
# Build context: backend-app/message/ (the isolated Go module root)

FROM golang:1.22-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w" \
    -o /out/reconciler ./cmd/reconciler

# ─────────────────────────────────────────────────────────────────────────────
# Runtime
# ─────────────────────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /out/reconciler /reconciler

USER nonroot:nonroot

ENTRYPOINT ["/reconciler"]
