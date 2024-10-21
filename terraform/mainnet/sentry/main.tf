provider "aws" {
  region = var.aws_region
  profile = "kiichain"
}

resource "aws_instance" "sentry" {
  ami                         = "" # FILL IN
  instance_type               = "" # FILL IN
  count                       = var.instance_count
  root_block_device {
    volume_size = 0  # FILL IN
  }

  vpc_security_group_ids = [aws_security_group.sentry_sg.id]

  user_data = <<-EOF
              #!/bin/bash
              sudo yum update -y
              sudo amazon-linux-extras install docker -y
              sudo service docker start
              # Similar setup as validators but skip validator-specific scripts
              EOF

  tags = {
    Name = "Testnet Sentry"
  }
}

resource "aws_security_group" "sentry_sg" {
  name_prefix = "sentry_sg_"

  # Allow public access
  ingress {
    from_port   = 0
    to_port     = 65535
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 65535
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}