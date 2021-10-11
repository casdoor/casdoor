FROM golang:1.17-alpine AS BACK
ARG BUILLD_LOCATION=CN
WORKDIR /go/src/casdoor
## cache dependencies
COPY go.mod go.sum ./
RUN if [ "$BUILLD_LOCATION" = "CN" ] ; then GOPROXY=https://goproxy.cn,direct go mod download; else go mod download ; fi
## build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server .

FROM node:14.17.6 AS FRONT
ARG BUILLD_LOCATION=CN
WORKDIR /web
# cache dependencies
COPY ./web/package.json ./web/yarn.lock ./
RUN if [ "$BUILLD_LOCATION" = "CN" ] ; then yarn config set registry https://registry.npm.taobao.org ; fi
RUN yarn install
# build
COPY ./web .
RUN yarn run build

FROM alpine:latest
ARG BUILLD_LOCATION=CN
RUN if [ "$BUILLD_LOCATION" = "CN" ] ; then sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories ; fi
LABEL MAINTAINER="https://casdoor.org/"

COPY --from=BACK /go/src/casdoor/ ./
RUN mkdir -p web/build && apk add --no-cache bash coreutils curl
COPY --from=FRONT /web/build /web/build
EXPOSE 8000
CMD ./server