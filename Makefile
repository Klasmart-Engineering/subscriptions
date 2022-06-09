openapi-generate:
	oapi-codegen -old-config-style -package api -templates ./openapi/templates ./openapi/spec.yaml > ./src/api/api.gen.go && \
	sed -i '' 's/newrelic "github\.com\/newrelic\/go-agent"/"github.com\/newrelic\/go-agent\/v3\/newrelic"/' ./src/api/api.gen.go
