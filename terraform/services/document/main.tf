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
  #   key            = "services/document/terraform.tfstate"
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
      Service     = "document"
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

# Use modules for different infrastructure components
module "storage" {
  source      = "../../modules/storage"
  bucket_name = "${var.project_name}-${var.environment}-document-service-documents"
  environment = var.environment
}

module "queue" {
  source      = "../../modules/queue"
  queue_name  = "${var.project_name}-${var.environment}-document-service-queue"
  dlq_name    = "${var.project_name}-${var.environment}-document-service-dlq"
  environment = var.environment
}

module "database" {
  source               = "../../modules/database"
  db_identifier        = "${var.project_name}-${var.environment}-document-service"
  db_name              = "document_service"
  db_username          = var.db_username
  db_password          = var.db_password
  environment          = var.environment
  vpc_id               = local.vpc_id
  subnet_ids           = local.subnet_ids
  create_db_instance   = var.create_db_instance
  db_connection_string = var.db_connection_string
}

# IAM module for application permissions
module "iam" {
  source            = "../../modules/iam"
  project_name      = var.project_name
  environment       = var.environment
  service_principal = var.service_principal
  s3_bucket_arn     = module.storage.bucket_arn
  sqs_queue_arn     = module.queue.queue_arn
  sqs_dlq_arn       = module.queue.dlq_arn
} 