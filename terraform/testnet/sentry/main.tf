provider "aws" {
  region = var.aws_region
  profile = "kiichain"
}

resource "aws_instance" "sentry" {
  ami                         = "ami-0ea142bd7dc67f09c"  # Ubuntu Server 20.04 LTS AMI (update the AMI ID as needed based on your region)
  instance_type               = "t2.xlarge"
  count                       = var.instance_count
  root_block_device {
    volume_size = 100
  }

  vpc_security_group_ids = [aws_security_group.sentry_sg.id]

  user_data = <<-EOF
              #!/bin/bash
              sudo apt-get update -y
              sudo apt-get install -y docker.io git make
              sudo systemctl start docker
              sudo systemctl enable docker

              # Clone the project repository
              git clone https://github.com/KiiChain/kiichain3.git

              # Change directory to the cloned repo
              cd kiichain3

              # Run the specified make command
              make run-local-node
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