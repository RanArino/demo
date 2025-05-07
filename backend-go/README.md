# Document Service - Go Backend

This service handles document management features including:

- Document upload and storage
- Document parsing and processing 
- Secure document retrieval with role-based access control
- Content extraction and transformation

## Architecture

The service follows a hexagonal architecture pattern:

- **Core domain**: Located in `internal/core`
  - Models define the business entities
  - Ports define the interfaces for external adapters
  - Services contain the business logic

- **Infrastructure adapters**: Located in `internal/infra`
  - Storage implementations (S3, PostgreSQL)
  - Queue implementations (SQS)

- **API layer**: Located in `internal/api`
  - HTTP handlers
  - API routes and middleware

## Key Features

- **RBAC**: Role-based access control with permission checking across all operations
- **Modular parsing**: Support for multiple file formats through extensible parsing strategies
- **Transactional operations**: Database transactions ensure data consistency
- **Scalable storage**: S3 storage with pre-signed URLs for direct client uploads/downloads
- **Asynchronous processing**: SQS queues for background document processing

## Setup

### Prerequisites

- Go 1.21+
- PostgreSQL
- AWS S3 (or compatible like MinIO)
- AWS SQS (or compatible)

### Environment Variables

The service is configured using environment variables:

```
# Server configuration
SERVER_PORT=8080
SERVER_HOST=
SERVER_READ_TIMEOUT=10
SERVER_WRITE_TIMEOUT=10
SERVER_SHUTDOWN_TIMEOUT=5

# Database configuration
DATABASE_URL=postgres://postgres:postgres@localhost:5432/document_service?sslmode=disable
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=25
DATABASE_CONN_MAX_LIFETIME_SECONDS=300

# S3 configuration
S3_REGION=us-east-1
S3_BUCKET=documents
S3_ENDPOINT=http://localhost:9000  # For MinIO/localstack
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_USE_IAM_ROLE=false
S3_FORCE_PATH_STYLE=true
S3_DISABLE_SSL=false
S3_PRESIGNED_URL_DURATION=900

# SQS configuration
SQS_QUEUE_URL=http://localhost:4566/000000000000/document-queue
SQS_DEAD_LETTER_QUEUE_URL=http://localhost:4566/000000000000/document-dlq
```

### Running locally

1. Install dependencies:
   ```
   go mod download
   ```

2. Run database migrations:
   ```
   go run cmd/migrate/main.go up
   ```

3. Start the server:
   ```
   go run cmd/server/main.go
   ```

### Docker

Build the Docker image:
```
docker build -t document-service:latest .
```

Run with Docker Compose:
```
docker-compose up -d
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| POST | /api/v1/documents | Upload a document |
| GET | /api/v1/documents | List documents |
| GET | /api/v1/documents/:id | Get document details |
| GET | /api/v1/documents/:id/content | Get document content |
| GET | /api/v1/documents/:id/download | Download document |
| DELETE | /api/v1/documents/:id | Delete document |
| POST | /api/v1/documents/upload-url | Get pre-signed upload URL |

## Development

### Adding a new parsing strategy

1. Create a new strategy in `internal/core/services/document/parser/strategies`
2. Implement the `DocumentParsingStrategy` interface
3. Register the strategy in `cmd/server/main.go`

### Running tests

```
go test ./...
```

## Infrastructure Management with Terraform

The project now supports infrastructure as code using Terraform. The Terraform configurations are located in the `terraform/` directory and define the following resources:

- S3 bucket for document storage
- SQS queues for document processing
- PostgreSQL database (RDS) for metadata storage
- IAM roles and policies for AWS service access

### Infrastructure Structure

- `terraform/`: Main Terraform configuration
  - `modules/`: Reusable Terraform modules
    - `storage/`: S3 configuration
    - `queue/`: SQS configuration
    - `database/`: PostgreSQL configuration
    - `iam/`: IAM roles and policies
  - `environments/`: Environment-specific configurations
    - `dev/`: Development environment (using LocalStack)
    - `prod/`: Production environment (using AWS)
  - `scripts/`: Helper scripts for Terraform operations

### Getting Started with Terraform

1. **Initialize Terraform for the dev environment:**

   ```bash
   cd terraform
   ./scripts/terraform-init.sh dev
   ```

2. **Apply the Terraform configuration:**

   ```bash
   ./scripts/terraform-apply.sh dev
   ```

3. **Load environment variables:**

   After applying the Terraform configuration, environment variables will be saved to `terraform/environments/dev/terraform_outputs.env`. You can load them with:

   ```bash
   source terraform/environments/dev/terraform_outputs.env
   ```

   Or in your development workflow:

   ```bash
   cp scripts/env-example.sh scripts/.env.sh
   # Edit scripts/.env.sh as needed
   source scripts/.env.sh
   ```

### Local Development with LocalStack

For local development, the Terraform configurations can use LocalStack to emulate AWS services. The `dev` environment is configured to use LocalStack by default.

Start LocalStack with:

```bash
docker run -d --name localstack -p 4566:4566 -p 4571:4571 localstack/localstack
```

Then apply the Terraform configuration as described above.

### Production Deployment

For production deployment, create a `terraform.tfvars` file in the `terraform/environments/prod` directory:

```bash
cd terraform/environments/prod
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your production values
```

Then initialize and apply the Terraform configuration:

```bash
cd ../../
./scripts/terraform-init.sh prod
./scripts/terraform-apply.sh prod
``` 