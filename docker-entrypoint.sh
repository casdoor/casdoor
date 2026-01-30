#!/bin/bash
if [ "${MYSQL_ROOT_PASSWORD}" = "" ] ;then MYSQL_ROOT_PASSWORD=123456 ;fi

# Initialize MariaDB data directory if it doesn't exist
if [ ! -d "/var/lib/mysql/mysql" ]; then
    echo "Initializing MariaDB data directory..."
    mysql_install_db --user=mysql --datadir=/var/lib/mysql
fi

# Ensure proper permissions
chown -R mysql:mysql /var/lib/mysql
chmod -R 755 /var/lib/mysql

service mariadb start

# Wait for MariaDB to be ready
echo "Waiting for MariaDB to start..."
for i in {1..30}; do
    if mysqladmin ping > /dev/null 2>&1; then
        echo "MariaDB is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "MariaDB failed to start within 30 seconds"
        exit 1
    fi
    sleep 1
done

mysqladmin -u root password ${MYSQL_ROOT_PASSWORD}

exec /server --createDatabase=true
