#!/bin/bash
set -e

./scripts/run-integration-k3d.sh &

until [ \
  "$(curl -s -w '%{http_code}' -o /dev/null "http://localhost:8020/healthcheck")" \
  -eq 200 ]
do
  echo "Waiting for application to start up."
  sleep 1
done

echo "Starting integration tests"
go clean -testcache
go test -v ./test/integration/...

echo "Finished integration tests"

tilt down -f ./environment/integration/Tiltfile

