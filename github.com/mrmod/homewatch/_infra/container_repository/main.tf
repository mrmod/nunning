locals {
  allow_docker_login = [
    "ecr:GetAuthorizationToken",
  ]

  all_resources = [
    "*"
  ]

  push_actions = [
    "ecr:CompleteLayerUpload",
    "ecr:PutImage",
    "ecr:UploadLayerPart",
  ]
}

variable "container_repository_name" {
  description = "The name of the container repository"
  type        = string
}

variable "tags" {
  type = map(any)
}

variable "authorized_pull_accounts" {
  description = "List of AWS principals authorzied to pull"
  type        = list(string)
}

variable "authorized_push_arns" {
  description = "List of AWS principals authorzied to push"
  type        = list(string)
}

resource "aws_ecr_repository" "container_repository" {
  image_tag_mutability = "MUTABLE"
  name                 = var.container_repository_name
  tags                 = var.tags
}

data "aws_iam_policy_document" "pull_container_repository_policy" {
  statement {
    actions = [
      "ecr:BatchCheckLayerAvailability",
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer",
      "ecr:InitiateLayerUpload",
      "ecr:ListImages",
    ]

    principals {
      type        = "AWS"
      identifiers = var.authorized_pull_accounts
    }
  }

  statement {
    actions = local.allow_docker_login

    principals {
      type        = "AWS"
      identifiers = var.authorized_pull_accounts
    }
  }
  statement {
    principals {
      type        = "AWS"
      identifiers = var.authorized_push_arns
    }
    actions = local.push_actions
  }
}

resource "aws_ecr_repository_policy" "container_policy" {
  policy     = data.aws_iam_policy_document.pull_container_repository_policy.json
  repository = aws_ecr_repository.container_repository.name
}

output "repository_name" {
  value = aws_ecr_repository.container_repository.name

}
output "repository_arn" {
  value = aws_ecr_repository.container_repository
}
