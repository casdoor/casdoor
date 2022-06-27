#!/bin/bash

service mariadb start

if [ "${MYSQL_ROOT_PASSWORD}" = "" ]; then
    MYSQL_ROOT_PASSWORD=123456
fi

mysqladmin -u root password ${MYSQL_ROOT_PASSWORD}

exec /app/server --createDatabase=true
