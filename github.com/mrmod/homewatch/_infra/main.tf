locals {
  application = jsondecode(file("application.json")).infra
}

module "homewatch_agent_container_repository" {
  authorized_pull_accounts  = local.application.authorized_pull_accounts
  authorized_push_arns      = local.application.authorized_push_arns
  container_repository_name = local.application.container_repository_name
  source                    = "./container_repository"
  tags                      = local.application.tags
}
