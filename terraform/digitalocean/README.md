# Digital Ocean deployment

## Requirements

- Digital Ocean account
- API token (Go to API - Personal access tokens) and generate Personal access token (with write permissions)
- `terraform` (1.0+) installed

## Deploy

To deploy:

```sh
terraform init
terraform plan -var-file=env/template.tfvars
terraform apply -var-file=env/template.tfvars
```

After deployment (usually takes 5-10 mins) go to [Apps List](https://cloud.digitalocean.com/apps), find an app with name `podsync` and check Runtime Logs.

## Destroy

To destroy:

```sh
terraform destroy -var-file=env/template.tfvars
```
