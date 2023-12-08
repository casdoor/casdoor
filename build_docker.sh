#!/bin/sh
DOCKER_PREFIX=jumager/casdoor
branch=$(basename $(git rev-parse --abbrev-ref HEAD))
#try to connect to google to determine whether user need to use proxy
curl www.google.com -o /dev/null --connect-timeout 5 2> /dev/null
if [ $? == 0 ]
then
    echo "Successfully connected to Google, no need to use Go proxy"
else
    echo "Google is blocked, Go proxy is enabled: GOPROXY=https://goproxy.cn,direct"
    export GOPROXY="--build-arg GOPROXY=https://goproxy.cn,direct"
fi
docker buildx build --platform linux/arm64,linux/amd64 --target STANDARD $GOPROXY -t ${DOCKER_PREFIX}:${branch} -f Dockerfile --push .
