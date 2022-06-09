#!/bin/bash

oapi-codegen -old-config-style -package api -templates ./openapi/templates ./openapi/spec.yaml > ./src/api/api.gen.go

if [[ $OSTYPE == 'darwin'* ]]; then
  sed -i '' 's/newrelic "github\.com\/newrelic\/go-agent"/"github.com\/newrelic\/go-agent\/v3\/newrelic"/' ./src/api/api.gen.go
else
  sed -i 's/newrelic "github\.com\/newrelic\/go-agent"/"github.com\/newrelic\/go-agent\/v3\/newrelic"/' ./src/api/api.gen.go
fi
