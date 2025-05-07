output "vpc_id" {
  description = "ID of the VPC"
  value       = var.create_vpc ? aws_vpc.main[0].id : null
}

output "public_subnet_ids" {
  description = "IDs of the public subnets"
  value       = var.create_vpc ? aws_subnet.public[*].id : null
}

output "private_subnet_ids" {
  description = "IDs of the private subnets"
  value       = var.create_vpc ? aws_subnet.private[*].id : null
}

output "availability_zones" {
  description = "List of availability zones used"
  value       = var.availability_zones
}

output "aws_region" {
  description = "AWS region"
  value       = var.aws_region
} 