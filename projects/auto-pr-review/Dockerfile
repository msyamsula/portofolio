FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/job-orchestrator ./cmd/job-orchestrator
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/job-consumer ./cmd/job-consumer

FROM alpine:3.20

RUN apk add --no-cache git openssh-client bash make gcc musl-dev go curl

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin v2.6.1

RUN adduser -D -H -u 10001 appuser
RUN mkdir -p /home/appuser/.ssh /go && chown -R appuser:appuser /home/appuser /go

ENV GOPATH=/go GOPRIVATE=github.com/astronautsid

USER appuser

RUN git config --global url."git@github-astronautsid:".insteadOf "https://github.com/"

COPY --from=build /out/api /api
COPY --from=build /out/job-orchestrator /job-orchestrator
COPY --from=build /out/job-consumer /job-consumer

EXPOSE 8080

CMD ["/api"]
