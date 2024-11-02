provider "aws" {
  region = var.aws_region
  profile = "kiichain"
}

resource "aws_s3_bucket" "my_bucket" {
  bucket_prefix = "sentry-genesis-"
}

resource "aws_instance" "sentry" {
  ami                         = "ami-024b5075fd81ab5d8"  # Update as needed
  instance_type               = "t2.xlarge"
  root_block_device {
    volume_size = 100
  }

  vpc_security_group_ids = [aws_security_group.sentry_sg.id]

  user_data = <<-EOF
        #!/bin/bash
        export NODE_ID=${var.instance_id}
      
        echo "Starting user data script..." >> /tmp/userdata.log

        # Update package list and install build-essential first to ensure make is available
        sudo apt-get update -y >> /tmp/userdata.log 2>&1
        sudo apt-get install -y build-essential wget git nginx software-properties-common certbot python3-certbot-nginx >> /tmp/userdata.log 2>&1
        sudo apt install apt-transport-https ca-certificates curl software-properties-common >> /tmp/userdata.log 2>&1
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
        sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu focal stable"
        sudo apt install -y docker-ce >> /tmp/userdata.log 2>&1

        # Install Docker packages separately to avoid conflicts
        sudo apt-get install -y containerd.io docker-buildx-plugin docker-compose-plugin >> /tmp/userdata.log 2>&1

        # Start and enable Docker
        sudo systemctl start docker >> /tmp/userdata.log 2>&1
        sudo systemctl enable docker >> /tmp/userdata.log 2>&1
        sudo usermod -aG docker ubuntu >> /tmp/userdata.log 2>&1

        # Install Go 1.21
        wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz >> /tmp/userdata.log 2>&1
        sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz >> /tmp/userdata.log 2>&1

        # Set up Go environment globally for all users
        echo "export PATH=\$PATH:/usr/local/go/bin" | sudo tee -a /etc/profile /home/ubuntu/.profile
        source /home/ubuntu/.profile >> /tmp/userdata.log 2>&1

        # Verify Go installation
        go version >> /tmp/userdata.log 2>&1 || { echo "Go command not found" >> /tmp/userdata.log; exit 1; }

        # Clone the project repository
        git clone https://<TOKEN>@github.com/KiiChain/kiichain3.git >> /tmp/userdata.log 2>&1

        # Navigate to project directory and execute make command
        cd kiichain3 >> /tmp/userdata.log 2>&1
        echo "export PROJECT_HOME=$(git rev-parse --show-toplevel)" | sudo tee -a /etc/profile /home/ubuntu/.profile
        echo "export GO_PKG_PATH=$(HOME)/go/pkg" | sudo tee -a /etc/profile /home/ubuntu/.profile
        echo "export GOCACHE=$(HOME)/.cache/go-build" | sudo tee -a /etc/profile /home/ubuntu/.profile
        source /home/ubuntu/.profile >> /tmp/userdata.log 2>&1

        echo "PROJECT_HOME: $PROJECT_HOME" >> /tmp/userdata.log 2>&1
        echo "GO_PKG_PATH: $GO_PKG_PATH" >> /tmp/userdata.log 2>&1
        echo "GOCACHE: $GOCACHE" >> /tmp/userdata.log 2>&1

        git config --global --add safe.directory $(PROJECT_HOME)

        sed -i 's/REPLACE/${var.instance_character}/g' terraform/default

        cp terraform/default /etc/nginx/sites-available/default
        sudo ln -s /etc/nginx/sites-available/default /etc/nginx/sites-enabled/

        sudo nginx -t >> /tmp/userdata.log 2>&1
        sudo systemctl restart nginx >> /tmp/userdata.log 2>&1

        sudo certbot --nginx -d ${var.instance_character}.sentry.testnet.v3.kiivalidator.com --register-unsafely-without-email --agree-tos >> /tmp/userdata.log 2>&1

        PROJECT_HOME=$PROJECT_HOME GO_PKG_PATH=$GO_PKG_PATH GOCACHE=$GOCACHE make ${var.make_command} >> /tmp/userdata.log 2>&1 || echo "Make command failed" >> /tmp/userdata.log
        echo "User data script completed." >> /tmp/userdata.log
        EOF

  tags = {
    Name = "Testnet Sentry KIICHAIN3 - ${var.instance_id}"
  }
}

resource "aws_security_group" "sentry_sg" {
  name_prefix = "sentry_sg_"

  ingress {
    from_port   = 26656
    to_port     = 26656
    protocol    = "tcp"
    cidr_blocks = ["172.31.0.0/16"]
  }

  egress {
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_vpc_security_group_ingress_rule" "val_1" {
  security_group_id = aws_security_group.sentry_sg.id
  cidr_ipv4         = "172.31.0.0/16"
  from_port         = 26659
  ip_protocol       = "tcp"
  to_port           = 26659
}

resource "aws_vpc_security_group_ingress_rule" "val_2" {
  security_group_id = aws_security_group.sentry_sg.id
  cidr_ipv4         = "172.31.0.0/16"
  from_port         = 26662
  ip_protocol       = "tcp"
  to_port           = 26662
}

resource "aws_vpc_security_group_ingress_rule" "allow_ssh" {
  security_group_id = aws_security_group.sentry_sg.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 22
  ip_protocol       = "tcp"
  to_port           = 22
}

resource "aws_vpc_security_group_ingress_rule" "allow_evm_rpc" {
  security_group_id = aws_security_group.sentry_sg.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 8545
  ip_protocol       = "tcp"
  to_port           = 8546
}

resource "aws_vpc_security_group_ingress_rule" "allow_rest_api" {
  security_group_id = aws_security_group.sentry_sg.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 1317
  ip_protocol       = "tcp"
  to_port           = 1317
}

resource "aws_vpc_security_group_ingress_rule" "allow_rpc" {
  security_group_id = aws_security_group.sentry_sg.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 26657
  ip_protocol       = "tcp"
  to_port           = 26657
}

resource "aws_vpc_security_group_ingress_rule" "allow_https" {
  security_group_id = aws_security_group.sentry_sg.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 443
  ip_protocol       = "tcp"
  to_port           = 443
}

resource "aws_vpc_security_group_ingress_rule" "allow_rpc_ssl" {
  security_group_id = aws_security_group.sentry_sg.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 26671
  ip_protocol       = "tcp"
  to_port           = 26671
}

resource "aws_route53_record" "sentry_record" {
  zone_id = "Z07671973O0LTK82CZVWD"
  name    = "${var.instance_character}.sentry.testnet.v3.kiivalidator.com"
  type    = "A"
  ttl     = 300
  records = [aws_instance.sentry.public_ip]
}

