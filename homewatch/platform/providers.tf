terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Configure the AWS Provider
provider "aws" {
  region = "us-west-2"
  default_tags {
    tags = {
      App = "cameras.syntheticanimal.com"
    }
  }
}
