variable "aws_region" {
  description = "The AWS region to launch instances."
  default     = "us-east-2"
}

variable "instance_count" {
  description = "Number of validator instances to create."
  default     = 1
}