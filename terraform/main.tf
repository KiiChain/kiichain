module "testnet_validator" {
  source = "./testnet/validator"
  aws_region = "us-east-2"
  instance_id = 0
}

# module "testnet_sentry" {
#   source = "./testnet/sentry"
#   aws_region = "us-east-2"
#   instance_id = 2
# }

# module "mainnet_validator" {
#   source = "./mainnet/validator"
#   aws_region = "us-east-2"
#   instance_id = 1
# }

# module "mainnet_sentry" {
#   source = "./mainnet/sentry"
#   aws_region = "us-east-2"
#   instance_id = 1
# }