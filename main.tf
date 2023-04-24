# main.tf

# Configure the Google Cloud provider
provider "google" {
  credentials = file(var.gcp_credentials_file)
  project     = var.gcp_project_id
  region      = "us-central1"
}

# Automate the Docker build and push process
resource "null_resource" "docker_build_push" {
  triggers = {
    always_run = "${timestamp()}"
  }

  provisioner "local-exec" {
    command = <<EOT
      docker build -t ${var.docker_hub_username}/dpshade22/go-scripture:latest .
      docker login -u ${var.docker_hub_username} -p ${var.docker_hub_password}
      docker push ${var.docker_hub_username}/dpshade22/go-scripture:latest
EOT
  }
}

# Reference the existing Google Cloud Run service
resource "google_cloud_run_service" "api_service" {
  name     = "<YOUR-EXISTING-CLOUD-RUN-SERVICE-NAME>"
  location = "us-central1"

  template {
    spec {
      containers {
        image = "gcr.io/${var.gcp_project_id}/dpshade22/go-scripture:latest"
      }
    }
  }

  # Ensure the new Docker image is pushed before deploying to Cloud Run
  depends_on = [null_resource.docker_build_push]

  traffic {
    percent         = 100
    latest_revision = true
  }
}
