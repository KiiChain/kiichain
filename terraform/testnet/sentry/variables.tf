variable "aws_region" {
  description = "The AWS region to launch instances."
  default     = "us-east-2"
}

variable "instance_id" {
  description = "Instance ID"
  default     = 0
}

variable "make_command" {
  description = "command for make file"
  default     = "run-rpc-node"
}