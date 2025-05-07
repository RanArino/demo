resource "aws_sqs_queue" "dead_letter_queue" {
  name                      = var.dlq_name
  message_retention_seconds = 1209600  # 14 days
  
  # Server-side encryption
  sqs_managed_sse_enabled = true
  
  tags = {
    Name        = var.dlq_name
    Environment = var.environment
  }
}

resource "aws_sqs_queue" "document_queue" {
  name                       = var.queue_name
  visibility_timeout_seconds = 300  # 5 minutes
  message_retention_seconds  = 345600  # 4 days
  max_message_size           = 262144  # 256 KiB
  delay_seconds              = 0
  receive_wait_time_seconds  = 20  # Long polling
  
  # Server-side encryption
  sqs_managed_sse_enabled = true
  
  # Configure dead letter queue
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dead_letter_queue.arn
    maxReceiveCount     = 5
  })
  
  tags = {
    Name        = var.queue_name
    Environment = var.environment
  }
} 