# Builder stage
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Install Python and build dependencies for Kafka client
RUN apt-get update && apt-get install -y python3 python3-venv python3-pip build-essential librdkafka-dev && rm -rf /var/lib/apt/lists/*

# Copy go.mod first to download dependencies for go_app
COPY go_app/go.mod go_app/go.sum ./go_app/
RUN cd go_app && go mod download

# Copy source code
COPY . .

# Install Poetry and build Python app if it exists
RUN pip3 install --break-system-packages poetry && \
    if [ -d python_app ]; then \
    cd python_app && \
    rm -rf .venv && \
    poetry config virtualenvs.in-project true && \
    poetry install && \
    poetry build && \
    poetry run pip install dist/*.whl; \
    fi

# Build Go binary with CGO enabled for Kafka client
RUN cd go_app && CGO_ENABLED=1 GOOS=linux go build -o /main ./cmd/main.go

# Final stage
FROM debian:bookworm-slim

WORKDIR /app

# Install Python runtime and Kafka dependencies
RUN apt-get update && apt-get install -y python3 python3-venv supervisor librdkafka1 && rm -rf /var/lib/apt/lists/*

# Copy built Go binary
COPY --from=builder /main /app/main

# Copy Python app and virtual environment
COPY --from=builder /app/python_app /app/python_app

# Copy supervisor configuration
COPY supervisord.conf /app/supervisord.conf
COPY scripts/start.sh /app/scripts/start.sh
RUN chmod +x /app/scripts/start.sh

EXPOSE 50055 8080

CMD ["sh", "/app/scripts/start.sh"]
