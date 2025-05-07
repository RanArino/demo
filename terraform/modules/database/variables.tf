variable "db_identifier" {
  description = "Identifier for the database instance"
  type        = string
}

variable "db_name" {
  description = "Name of the database to create"
  type        = string
  default     = "document_service"
}

variable "db_username" {
  description = "Username for the database"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Password for the database"
  type        = string
  sensitive   = true
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "vpc_id" {
  description = "VPC ID for the database security group"
  type        = string
  default     = ""
}

variable "subnet_ids" {
  description = "Subnet IDs for the database subnet group"
  type        = list(string)
  default     = []
}

variable "create_db_instance" {
  description = "Whether to create a new RDS instance"
  type        = bool
  default     = true
}

variable "db_connection_string" {
  description = "Database connection string (used if create_db_instance is false)"
  type        = string
  default     = ""
  sensitive   = true
} 