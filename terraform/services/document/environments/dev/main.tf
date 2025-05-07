module "document_service" {
  source = "../../"
  
  # Environment-specific configuration
  environment         = "dev"
  project_name        = "demo"
  aws_region          = "us-east-1"
  
  # Local development with LocalStack
  use_localstack      = true
  localstack_endpoint = "http://localhost:4566"
  
  # Use shared infrastructure if it exists
  use_shared_infrastructure = false
  
  # Database configuration (for local development)
  db_username          = "postgres"
  db_password          = "postgres"
  create_db_instance   = false
  db_connection_string = "postgres://postgres:postgres@localhost:5432/document_service?sslmode=disable"
  
  # IAM configuration
  service_principal    = "ec2.amazonaws.com"
} 