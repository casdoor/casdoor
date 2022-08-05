#!/bin/bash
#try to connect to google to determine whether user need to use proxy
curl google.com -o /dev/null --connect-timeout 5 2> /dev/null
if [ $? == 0 ]
then
    echo "Successfully connected to Google, no need to use Go proxy"
    GOOS=linux go mod tidy
    GOOS=linux go build -ldflags="-linkmode external -extldflags '-static' -w -s" -o server .
else
    echo "Google is blocked, Go proxy is enabled: GOPROXY=https://goproxy.cn,direct"
    GOOS=linux GOPROXY=https://goproxy.cn,direct go mod tidy
    GOOS=linux GOPROXY=https://goproxy.cn,direct go build -ldflags="-linkmode external -extldflags '-static' -w -s" -o server .
fi
