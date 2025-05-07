#!/bin/bash

# Script to create a new service infrastructure
# Usage: ./create-service.sh <service_name>
# Example: ./create-service.sh auth

set -e

if [ $# -lt 1 ]; then
  echo "Usage: $0 <service_name>"
  echo "Example: $0 auth"
  exit 1
fi

SERVICE_NAME=$1
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
SERVICE_DIR="$ROOT_DIR/services/$SERVICE_NAME"

# Check if service already exists
if [ -d "$SERVICE_DIR" ]; then
  echo "Error: Service '$SERVICE_NAME' already exists."
  exit 1
fi

# Create service directory structure
echo "Creating directory structure for $SERVICE_NAME service..."
mkdir -p "$SERVICE_DIR/environments/dev" "$SERVICE_DIR/environments/prod"

# Create main.tf for the service
cat > "$SERVICE_DIR/main.tf" << EOF
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  required_version = ">= 1.0.0"
  
  # For production, use a remote backend
  # backend "s3" {
  #   bucket         = "terraform-state-{company}"
  #   key            = "services/$SERVICE_NAME/terraform.tfstate"
  #   region         = "us-east-1"
  #   dynamodb_table = "terraform-locks"
  #   encrypt        = true
  # }
}

provider "aws" {
  region = var.aws_region
  
  # For local development with LocalStack
  dynamic "endpoints" {
    for_each = var.use_localstack ? [1] : []
    content {
      s3  = var.localstack_endpoint
      sqs = var.localstack_endpoint
    }
  }
  
  # For local development with LocalStack
  skip_credentials_validation = var.use_localstack
  skip_metadata_api_check     = var.use_localstack
  skip_requesting_account_id  = var.use_localstack
  
  default_tags {
    tags = {
      Environment = var.environment
      Project     = var.project_name
      Service     = "$SERVICE_NAME"
      ManagedBy   = "terraform"
    }
  }
}

# Get data from shared infrastructure if needed
data "terraform_remote_state" "shared" {
  count = var.use_shared_infrastructure ? 1 : 0
  
  backend = var.shared_state_backend
  config  = var.shared_state_config
}

locals {
  # Use VPC values from shared infrastructure or directly provided values
  vpc_id = var.use_shared_infrastructure ? data.terraform_remote_state.shared[0].outputs.vpc_id : var.vpc_id
  subnet_ids = var.use_shared_infrastructure ? data.terraform_remote_state.shared[0].outputs.private_subnet_ids : var.subnet_ids
}

# Add your service-specific resources here
# Example:
# module "storage" {
#   source      = "../../modules/storage"
#   bucket_name = "\${var.project_name}-\${var.environment}-$SERVICE_NAME-data"
#   environment = var.environment
# }

# IAM module for application permissions
module "iam" {
  source            = "../../modules/iam"
  project_name      = var.project_name
  environment       = var.environment
  service_principal = var.service_principal
  s3_bucket_arn     = "arn:aws:s3:::dummy-bucket"  # Replace with actual S3 bucket ARN
  sqs_queue_arn     = "arn:aws:sqs:region:account:dummy-queue"  # Replace with actual SQS queue ARN
  sqs_dlq_arn       = "arn:aws:sqs:region:account:dummy-dlq"  # Replace with actual SQS DLQ ARN
}
EOF

# Create variables.tf for the service
cat > "$SERVICE_DIR/variables.tf" << EOF
variable "aws_region" {
  description = "The AWS region to deploy resources to"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "demo"
}

variable "use_localstack" {
  description = "Whether to use LocalStack for local development"
  type        = bool
  default     = false
}

variable "localstack_endpoint" {
  description = "The endpoint URL for LocalStack"
  type        = string
  default     = "http://localhost:4566"
}

variable "use_shared_infrastructure" {
  description = "Whether to use shared infrastructure from remote state"
  type        = bool
  default     = false
}

variable "shared_state_backend" {
  description = "Backend type for shared infrastructure state"
  type        = string
  default     = "local"
}

