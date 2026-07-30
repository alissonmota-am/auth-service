variable "aws_region" {
  description = "Regiao AWS"
  type        = string
}

variable "project_name" {
  description = "Nome do projeto/servico (prefixo dos recursos)"
  type        = string
}

# Remote State da plataforma
variable "platform_state_bucket" {
  description = "Bucket S3 onde esta o state da plataforma"
  type        = string
}

variable "platform_state_key" {
  description = "Key do state da plataforma no S3"
  type        = string
}

# RDS
variable "db_instance_class" {
  description = "Classe da instancia RDS"
  type        = string
}

variable "db_name" {
  description = "Nome do banco de dados"
  type        = string
}

variable "db_username" {
  description = "Usuario master do RDS"
  type        = string
}

variable "db_password" {
  description = "Senha do usuario master"
  type        = string
  sensitive   = true
}
