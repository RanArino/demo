# Builder stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/main.go
# Final stage
FROM alpine:latest
WORKDIR /
COPY --from=builder /main /main
EXPOSE 50051 8081
ENTRYPOINT ["/main"]
