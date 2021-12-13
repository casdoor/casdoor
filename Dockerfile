FROM golang:1.17.5 AS BACK
WORKDIR /go/src/casdoor
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn,direct go build -ldflags="-w -s" -o server . \
    && apt update && apt install wait-for-it && chmod +x /usr/bin/wait-for-it

FROM node:16.13.0 AS FRONT
WORKDIR /web
COPY ./web .
RUN yarn config set registry https://registry.npm.taobao.org
RUN yarn install && yarn run build

FROM alpine:latest
RUN sed -i 's/https/http/' /etc/apk/repositories
RUN apk add curl
LABEL MAINTAINER="https://casdoor.org/"

COPY --from=BACK /go/src/casdoor/ ./
COPY --from=BACK /usr/bin/wait-for-it ./
RUN mkdir -p web/build && apk add --no-cache bash coreutils
COPY --from=FRONT /web/build /web/build
CMD ./wait-for-it db:3306 -- ./server