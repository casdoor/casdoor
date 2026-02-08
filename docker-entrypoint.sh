#!/bin/bash

if [ -z "${driverName:-}" ]; then
  export driverName=sqlite
fi
if [ -z "${dataSourceName:-}" ]; then
  export dataSourceName="file:casdoor.db?cache=shared"
fi
if [ -z "${runmode:-}" ]; then
  export runmode=prod
fi

exec /server
