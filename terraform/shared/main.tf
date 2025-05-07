terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  required_version = ">= 1.0.0"
  
  # Production should use a remote backend
  # backend "s3" {
  #   bucket         = "terraform-state-{company}"
  #   key            = "shared/terraform.tfstate"
  #   region         = "us-east-1"
  #   dynamodb_table = "terraform-locks"
  #   encrypt        = true
  # }
}

provider "aws" {
  region = var.aws_region
  
  # For local development with tools like LocalStack
  dynamic "endpoints" {
    for_each = var.use_localstack ? [1] : []
    content {
      s3  = var.localstack_endpoint
      sqs = var.localstack_endpoint
    }
  }
  
  # For local development with LocalStack
  skip_credentials_validation = var.use_localstack
  skip_metadata_api_check     = var.use_localstack
  skip_requesting_account_id  = var.use_localstack
  
  default_tags {
    tags = {
      Environment = var.environment
      Project     = var.project_name
      ManagedBy   = "terraform"
    }
  }
}

# VPC for shared infrastructure
resource "aws_vpc" "main" {
  count = var.create_vpc ? 1 : 0
  
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true
  
  tags = {
    Name = "${var.project_name}-${var.environment}-vpc"
  }
}

# Create public subnets
resource "aws_subnet" "public" {
  count = var.create_vpc ? length(var.public_subnet_cidrs) : 0
  
  vpc_id                  = aws_vpc.main[0].id
  cidr_block              = var.public_subnet_cidrs[count.index]
  availability_zone       = var.availability_zones[count.index % length(var.availability_zones)]
  map_public_ip_on_launch = true
  
  tags = {
    Name = "${var.project_name}-${var.environment}-public-${count.index + 1}"
  }
}

# Create private subnets
resource "aws_subnet" "private" {
  count = var.create_vpc ? length(var.private_subnet_cidrs) : 0
  
  vpc_id            = aws_vpc.main[0].id
  cidr_block        = var.private_subnet_cidrs[count.index]
  availability_zone = var.availability_zones[count.index % length(var.availability_zones)]
  
  tags = {
    Name = "${var.project_name}-${var.environment}-private-${count.index + 1}"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "igw" {
  count = var.create_vpc ? 1 : 0
  
  vpc_id = aws_vpc.main[0].id
  
  tags = {
    Name = "${var.project_name}-${var.environment}-igw"
  }
}

# Route table for public subnets
resource "aws_route_table" "public" {
  count = var.create_vpc ? 1 : 0
  
  vpc_id = aws_vpc.main[0].id
  
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw[0].id
  }
  
  tags = {
    Name = "${var.project_name}-${var.environment}-public-rt"
  }
}

# Associate public subnets with the public route table
resource "aws_route_table_association" "public" {
  count = var.create_vpc ? length(var.public_subnet_cidrs) : 0
  
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public[0].id
}

# Route table for private subnets
resource "aws_route_table" "private" {
  count = var.create_vpc ? 1 : 0
  
  vpc_id = aws_vpc.main[0].id
  
  tags = {
    Name = "${var.project_name}-${var.environment}-private-rt"
  }
}

# Associate private subnets with the private route table
resource "aws_route_table_association" "private" {
  count = var.create_vpc ? length(var.private_subnet_cidrs) : 0
  
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[0].id
} 