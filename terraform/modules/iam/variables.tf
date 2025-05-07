variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "service_principal" {
  description = "The AWS service that will assume this role (e.g., ec2.amazonaws.com, ecs-tasks.amazonaws.com)"
  type        = string
  default     = "ec2.amazonaws.com"
}

variable "s3_bucket_arn" {
  description = "ARN of the S3 bucket to grant access to"
  type        = string
}

variable "sqs_queue_arn" {
  description = "ARN of the SQS queue to grant access to"
  type        = string
}

variable "sqs_dlq_arn" {
  description = "ARN of the SQS dead letter queue to grant access to"
  type        = string
} 