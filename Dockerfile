# syntax=docker/dockerfile:1.4

# Stage 1: build
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ingestion-service .

# Stage 2: runtime
FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/ingestion-service .

EXPOSE 8080

ENTRYPOINT ["./ingestion-service"]
