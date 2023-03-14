FROM golang:1.20-buster as builder

ARG VERSION
WORKDIR /app
COPY . .

RUN go build -v -o banhammer -ldflags "-s -v -w -X 'main.version=${VERSION}'" ./cmd/banhammer/main.go

FROM debian:buster-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/banhammer /app/banhammer

CMD ["/app/banhammer"]
