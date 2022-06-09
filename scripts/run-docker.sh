#!/bin/bash
set -e

debug=$1

docker-compose down --remove-orphans
docker-compose rm -fv

rm -rf ./postgres-data

if [ -z "$debug" ]
then
  docker-compose up --build web
else
  docker-compose up --build web-debug
fi
