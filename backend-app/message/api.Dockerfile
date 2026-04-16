# syntax=docker/dockerfile:1
# Build context: backend-app/message/ (the isolated Go module root)

FROM golang:1.22-alpine AS builder

WORKDIR /src

# Install git for modules that require VCS metadata.
RUN apk add --no-cache ca-certificates git

# Layer-cache trick: download dependencies before copying source so that
# source-only changes don't invalidate the dependency download layer.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w" \
    -o /out/api ./cmd/api

# ─────────────────────────────────────────────────────────────────────────────
# Runtime: minimal static image — no shell, no libc, no attack surface.
# ─────────────────────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /out/api /api

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/api"]
