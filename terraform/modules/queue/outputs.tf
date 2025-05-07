output "queue_url" {
  description = "URL of the SQS queue"
  value       = aws_sqs_queue.document_queue.url
}

output "queue_arn" {
  description = "ARN of the SQS queue"
  value       = aws_sqs_queue.document_queue.arn
}

output "dlq_url" {
  description = "URL of the SQS dead letter queue"
  value       = aws_sqs_queue.dead_letter_queue.url
}

output "dlq_arn" {
  description = "ARN of the SQS dead letter queue"
  value       = aws_sqs_queue.dead_letter_queue.arn
} 