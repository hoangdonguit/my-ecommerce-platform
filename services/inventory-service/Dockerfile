FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o inventory-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o inventory-consumer ./cmd/consumer

FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /app/inventory-api /app/inventory-api
COPY --from=builder /app/inventory-consumer /app/inventory-consumer

EXPOSE 8082