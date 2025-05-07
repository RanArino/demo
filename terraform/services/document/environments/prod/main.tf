module "document_service" {
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
  
  # Database configuration (use secure values from variables)
  db_username    = var.db_username
  db_password    = var.db_password
  
  # IAM configuration (ECS for containerized applications)
  service_principal = "ecs-tasks.amazonaws.com"
} 