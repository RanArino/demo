# Build stage
FROM golang:1.24 AS builder
# Install build dependencies for CGO
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential ca-certificates \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
# Copy only go.mod and go.sum first for efficient caching
COPY ms_knowledge/go.mod ms_knowledge/go.sum ./
# Also copy the sibling ms_user module so the replace directive can be resolved during download
COPY ms_user /ms_user
RUN go mod download
# Copy the rest of the source
COPY ms_knowledge/ .
# Enable CGO; rely on confluent-kafka-go vendored glibc static library
ENV CGO_ENABLED=1 GOOS=linux
RUN go build -o /app/server ./cmd/main.go

# Final stage
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/server .
COPY ms_knowledge/.env.local .
EXPOSE 8080
CMD ["./server"]
