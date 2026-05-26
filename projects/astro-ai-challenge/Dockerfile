FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o calendar-ai .

FROM node:22-alpine AS claude-installer
RUN npm install -g @anthropic-ai/claude-code

FROM alpine:3.22

RUN apk add --no-cache ca-certificates libgcc libstdc++

COPY --from=builder /app/calendar-ai /usr/local/bin/calendar-ai
COPY --from=builder /app/static /app/static
COPY --from=claude-installer /usr/local/lib/node_modules/@anthropic-ai/claude-code/bin/claude.exe /usr/local/bin/claude

WORKDIR /app

ENV PORT=8080
ENV HOST=0.0.0.0
ENV CLAUDE_PATH=claude

EXPOSE 8080

CMD ["sh", "-c", "set -a && . /etc/secrets/.env && set +a && calendar-ai"]
