FROM golang:1.19-alpine3.17 AS builder

WORKDIR /build

COPY go.mod go.sum domain-exporter.go ./
RUN go mod download && go mod verify

RUN CGO_ENABLED=0 GOOS=linux go build -o ./domain-exporter ./domain-exporter.go

FROM ubuntu:latest

RUN apt-get update && \
    apt install -y netbase && \
    apt-get install -y whois && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/domain-exporter ./

CMD ["./domain-exporter"]