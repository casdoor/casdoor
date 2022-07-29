#!/bin/bash
if [ "${MYSQL_ROOT_PASSWORD}" = "" ] ;then MYSQL_ROOT_PASSWORD=123456 ;fi

service mariadb start

mysqladmin -u root password ${MYSQL_ROOT_PASSWORD}

exec /server --createDatabase=true
