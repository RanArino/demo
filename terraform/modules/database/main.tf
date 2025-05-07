resource "aws_db_subnet_group" "default" {
  count      = var.create_db_instance && length(var.subnet_ids) > 0 ? 1 : 0
  name       = "${var.db_identifier}-subnet-group"
  subnet_ids = var.subnet_ids

  tags = {
    Name        = "${var.db_identifier}-subnet-group"
    Environment = var.environment
  }
}

resource "aws_security_group" "db_sg" {
  count       = var.create_db_instance && var.vpc_id != "" ? 1 : 0
  name        = "${var.db_identifier}-sg"
  description = "Security group for ${var.db_identifier} database"
  vpc_id      = var.vpc_id

  # PostgreSQL port
  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # Restrict in production
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.db_identifier}-sg"
    Environment = var.environment
  }
}

resource "aws_db_instance" "postgres" {
  count                = var.create_db_instance ? 1 : 0
  identifier           = var.db_identifier
  engine               = "postgres"
  engine_version       = "14"
  instance_class       = "db.t3.micro"
  allocated_storage    = 20
  max_allocated_storage = 100
  
  # Database credentials
  db_name              = var.db_name
  username             = var.db_username
  password             = var.db_password
  
  # Network configuration
  db_subnet_group_name   = length(var.subnet_ids) > 0 ? aws_db_subnet_group.default[0].name : null
  vpc_security_group_ids = var.vpc_id != "" ? [aws_security_group.db_sg[0].id] : null
  publicly_accessible    = true # Set to false in production
  
  # Backup and maintenance
  backup_retention_period = 7
  maintenance_window      = "Mon:00:00-Mon:03:00"
  backup_window           = "03:00-06:00"
  
  # Performance and reliability
  multi_az               = false # Set to true in production
  storage_type           = "gp2"
  storage_encrypted      = true
  
  # Skip final snapshot for easier cleanup (not recommended for production)
  skip_final_snapshot    = true
  
  # Enable deletion protection in production
  deletion_protection    = false
  
  tags = {
    Name        = var.db_identifier
    Environment = var.environment
  }
} 