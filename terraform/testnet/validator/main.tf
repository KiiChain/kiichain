provider "aws" {
  region = var.aws_region
  profile = "kiichain"
}

resource "aws_instance" "validator" {
  ami                         = "ami-024b5075fd81ab5d8"  # Update as needed
  instance_type               = "t2.xlarge"
  root_block_device {
    volume_size = 100
  }

  vpc_security_group_ids = [aws_security_group.validator_sg.id]

  user_data = <<-EOF
      #!/bin/bash
      export NODE_ID=${var.instance_id}
    
      echo "Starting user data script..." >> /tmp/userdata.log

      sudo apt-get update -y >> /tmp/userdata.log 2>&1
      sudo apt-get install -y build-essential docker.io git make wget >> /tmp/userdata.log 2>&1

      sudo systemctl start docker >> /tmp/userdata.log 2>&1
      sudo systemctl enable docker >> /tmp/userdata.log 2>&1
      sudo usermod -aG docker ubuntu >> /tmp/userdata.log 2>&1

      wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz >> /tmp/userdata.log 2>&1
      sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz >> /tmp/userdata.log 2>&1

      echo "export PATH=\$PATH:/usr/local/go/bin" >> /home/ubuntu/.profile
      source /home/ubuntu/.profile >> /tmp/userdata.log 2>&1

      git clone https://<TOKEN>@github.com/KiiChain/kiichain3.git >> /tmp/userdata.log 2>&1

      cd kiichain3 >> /tmp/userdata.log 2>&1
      make run-local-node >> /tmp/userdata.log 2>&1 || echo "Make command failed" >> /tmp/userdata.log
      EOF

  tags = {
    Name = "Testnet Validator - ${var.instance_id}"
  }
}

resource "aws_security_group" "validator_sg" {
  name_prefix = "validator_sg_"

  ingress {
    from_port   = 26668
    to_port     = 26670
    protocol    = "tcp"
    # cidr_blocks = ["172.31.0.0/16"]
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_vpc_security_group_ingress_rule" "allow_ssh" {
  security_group_id = aws_security_group.validator_sg.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 22
  ip_protocol       = "tcp"
  to_port           = 22
}
