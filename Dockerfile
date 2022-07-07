FROM node:16.13.0 AS FRONT
WORKDIR /web
COPY ./web .
RUN yarn config set registry https://registry.npmmirror.com
RUN yarn install && yarn run build


FROM golang:1.17.5 AS BACK
WORKDIR /go/src/casdoor
COPY . .
RUN ./build.sh


FROM alpine:latest AS STANDARD
LABEL MAINTAINER="https://casdoor.org/"

WORKDIR /
COPY --from=BACK /go/src/casdoor/server ./server
COPY --from=BACK /go/src/casdoor/swagger ./swagger
COPY --from=BACK /go/src/casdoor/conf/app.conf ./conf/app.conf
COPY --from=FRONT /web/build ./web/build
ENTRYPOINT ["/server"]


FROM debian:latest AS db
RUN apt update \
    && apt install -y \
        mariadb-server \
        mariadb-client \
    && rm -rf /var/lib/apt/lists/*


FROM db AS ALLINONE
LABEL MAINTAINER="https://casdoor.org/"

ENV MYSQL_ROOT_PASSWORD=123456

WORKDIR /
COPY --from=BACK /go/src/casdoor/server ./server
COPY --from=BACK /go/src/casdoor/swagger ./swagger
COPY --from=BACK /go/src/casdoor/docker-entrypoint.sh /docker-entrypoint.sh
COPY --from=BACK /go/src/casdoor/conf/app.conf ./conf/app.conf
COPY --from=FRONT /web/build ./web/build

ENTRYPOINT ["/bin/bash"]
CMD ["/docker-entrypoint.sh"]
