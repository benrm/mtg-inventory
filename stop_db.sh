#!/bin/sh
set -e

podman stop db
podman rm db
pkill exe
