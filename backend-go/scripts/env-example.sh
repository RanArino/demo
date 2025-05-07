#!/bin/bash

# Example environment variables script
# Copy this to .env.sh and modify as needed

# Server configuration
export SERVER_PORT=8080
export SERVER_HOST=
export SERVER_READ_TIMEOUT=10
export SERVER_WRITE_TIMEOUT=10
export SERVER_SHUTDOWN_TIMEOUT=5

# AWS configuration
export AWS_REGION=us-east-1
export AWS_USE_IAM_ROLE=false

# For local development with explicit credentials
export AWS_ACCESS_KEY_ID=
export AWS_SECRET_ACCESS_KEY=

# For using IAM roles (populated by Terraform)
export AWS_ROLE_ARN=

# S3 configuration (populated by Terraform)
export S3_BUCKET=
export S3_PRESIGNED_URL_DURATION=900

# For LocalStack/MinIO
export S3_ENDPOINT=http://localhost:4566
export S3_FORCE_PATH_STYLE=true
export S3_DISABLE_SSL=true

# SQS configuration (populated by Terraform)
export SQS_QUEUE_URL=
export SQS_DEAD_LETTER_QUEUE_URL=
export SQS_REGION=us-east-1

# Database configuration (populated by Terraform)
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/document_service?sslmode=disable
export DATABASE_MAX_OPEN_CONNS=25
export DATABASE_MAX_IDLE_CONNS=25
export DATABASE_CONN_MAX_LIFETIME_SECONDS=300

# Uncomment to source Terraform outputs (replace with your environment)
# source ../terraform/environments/dev/terraform_outputs.env

echo "Environment variables set successfully" 