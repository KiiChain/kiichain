module "prime_validator" {
  source = "./testnet/validator"
  aws_region = "us-east-2"
  instance_id = 0
  make_command = "docker-cluster-start"
}

module "sentry_1" {
  source = "./testnet/sentry"
  aws_region = "us-east-2"
  instance_id = 1
  instance_character = "alpha"
  make_command = "run-rpc-node"
}

module "sentry_2" {
  source = "./testnet/sentry"
  aws_region = "us-east-2"
  instance_id = 2
  instance_character = "beta"
  make_command = "run-rpc-node"
}
