aws_region   = "us-east-1"
project_name = "auth-service-prod"

# Remote State da plataforma
platform_state_bucket = "BUCKET_NAME_AQUI"
platform_state_key    = "platform-prod/terraform.tfstate"

# RDS
db_instance_class = "db.t3.micro"
db_name           = "auth_db"
db_username       = "fiap"
