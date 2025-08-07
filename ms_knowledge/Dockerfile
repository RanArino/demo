# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/main.go

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY .env.local .
EXPOSE 8080
CMD ["./server"]
