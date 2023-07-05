#!/bin/bash
if [ "${MYSQL_ROOT_PASSWORD}" = "" ] ;then MYSQL_ROOT_PASSWORD=123456 ;fi

service mariadb start

mysqladmin -u root password "${MYSQL_ROOT_PASSWORD}"

_create_db="true"
[[ -n "$NO_CREATE_DATABASE" ]] && _create_db="false"

exec /server --createDatabase="$_create_db"
