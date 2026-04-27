FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o notification-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o notification-consumer ./cmd/consumer

FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /app/notification-api /app/notification-api
COPY --from=builder /app/notification-consumer /app/notification-consumer

EXPOSE 8084