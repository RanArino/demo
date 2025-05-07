output "s3_bucket_name" {
  description = "Name of the S3 bucket for document storage"
  value       = module.storage.bucket_name
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket for document storage"
  value       = module.storage.bucket_arn
}

output "sqs_queue_url" {
  description = "URL of the SQS queue for document processing"
  value       = module.queue.queue_url
}

output "sqs_dlq_url" {
  description = "URL of the SQS dead letter queue"
  value       = module.queue.dlq_url
}

output "database_url" {
  description = "Connection URL for the PostgreSQL database"
  value       = module.database.connection_string
  sensitive   = true
}

output "app_role_arn" {
  description = "ARN of the IAM role for the application"
  value       = module.iam.app_role_arn
}

output "app_role_name" {
  description = "Name of the IAM role for the application"
  value       = module.iam.app_role_name
} 