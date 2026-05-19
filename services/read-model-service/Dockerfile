FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o read-model-service ./cmd/api

FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /app/read-model-service /app/read-model-service

EXPOSE 8085

CMD ["/app/read-model-service"]
