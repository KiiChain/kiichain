module "prime_validator" {
  source = "./testnet/validator"
  aws_region = "us-east-2"
  instance_id = 0
  make_command = "run-prime-node"
}

# module "validator_1" {
#   source = "./testnet/validator"
#   aws_region = "us-east-2"
#   instance_id = 1
#   make_command = "run-local-node"
# }

# module "validator_2" {
#   source = "./testnet/validator"
#   aws_region = "us-east-2"
#   instance_id = 2
#   make_command = "run-local-node"
# }

# module "testnet_sentry_1" {
#   source = "./testnet/sentry"
#   aws_region = "us-east-2"
#   instance_id = 2
# }

# module "testnet_sentry_2" {
#   source = "./mainnet/validator"
#   aws_region = "us-east-2"
#   instance_id = 1
# }
