resource "aws_s3_bucket" "document_bucket" {
  bucket = var.bucket_name

  # Enable versioning if needed
  # versioning {
  #   enabled = true
  # }

  # Configure lifecycle rules if needed
  lifecycle_rule {
    id      = "expire-old-versions"
    enabled = true

    expiration {
      days = 90
    }

    noncurrent_version_expiration {
      days = 30
    }

    prefix = "documents/"
  }

  tags = {
    Name        = var.bucket_name
    Environment = var.environment
  }
}

# Configure bucket for server-side encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "encryption" {
  bucket = aws_s3_bucket.document_bucket.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Configure bucket to block public access
resource "aws_s3_bucket_public_access_block" "block_public_access" {
  bucket = aws_s3_bucket.document_bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Configure CORS if needed (for direct uploads from browser)
resource "aws_s3_bucket_cors_configuration" "cors" {
  bucket = aws_s3_bucket.document_bucket.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST", "DELETE"]
    allowed_origins = ["*"] # Restrict to your domains in production
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
} 