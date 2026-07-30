output "rds_endpoint" {
  description = "Endpoint do RDS (usar em DATABASE_URL)"
  value       = module.rds.endpoint
}

output "rds_port" {
  description = "Porta do RDS"
  value       = module.rds.port
}

output "db_name" {
  description = "Nome do banco de dados"
  value       = module.rds.db_name
}
