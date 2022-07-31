#!/bin/bash
#try to connect to google to determine whether user need to use proxy
curl google.com -o /dev/null --connect-timeout 5 2>/dev/null
if [ $? == 0 ]
then
    echo "Successfully connected to Google, no need to use Go proxy"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server .
else
    echo "Google is blocked, Go proxy is enabled: GOPROXY=https://goproxy.cn,direct"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn,direct go build -ldflags="-w -s" -o server .
fi
