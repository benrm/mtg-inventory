#!/bin/sh
set -e

podman run -d --name db --env-file ./.env \
    --env MYSQL_DATABASE=mtg_inventory --publish 33006:3306 \
    --volume $(pwd)/sql:/docker-entrypoint-initdb.d \
    --volume mtg-data:/var/lib/mysql \
    docker.io/mysql:8
