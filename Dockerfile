FROM golang:1.20-buster as builder

ARG TARGETPLATFORM=linux/amd64
ARG VERSION=0.0

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . ./

RUN go build -v -o banhammer ./cmd/banhammer/main.go -ldflags "-X main.version=${VERSION}"

FROM debian:buster-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/banhammer /app/banhammer

CMD ["/app/banhammer"]
