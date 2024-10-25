module "testnet_validator" {
  source = "./testnet/validator"
  aws_region = "us-east-2"
  instance_count = 1
}

# module "testnet_sentry" {
#   source = "./testnet/sentry"
#   aws_region = "us-east-2"
#   instance_count = 2
# }

# module "mainnet_validator" {
#   source = "./mainnet/validator"
#   aws_region = "us-east-2"
#   instance_count = 1
# }

# module "mainnet_sentry" {
#   source = "./mainnet/sentry"
#   aws_region = "us-east-2"
#   instance_count = 1
# }