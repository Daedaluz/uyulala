#!/usr/bin/bash

docker compose -f compose/docker-compose.dev.yml up -d
goland . &
webstorm ./frontend &

