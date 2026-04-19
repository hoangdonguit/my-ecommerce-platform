FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o order-service ./cmd/api

FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /app/order-service /app/order-service

EXPOSE 8081

CMD ["/app/order-service"]