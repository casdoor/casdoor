FROM --platform=$BUILDPLATFORM node:18.19.0 AS front
WORKDIR /web
COPY ./web .
RUN yarn install --frozen-lockfile --network-timeout 1000000 && yarn run build


FROM --platform=$BUILDPLATFORM golang:1.20.12 AS back
WORKDIR /go/src/casdoor
COPY . .
RUN ./build.sh
RUN go test -v -run TestGetVersionInfo ./util/system_test.go ./util/system.go > version_info.txt

FROM alpine:latest AS standard
LABEL MAINTAINER="https://casdoor.org/"
ARG USER=casdoor
ARG TARGETOS
ARG TARGETARCH
ENV BUILDX_ARCH="${TARGETOS:-linux}_${TARGETARCH:-amd64}"

RUN sed -i 's/https/http/' /etc/apk/repositories
RUN apk add --update sudo
RUN apk add tzdata
RUN apk add curl
RUN apk add ca-certificates && update-ca-certificates

RUN adduser -D $USER -u 1000 \
    && echo "$USER ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USER \
    && chmod 0440 /etc/sudoers.d/$USER \
    && mkdir logs \
    && chown -R $USER:$USER logs

USER 1000
WORKDIR /
COPY --from=back --chown=$USER:$USER /go/src/casdoor/server_${BUILDX_ARCH} ./server
COPY --from=back --chown=$USER:$USER /go/src/casdoor/swagger ./swagger
COPY --from=back --chown=$USER:$USER /go/src/casdoor/conf/app.conf ./conf/app.conf
COPY --from=back --chown=$USER:$USER /go/src/casdoor/version_info.txt ./go/src/casdoor/version_info.txt
COPY --from=front --chown=$USER:$USER /web/build ./web/build

ENTRYPOINT ["/server"]


FROM debian:latest AS db
RUN apt update \
    && apt install -y \
        mariadb-server \
        mariadb-client \
    && rm -rf /var/lib/apt/lists/*


FROM db AS allinone
LABEL MAINTAINER="https://casdoor.org/"
ARG TARGETOS
ARG TARGETARCH
ENV BUILDX_ARCH="${TARGETOS:-linux}_${TARGETARCH:-amd64}"

RUN apt update
RUN apt install -y ca-certificates && update-ca-certificates

WORKDIR /
COPY --from=back /go/src/casdoor/server_${BUILDX_ARCH} ./server
COPY --from=back /go/src/casdoor/swagger ./swagger
COPY --from=back /go/src/casdoor/docker-entrypoint.sh /docker-entrypoint.sh
COPY --from=back /go/src/casdoor/conf/app.conf ./conf/app.conf
COPY --from=back /go/src/casdoor/version_info.txt ./go/src/casdoor/version_info.txt
COPY --from=front /web/build ./web/build

ENTRYPOINT ["/bin/bash"]
CMD ["/docker-entrypoint.sh"]
