#!/bin/bash
set -e

docker-compose -f docker-compose-integration-tests.yml down --remove-orphans
docker-compose -f docker-compose-integration-tests.yml rm -fv

rm -rf ./postgres-integration-data

docker-compose -f docker-compose-integration-tests.yml up -d --build web

go test -v ./test/integration/...

docker-compose -f docker-compose-integration-tests.yml down --remove-orphans
docker-compose -f docker-compose-integration-tests.yml rm -fv