variable "shared_state_config" {
  description = "Configuration for shared infrastructure state backend"
  type        = map(string)
  default     = {
    path = "../../../shared/terraform.tfstate"
  }
}

variable "vpc_id" {
  description = "VPC ID (if not using shared infrastructure)"
  type        = string
  default     = ""
}

variable "subnet_ids" {
  description = "Subnet IDs (if not using shared infrastructure)"
  type        = list(string)
  default     = []
}

variable "service_principal" {
  description = "The AWS service that will assume the IAM role (e.g., ec2.amazonaws.com, ecs-tasks.amazonaws.com)"
  type        = string
  default     = "ec2.amazonaws.com"
}

# Add your service-specific variables here
EOF

# Create outputs.tf for the service
cat > "$SERVICE_DIR/outputs.tf" << EOF
output "app_role_arn" {
  description = "ARN of the IAM role for the application"
  value       = module.iam.app_role_arn
}

output "app_role_name" {
  description = "Name of the IAM role for the application"
  value       = module.iam.app_role_name
}

# Add your service-specific outputs here
EOF

# Create dev environment files
cat > "$SERVICE_DIR/environments/dev/main.tf" << EOF
module "${SERVICE_NAME}_service" {
  source = "../../"
  
  # Environment-specific configuration
  environment         = "dev"
  project_name        = "demo"
  aws_region          = "us-east-1"
  
  # Local development with LocalStack
  use_localstack      = true
  localstack_endpoint = "http://localhost:4566"
  
  # Use shared infrastructure if it exists
  use_shared_infrastructure = false
  
  # IAM configuration
  service_principal    = "ec2.amazonaws.com"
  
  # Add your service-specific variables here
}
EOF

cat > "$SERVICE_DIR/environments/dev/terraform.tfvars.example" << EOF
# Environment-specific configuration
environment = "dev"
project_name = "demo"
aws_region = "us-east-1"

# Local development with LocalStack
use_localstack = true
localstack_endpoint = "http://localhost:4566"

# Use shared infrastructure if it exists
use_shared_infrastructure = false

# IAM configuration
service_principal = "ec2.amazonaws.com"

# Add your service-specific variables here
EOF

# Create prod environment files
cat > "$SERVICE_DIR/environments/prod/main.tf" << EOF
module "${SERVICE_NAME}_service" {
  source = "../../"
  
  # Environment-specific configuration
  environment   = "prod"
  project_name  = "demo"
  aws_region    = "us-east-1"
  
  # Production uses real AWS resources
  use_localstack = false
  
  # Use shared infrastructure
  use_shared_infrastructure = true
  shared_state_backend     = "s3"
  shared_state_config      = {
    bucket         = "terraform-state-{company}"
    key            = "shared/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
  }
  
  # IAM configuration (ECS for containerized applications)
  service_principal = "ecs-tasks.amazonaws.com"
  
  # Add your service-specific variables here
}
EOF

cat > "$SERVICE_DIR/environments/prod/variables.tf" << EOF
# Add any environment-specific variables here
# Example: Database credentials (don't store actual values in version control)
EOF

cat > "$SERVICE_DIR/environments/prod/terraform.tfvars.example" << EOF
# Environment-specific configuration
environment = "prod"
project_name = "demo"
aws_region = "us-east-1"

# Production uses real AWS resources
use_localstack = false

# Use shared infrastructure
use_shared_infrastructure = true
shared_state_backend = "s3"
shared_state_config = {
  bucket         = "terraform-state-{company}"
  key            = "shared/terraform.tfstate"
  region         = "us-east-1"
  encrypt        = true
}

# IAM configuration (ECS for containerized applications)
service_principal = "ecs-tasks.amazonaws.com"

# Add your service-specific variables here
EOF

echo "New service '$SERVICE_NAME' created successfully at: $SERVICE_DIR"
echo "Next steps:"
echo "1. Customize the Terraform files to add your service-specific resources"
echo "2. Deploy the infrastructure with: ./deploy-service.sh $SERVICE_NAME dev" 