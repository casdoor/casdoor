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
gitCommit=$(git rev-parse HEAD)
gitTag=$(git describe --tags --abbrev=0)
gitDesc=$(git describe --tags)
flags="-w -s \
-X github.com/casdoor/casdoor/util.Commit=${gitCommit} \
-X github.com/casdoor/casdoor/util.Version=${gitTag} \
-X github.com/casdoor/casdoor/util.Desc=${gitDesc}
"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags="""${flags}""" -o server_linux_amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build --ldflags="""${flags}""" -o server_linux_arm64 .
