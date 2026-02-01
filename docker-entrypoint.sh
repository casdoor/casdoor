#!/bin/bash

if [ -z "${driverName:-}" ]; then
  export driverName=sqlite
fi
if [ -z "${dataSourceName:-}" ]; then
  export dataSourceName="file:casdoor.db?cache=shared"
fi

exec /server
