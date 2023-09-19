# Profile: BoomedUpAdmin
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
    }
  }
  backend "s3" {}
}

# Configure the AWS Provider
provider "aws" {
  region                      = "us-west-2"
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_credentials_validation = true
}

# Resources
resource "random_id" "id" {
  byte_length = 4
}

module "plans_storage" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = ">= 3.0"

  bucket        = "plans-${random_id.id.hex}"
  force_destroy = true

  tags = {
    project = "terrastate"
  }

  versioning = {
    status = true
  }
}

output "storage" {
  value = module.plans_storage.s3_bucket_id
}

module "plans_datastore" {
  source = "terraform-aws-modules/dynamodb-table/aws"

  name      = "plans-${random_id.id.hex}"
  hash_key  = "planWindow"
  range_key = "createdAtUtc"

  attributes = [
    {
      name = "planWindow"
      type = "S"
    },
    {
      name = "createdAtUtc"
      type = "N"
    }
  ]
}

output "datastore" {
  value = module.plans_datastore.dynamodb_table_id
}
