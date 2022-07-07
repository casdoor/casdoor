#!/bin/bash

service mariadb start

mysqladmin -u root password ${MYSQL_ROOT_PASSWORD}

exec /server --createDatabase=true
