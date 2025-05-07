#!/bin/bash

# Script to deploy shared infrastructure
# Usage: ./deploy-shared.sh <environment>
# Example: ./deploy-shared.sh dev

set -e

if [ $# -lt 1 ]; then
  echo "Usage: $0 <environment>"
  echo "Example: $0 dev"
  exit 1
fi

ENVIRONMENT=$1
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
SHARED_ENV_DIR="$ROOT_DIR/shared/environments/$ENVIRONMENT"

# Check if environment exists
if [ ! -d "$SHARED_ENV_DIR" ]; then
  echo "Error: Environment '$ENVIRONMENT' not found. Available environments:"
  ls -1 "$ROOT_DIR/shared/environments"
  exit 1
fi

# Create terraform.tfvars if it doesn't exist
if [ ! -f "$SHARED_ENV_DIR/terraform.tfvars" ]; then
  if [ -f "$SHARED_ENV_DIR/terraform.tfvars.example" ]; then
    echo "terraform.tfvars not found. Creating from example..."
    cp "$SHARED_ENV_DIR/terraform.tfvars.example" "$SHARED_ENV_DIR/terraform.tfvars"
    echo "Please review and modify $SHARED_ENV_DIR/terraform.tfvars as needed."
  else
    echo "Warning: Neither terraform.tfvars nor terraform.tfvars.example found."
    echo "You may need to create terraform.tfvars manually."
  fi
fi

# Deploy shared infrastructure
echo "Deploying shared infrastructure for environment $ENVIRONMENT..."
cd "$SHARED_ENV_DIR"
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
  echo "Shared infrastructure deployed successfully. Outputting environment variables:"
  
  export_env_file="$SHARED_ENV_DIR/terraform_outputs.env"
  echo "# Terraform outputs for shared infrastructure in $ENVIRONMENT environment" > "$export_env_file"
  echo "# Generated on $(date)" >> "$export_env_file"
  
  # Add shared infrastructure exports
  echo "export VPC_ID=$(terraform output -raw vpc_id)" | tee -a "$export_env_file"
  echo "export PUBLIC_SUBNET_IDS=$(terraform output -json public_subnet_ids | tr -d '\n')" | tee -a "$export_env_file"
  echo "export PRIVATE_SUBNET_IDS=$(terraform output -json private_subnet_ids | tr -d '\n')" | tee -a "$export_env_file"
  echo "export AWS_REGION=$(terraform output -raw aws_region)" | tee -a "$export_env_file"
  
  echo ""
  echo "Environment variables saved to: $export_env_file"
  echo "You can load them with: source $export_env_file"
  echo ""
  echo "Services that depend on shared infrastructure can now be deployed."
else
  echo "Deployment cancelled."
  # Clean up the plan file
  rm -f tfplan
fi 