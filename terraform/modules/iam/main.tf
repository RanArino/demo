# IAM role for the application (EC2, ECS, or Lambda)
resource "aws_iam_role" "app_role" {
  name = "${var.project_name}-${var.environment}-app-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = var.service_principal
        }
      },
    ]
  })

  tags = {
    Name        = "${var.project_name}-${var.environment}-app-role"
    Environment = var.environment
  }
}

# Policy for S3 access
resource "aws_iam_policy" "s3_access" {
  name        = "${var.project_name}-${var.environment}-s3-access"
  description = "Policy for S3 bucket access"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket",
        ]
        Effect = "Allow"
        Resource = [
          "${var.s3_bucket_arn}",
          "${var.s3_bucket_arn}/*"
        ]
      }
    ]
  })
}

# Policy for SQS access
resource "aws_iam_policy" "sqs_access" {
  name        = "${var.project_name}-${var.environment}-sqs-access"
  description = "Policy for SQS queue access"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "sqs:SendMessage",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:GetQueueUrl",
          "sqs:ChangeMessageVisibility"
        ]
        Effect = "Allow"
        Resource = [
          "${var.sqs_queue_arn}",
          "${var.sqs_dlq_arn}"
        ]
      }
    ]
  })
}

# Attach policies to the role
resource "aws_iam_role_policy_attachment" "s3_attachment" {
  role       = aws_iam_role.app_role.name
  policy_arn = aws_iam_policy.s3_access.arn
}

resource "aws_iam_role_policy_attachment" "sqs_attachment" {
  role       = aws_iam_role.app_role.name
  policy_arn = aws_iam_policy.sqs_access.arn
} 