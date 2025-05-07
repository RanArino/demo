variable "queue_name" {
  description = "Name of the SQS queue for document processing"
  type        = string
}

variable "dlq_name" {
  description = "Name of the SQS dead letter queue"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
} 