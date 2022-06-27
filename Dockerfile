FROM golang:1.17.5 AS BACK
WORKDIR /go/src/casdoor
COPY . .
RUN ./build.sh


FROM node:16.13.0 AS FRONT
WORKDIR /web
COPY ./web .
RUN yarn config set registry https://registry.npmmirror.com
RUN yarn install && yarn run build


FROM alpine:latest AS STANDARD
LABEL MAINTAINER="https://casdoor.org/"

WORKDIR /app
COPY --from=BACK /go/src/casdoor/server ./
COPY --from=BACK /go/src/casdoor/conf/app.conf ./conf/app.conf
COPY --from=FRONT /web/build ./web/build
VOLUME /app/files /app/logs
ENTRYPOINT ["/app/server"]


FROM debian:latest AS ALLINONE
LABEL MAINTAINER="https://casdoor.org/"

RUN apt update \
    && apt install -y \
        mariadb-server \
        mariadb-client \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=BACK /go/src/casdoor/server ./
COPY --from=BACK /go/src/casdoor/docker-entrypoint.sh /docker-entrypoint.sh
COPY --from=BACK /go/src/casdoor/conf/app.conf ./conf/app.conf
COPY --from=FRONT /web/build ./web/build

ENTRYPOINT ["/bin/bash"]
CMD ["/docker-entrypoint.sh"]
