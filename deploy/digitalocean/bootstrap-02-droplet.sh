#!/usr/bin/env bash
if [ -z ${DIGITALOCEAN_TOKEN} ]; then
    echo "DIGITALOCEAN_TOKEN is unset or empty"
    echo $DIGITALOCEAN_TOKEN
    exit -1
fi

if [ -z $1 ]; then
  echo "Missing first argument: reserved IP to assign to the new droplet"
  exit -1
fi

response=$(curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
  -d '{
    "name":"nacre-web-01",
    "region":"sfo3",
    "size":"s-1vcpu-512mb-10gb",
    "image":"ubuntu-23-10-x64",
    "ssh_keys":["c2:b0:59:67:33:19:c8:43:46:7b:e3:e8:78:51:a1:66"],
    "backups":false,
    "ipv6":true,
    "monitoring":true,
    "with_droplet_agent":true,
    "tags":["env:prod","nacre:web"],
    "user_data":"'"$(cat user_data.sh)"'"
  }' \
  "https://api.digitalocean.com/v2/droplets")
echo $response | jq
