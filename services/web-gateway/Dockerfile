FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o web-gateway ./cmd/api

FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /app/web-gateway /app/web-gateway

EXPOSE 8090

CMD ["/app/web-gateway"]