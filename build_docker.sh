#!/bin/sh
set -e -o pipefail
DOCKER_PREFIX=jumager/casdoor
#branch=$(basename $(git rev-parse --abbrev-ref HEAD))
branch=latest
docker buildx build --platform linux/arm64,linux/amd64 --target STANDARD -t ${DOCKER_PREFIX}:${branch} -f Dockerfile --push .
