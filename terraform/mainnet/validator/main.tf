provider "aws" {
  region = var.aws_region
  profile = "kiichain"
}

resource "aws_instance" "validator" {
  ami                         = ""  # FILL IN
  instance_type               = ""  # FILL IN
  count                       = var.instance_count
  root_block_device {
    volume_size = 0  # FILL IN
  }

  vpc_security_group_ids = [aws_security_group.validator_sg.id]

  user_data = <<-EOF
              #!/bin/bash
              sudo yum update -y
              sudo amazon-linux-extras install docker -y
              sudo service docker start
              # Run your docker validator setup commands here
              EOF

  tags = {
    Name = "Testnet Validator"
  }
}

resource "aws_security_group" "validator_sg" {
  name_prefix = "validator_sg_"

  # Allow only internal communication

  ingress {
    from_port   = 0
    to_port     = 65535
    protocol    = "-1"
    cidr_blocks = ["10.0.0.0/16"] # Change as per your VPC CIDR
  }

  egress {
    from_port   = 0
    to_port     = 65535
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}