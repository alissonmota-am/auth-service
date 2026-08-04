################################################################################
# Remote State — dados da plataforma (VPC + EKS)
################################################################################
data "terraform_remote_state" "platform" {
  backend = "s3"

  config = {
    bucket = var.platform_state_bucket
    key    = var.platform_state_key
    region = var.aws_region
  }
}

################################################################################
# RDS PostgreSQL
################################################################################
module "rds" {
  source = "git::https://github.com/alissonmota-am/toggle-master-infra.git//modules/rds?ref=develop"

  project_name               = var.project_name
  vpc_id                     = data.terraform_remote_state.platform.outputs.vpc_id
  subnet_ids                 = data.terraform_remote_state.platform.outputs.private_subnet_ids
  eks_node_security_group_id = data.terraform_remote_state.platform.outputs.eks_node_security_group_id
  instance_class             = var.db_instance_class
  db_name                    = var.db_name
  username                   = var.db_username
  password                   = var.db_password
}
