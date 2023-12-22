resource "aws_vpc" "default" {
  cidr_block = "172.31.0.0/16"
}

resource "aws_subnet" "default" {
  vpc_id     = aws_vpc.default.id
  cidr_block = "172.31.48.0/20"
}
resource "aws_subnet" "backup" {
  vpc_id                  = aws_vpc.default.id
  cidr_block              = "172.31.32.0/20"
  map_public_ip_on_launch = true
}

module "eks" {
  source                         = "terraform-aws-modules/eks/aws"
  version                        = "19.21.0"
  cluster_name                   = "cameras-synthetic-animal-com"
  vpc_id                         = aws_vpc.default.id
  subnet_ids                     = [aws_subnet.default.id, aws_subnet.backup.id]
  cluster_endpoint_public_access = true
}

provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    # This requires the awscli to be installed locally where Terraform is executed
    args = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
  }
}
