#!/bin/bash
set -e

docker-compose -f docker-compose-integration-tests.yml down --remove-orphans
docker-compose -f docker-compose-integration-tests.yml rm -fv

rm -rf ./postgres-integration-data

docker-compose -f docker-compose-integration-tests.yml up -d --build web

echo "Starting integration tests"

go test -v ./test/integration/...

echo "Finished integration tests"

docker-compose -f docker-compose-integration-tests.yml down --remove-orphans
docker-compose -f docker-compose-integration-tests.yml rm -fv
