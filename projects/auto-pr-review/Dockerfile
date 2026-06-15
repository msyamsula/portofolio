FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/job-orchestrator ./cmd/job-orchestrator
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/job-consumer ./cmd/job-consumer

FROM alpine:3.20

RUN apk add --no-cache git

RUN adduser -D -H -u 10001 appuser
USER appuser

COPY --from=build /out/api /api
COPY --from=build /out/job-orchestrator /job-orchestrator
COPY --from=build /out/job-consumer /job-consumer

EXPOSE 8080

CMD ["/api"]
