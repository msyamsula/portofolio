# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w" -o /out/http-server ./backend-app/binary/http

FROM gcr.io/distroless/static:nonroot
WORKDIR /

COPY --from=builder /out/http-server /http-server

EXPOSE 5000
USER nonroot:nonroot

ENTRYPOINT ["/http-server"]
