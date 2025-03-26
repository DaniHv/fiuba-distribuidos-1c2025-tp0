#!/bin/bash
output_path=$1
clients=$2

if ! [[ $clients =~ ^[0-9]+$ && $clients -ge 0 ]] ; then
  echo "Invalid usage: clients is expected to be an int greater or equal than 0. Expected usage: $0 {output} {clients}." >&2; exit 1
fi

echo "Generating docker-compose file for $clients clients"

output="name: tp0

# Shared/Commons
x-client-base: &client-base
  image: client:latest
  entrypoint: /client
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
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini
      - ./server/bets.csv:/bets.csv

$(for i in $(seq 1 $clients); do echo "\
  client$i:
    <<: *client-base
    container_name: client$i
    environment:
      - CLI_ID=$i
      - NOMBRE=DAN$i
      - APELLIDO=HUR$i
      - DOCUMENTO=4000000$i
      - NACIMIENTO=2000-01-0$i
      - NUMERO=2000$i
    volumes:
      - ./client/config.yaml:/config.yaml
      - ./.data/agency-$i.csv:/bets.csv
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