# AWS DNS Updater

Updates AWS Route53 DNS records with the network's external IP address (using https://checkip.amazonaws.com). 

Uses a `.env` file to read the following:

- `DOMAIN`: the exact domain to be updated, including trailing dot (e.g. test.com.)
- `TYPE`: the type of record you want to update
- `RECORDS`: comma-separated list of records to be udpated (e.g. test.com,\*.test.com)
- `HOSTED_ZONE_ID`: AWS Hosted Zone ID to be updated
- `AWS_ACCESS_KEY_ID`: AWS access key id
- `AWS_SECRET_ACCESS_KEY`: AWS secret access key

## Build

- `docker build --tag <my-tag> .`

## Run

- `docker run --rm -v ${PWD}/.env:/.env <my-tag>`

