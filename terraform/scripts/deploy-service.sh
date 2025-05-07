#!/bin/bash

# Script to deploy a specific service infrastructure
# Usage: ./deploy-service.sh <service_name> <environment>
# Example: ./deploy-service.sh document dev

set -e

if [ $# -lt 2 ]; then
  echo "Usage: $0 <service_name> <environment>"
  echo "Example: $0 document dev"
  exit 1
fi

SERVICE_NAME=$1
ENVIRONMENT=$2
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
SERVICE_ENV_DIR="$ROOT_DIR/services/$SERVICE_NAME/environments/$ENVIRONMENT"

# Check if service exists
if [ ! -d "$ROOT_DIR/services/$SERVICE_NAME" ]; then
  echo "Error: Service '$SERVICE_NAME' not found. Available services:"
  ls -1 "$ROOT_DIR/services"
  exit 1
fi

# Check if environment exists
if [ ! -d "$SERVICE_ENV_DIR" ]; then
  echo "Error: Environment '$ENVIRONMENT' not found for service '$SERVICE_NAME'. Available environments:"
  ls -1 "$ROOT_DIR/services/$SERVICE_NAME/environments"
  exit 1
fi

# Check if shared infrastructure is needed
USE_SHARED=$(grep "use_shared_infrastructure" "$SERVICE_ENV_DIR/terraform.tfvars" 2>/dev/null | grep -q "true" && echo "true" || echo "false")

if [ "$USE_SHARED" = "true" ]; then
  echo "This service uses shared infrastructure."
  SHARED_ENV_DIR="$ROOT_DIR/shared/environments/$ENVIRONMENT"
  
  if [ ! -d "$SHARED_ENV_DIR" ]; then
    echo "Error: Shared environment '$ENVIRONMENT' not found. Available shared environments:"
    ls -1 "$ROOT_DIR/shared/environments"
    exit 1
  fi
  
  echo "Checking if shared infrastructure is deployed..."
  if [ ! -f "$SHARED_ENV_DIR/terraform.tfstate" ] && [ ! -f "$SHARED_ENV_DIR/.terraform/terraform.tfstate" ]; then
    echo "Shared infrastructure for environment '$ENVIRONMENT' may not be deployed yet."
    read -p "Would you like to deploy shared infrastructure first? (y/n) " DEPLOY_SHARED
    if [[ "$DEPLOY_SHARED" =~ ^[Yy]$ ]]; then
      echo "Deploying shared infrastructure..."
      cd "$SHARED_ENV_DIR"
      terraform init
      terraform apply
    fi
  fi
fi

# Create terraform.tfvars if it doesn't exist
if [ ! -f "$SERVICE_ENV_DIR/terraform.tfvars" ]; then
  if [ -f "$SERVICE_ENV_DIR/terraform.tfvars.example" ]; then
    echo "terraform.tfvars not found. Creating from example..."
    cp "$SERVICE_ENV_DIR/terraform.tfvars.example" "$SERVICE_ENV_DIR/terraform.tfvars"
    echo "Please review and modify $SERVICE_ENV_DIR/terraform.tfvars as needed."
  else
    echo "Warning: Neither terraform.tfvars nor terraform.tfvars.example found."
    echo "You may need to create terraform.tfvars manually."
  fi
fi

# Deploy service infrastructure
echo "Deploying $SERVICE_NAME service for environment $ENVIRONMENT..."
cd "$SERVICE_ENV_DIR"
terraform init

echo "Planning infrastructure changes..."
terraform plan -out=tfplan

echo ""
echo "Would you like to apply these changes? (y/n)"
read -r apply_changes

if [[ "$apply_changes" =~ ^[Yy]$ ]]; then
  echo "Applying infrastructure changes..."
  terraform apply tfplan
  
  # Export outputs as environment variables
  echo ""
  echo "Infrastructure deployed successfully. Outputting environment variables:"
  
  export_env_file="$SERVICE_ENV_DIR/terraform_outputs.env"
  echo "# Terraform outputs for $SERVICE_NAME service in $ENVIRONMENT environment" > "$export_env_file"
  echo "# Generated on $(date)" >> "$export_env_file"
  
  # Add service-specific exports
  # For document service
  if [ "$SERVICE_NAME" = "document" ]; then
    echo "export S3_BUCKET=$(terraform output -raw s3_bucket_name)" | tee -a "$export_env_file"
    echo "export SQS_QUEUE_URL=$(terraform output -raw sqs_queue_url)" | tee -a "$export_env_file"
    echo "export SQS_DEAD_LETTER_QUEUE_URL=$(terraform output -raw sqs_dlq_url)" | tee -a "$export_env_file"
    # Database URL is sensitive, don't print it to console
    echo "export DATABASE_URL=$(terraform output -raw database_url)" >> "$export_env_file"
    echo "export AWS_ROLE_ARN=$(terraform output -raw app_role_arn)" | tee -a "$export_env_file"
  # Add other services as needed
  # elif [ "$SERVICE_NAME" = "auth" ]; then
  #   echo "export AUTH_SPECIFIC_VAR=$(terraform output -raw auth_specific_var)" | tee -a "$export_env_file"
  fi
  
  echo ""
  echo "Environment variables saved to: $export_env_file"
  echo "You can load them with: source $export_env_file"
else
  echo "Deployment cancelled."
  # Clean up the plan file
  rm -f tfplan
fi 