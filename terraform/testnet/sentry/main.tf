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
    echo "Starting user data script..." >> /tmp/userdata.log

    # Update package lists
    sudo apt-get update -y >> /tmp/userdata.log 2>&1

    # Install required packages
    sudo apt-get install -y build-essential docker.io git make wget >> /tmp/userdata.log 2>&1
    sudo systemctl start docker >> /tmp/userdata.log 2>&1
    sudo systemctl enable docker >> /tmp/userdata.log 2>&1
    sudo usermod -aG docker ubuntu >> /tmp/userdata.log 2>&1

    # Install Go 1.21
    echo "Installing Go 1.21..." >> /tmp/userdata.log
    wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz >> /tmp/userdata.log 2>&1
    sudo rm -rf /usr/local/go >> /tmp/userdata.log 2>&1
    sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz >> /tmp/userdata.log 2>&1

    # Set up Go environment variables
    echo "export PATH=\$PATH:/usr/local/go/bin" >> /home/ubuntu/.profile
    echo "export GOPATH=/home/ubuntu/go" >> /home/ubuntu/.profile
    echo "export GOBIN=\$GOPATH/bin" >> /home/ubuntu/.profile
    echo "export PATH=\$PATH:\$GOBIN" >> /home/ubuntu/.profile

    # Source the profile to apply the changes immediately
    source /home/ubuntu/.profile >> /tmp/userdata.log 2>&1

    echo "Go 1.21 installation completed." >> /tmp/userdata.log

    # Clone the project repository
    echo "Cloning the repository..." >> /tmp/userdata.log
    git clone https://<TOKEN>@github.com/KiiChain/kiichain3.git >> /tmp/userdata.log 2>&1

    cd kiichain3 >> /tmp/userdata.log 2>&1
    echo "Verifying Makefile..." >> /tmp/userdata.log
    cat Makefile >> /tmp/userdata.log 2>&1
    pwd >> /tmp/userdata.log
    ls -la >> /tmp/userdata.log

    # Add a short delay to ensure Docker is up and running
    sleep 10

    # Run the Makefile command and log any failure
    make run-local-node >> /tmp/userdata.log 2>&1 || echo "Make command failed" >> /tmp/userdata.log
    echo "User data script completed." >> /tmp/userdata.log
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