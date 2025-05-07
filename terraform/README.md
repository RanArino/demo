# Multi-Service Terraform Infrastructure

This directory contains Terraform configurations for managing infrastructure across multiple services in a modular and scalable way.

## Structure

```
terraform/
├── modules/              # Reusable modules for all services
│   ├── storage/          # S3 storage configuration
│   ├── queue/            # SQS queue configuration 
│   ├── database/         # Database configuration
│   └── iam/              # IAM roles and policies
├── services/             # Service-specific infrastructure
│   └── document/         # Document service infrastructure
│       ├── main.tf
│       ├── variables.tf
│       ├── outputs.tf
│       └── environments/
│           ├── dev/      # Development environment
│           └── prod/     # Production environment
└── shared/               # Shared infrastructure (VPC, networking)
    ├── main.tf
    ├── variables.tf
    ├── outputs.tf
    └── environments/
        ├── dev/
        └── prod/
```

## Key Concepts

### Shared Infrastructure

The `shared/` directory contains resources that are common across multiple services:

- VPC and subnets
- Network configurations (route tables, internet gateways)
- Security groups
- Cross-service IAM roles

### Service-Specific Infrastructure

Each service has its own directory under `services/` with:

- Service-specific resources (S3 buckets, SQS queues, databases)
- Service-specific IAM roles
- Environment-specific configurations

## Usage

### Setting Up Shared Infrastructure

1. **Initialize the shared infrastructure for development:**

   ```bash
   cd shared/environments/dev
   terraform init
   terraform apply
   ```

2. **Initialize the shared infrastructure for production:**

   ```bash
   cd shared/environments/prod
   terraform init
   terraform apply
   ```

### Deploying Service Infrastructure

1. **Deploy the document service infrastructure for development:**

   ```bash
   cd services/document/environments/dev
   terraform init
   terraform apply
   ```

2. **Deploy the document service infrastructure for production:**

   ```bash
   cd services/document/environments/prod
   terraform init
   terraform apply
   ```

## Adding a New Service

To add a new service:

1. Create a new directory under `services/` for your service
2. Create main.tf, variables.tf, and outputs.tf files for your service
3. Create environment-specific configurations under `environments/`
4. Reuse existing modules from the `modules/` directory
5. Reference shared infrastructure using the remote state data source

Example for a new authentication service:

```bash
mkdir -p services/auth/environments/{dev,prod}
touch services/auth/{main.tf,variables.tf,outputs.tf}
touch services/auth/environments/{dev,prod}/{main.tf,variables.tf,terraform.tfvars.example}
```

## State Management

Each service and the shared infrastructure have their own Terraform state:

- In development, state is typically stored locally
- In production, state should be stored remotely (e.g., in an S3 bucket with DynamoDB locking)

Services reference shared infrastructure using Terraform's remote state data source.

## Best Practices

1. **Module Reuse:** Use modules from the `modules/` directory across services
2. **Environment Consistency:** Keep environments consistent across services
3. **Service Independence:** Services should be independently deployable
4. **Naming Conventions:** Use consistent naming with service prefixes
5. **State Management:** Use remote state for production environments
6. **Documentation:** Document service-specific infrastructure requirements

## Additional Resources

- [Terraform Documentation](https://www.terraform.io/docs)
- [AWS Provider Documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) 