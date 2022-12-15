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
    "region":"nyc1",
    "size":"s-1vcpu-512mb-10gb",
    "image":"ubuntu-22-10-x64",
    "ssh_keys":["b7:fd:cc:bb:2b:74:e4:65:71:e2:bb:3e:19:3f:f6:d8"],
    "backups":false,
    "ipv6":true,
    "monitoring":true,
    "with_droplet_agent":true,
    "tags":["env:prod","nacre:web"],
    "user_data":"'"$(cat user_data.sh)"'"
  }' \
  "https://api.digitalocean.com/v2/droplets")
echo $response | jq