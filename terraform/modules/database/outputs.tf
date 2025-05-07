output "connection_string" {
  description = "PostgreSQL connection string"
  value       = var.create_db_instance ? "postgres://${var.db_username}:${var.db_password}@${aws_db_instance.postgres[0].endpoint}/${var.db_name}?sslmode=require" : var.db_connection_string
  sensitive   = true
}

output "endpoint" {
  description = "Database endpoint"
  value       = var.create_db_instance ? aws_db_instance.postgres[0].endpoint : null
  sensitive   = true
}

output "port" {
  description = "Database port"
  value       = var.create_db_instance ? aws_db_instance.postgres[0].port : null
}

output "host" {
  description = "Database host"
  value       = var.create_db_instance ? aws_db_instance.postgres[0].address : null
  sensitive   = true
} 