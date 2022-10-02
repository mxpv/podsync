resource "digitalocean_app" "podsync" {
  spec {
    name   = "podsync"
    region = var.region

    worker {
      name               = "podsync-service"
      instance_count     = var.instance_count
      instance_size_slug = var.instance_size_slug
      dockerfile_path    = "Dockerfile"

      env {
        key   = "PODSYNC_CONFIG_GIST_ID"
        value = var.gist_id
      }
      env {
        key   = "PODSYNC_GITHUB_TOKEN"
        value = var.github_token
      }
      env {
        key   = "PODSYNC_YOUTUBE_API_KEY"
        value = var.youtube_api_key
      }
      env {
        key   = "AWS_ACCESS_KEY_ID"
        value = var.access_id
      }
      env {
        key   = "AWS_SECRET_ACCESS_KEY"
        value = var.secret_key
      }

      git {
        repo_clone_url = var.repo
        branch         = "terraform"
      }
    }
  }
}

#resource "digitalocean_spaces_bucket" "podsync-s3" {
#  name   = "podsync"
#  region = var.region
#  acl = "public-read"
#}