#!/usr/bin/env bash
if [ -z ${DIGITALOCEAN_TOKEN} ]; then
    echo "DIGITALOCEAN_TOKEN is unset or empty"
    echo $DIGITALOCEAN_TOKEN
    exit -1
fi
response=$(curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
  -d '{
    "region":"nyc1"
    }' \
  "https://api.digitalocean.com/v2/reserved_ips")
echo $response | jq
# {"reserved_ip":{... "ip":".."}}