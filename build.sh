#!/bin/bash
#try to connect to google to determine whether user need to use proxy
curl www.google.com -o /dev/null --connect-timeout 5 2> /dev/null
if [ $? == 0 ]
then
    echo "Successfully connected to Google, no need to use Go proxy"
else
    echo "Google is blocked, Go proxy is enabled: GOPROXY=https://goproxy.cn,direct"
    export GOPROXY="https://goproxy.cn,direct"
fi
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server_linux_amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o server_linux_arm64 .
go test -v -run TestGetVersionInfo ./util/system_test.go ./util/system.go | grep -oE  "v[0-9]+.[0-9]+.[0-9]+\s+[0-9a-f]+\s+[0-9]+" > version_info.txt
