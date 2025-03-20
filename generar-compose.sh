#!/bin/bash
output_path=$1
clients=$2

if ! [[ $clients =~ ^[0-9]+$ && $clients -ge 1 ]] ; then
  echo "Invalid usage: clients is expected to be an int greater than 0. Expected usage: $0 {output} {clients}." >&2; exit 1
fi

echo "Generating docker-compose file for $clients clients"

output="name: tp0

# Shared/Commons
x-client-base: &client-base
  container_name: client$i
  image: client:latest
  entrypoint: /client
  environment:
    - CLI_ID=1
    - CLI_LOG_LEVEL=DEBUG
  networks:
    - testing_net
  depends_on:
    - server

services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net

$(for i in $(seq 1 $clients); do echo "\
  client$i:
    container_name: client$i
    <<: *client-base
  "
  done
)
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"

echo "$output" > $output_path
echo "Docker-compose file successfully generated at $output_path"