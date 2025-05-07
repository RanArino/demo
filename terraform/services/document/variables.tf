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
  description = "VPC ID for the database (if not using shared infrastructure)"
  type        = string
  default     = ""
}

variable "subnet_ids" {
  description = "Subnet IDs for the database (if not using shared infrastructure)"
  type        = list(string)
  default     = []
}

variable "service_principal" {
  description = "The AWS service that will assume the IAM role (e.g., ec2.amazonaws.com, ecs-tasks.amazonaws.com)"
  type        = string
  default     = "ec2.amazonaws.com"
}

# Database variables
variable "db_username" {
  description = "Username for the database"
  type        = string
  default     = "postgres"
  sensitive   = true
}

variable "db_password" {
  description = "Password for the database"
  type        = string
  sensitive   = true
}

variable "create_db_instance" {
  description = "Whether to create a new RDS instance"
  type        = bool
  default     = true
}

variable "db_connection_string" {
  description = "Database connection string (used if create_db_instance is false)"
  type        = string
  default     = ""
  sensitive   = true
} 