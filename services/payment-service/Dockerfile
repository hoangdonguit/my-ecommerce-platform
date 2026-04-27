FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o payment-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o payment-consumer ./cmd/consumer

FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /app/payment-api /app/payment-api
COPY --from=builder /app/payment-consumer /app/payment-consumer

EXPOSE 8083