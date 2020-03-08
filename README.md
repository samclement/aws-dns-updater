# AWS DNS Updater

Updates AWS Route53 DNS records with the network's external IP address (using https://checkip.amazonaws.com). 

Configuration is through the following environment variables:

- `DOMAIN`: the exact domain in AWS Route53  to be updated, including trailing dot (e.g. example.com.)
- `TYPE`: the type of record you want to update
- `RECORDS`: comma-separated list of records to be udpated (e.g. example.com,\*.example.com)
- `HOSTED_ZONE_ID`: AWS Hosted Zone ID to be updated
- `AWS_ACCESS_KEY_ID`: AWS access key id
- `AWS_SECRET_ACCESS_KEY`: AWS secret access key

The `env` file includes and example configuration that can be used locally with `source ./env`.

## Build

- `docker build --tag <my-tag> .`

## Run

- `docker run --rm -v ${PWD}/.env:/.env <my-tag>`

## k8s

Create a cronjob to update dns on a schedule:

- Update `k8s/cronjob.yaml` with AWS credentials
- Update `k8s/cronjob.yaml` with envrionment variables
- `kubectl apply -f k8s/cronjob.yaml`

