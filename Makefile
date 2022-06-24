build:
	go build -o ./subscriptions-app ./src/main.go

test-unit:
	go test -v ./test/unit/...

test-integration: kill-k3d setup-k3d
	./scripts/run-integration-tests-k3d.sh

openapi-generate:
	./scripts/openapi-generate.sh

setup-k3d:
	./scripts/setup-k3d.sh

run-k3d:
	./scripts/run-k3d.sh

kill-k3d:
	k3d cluster delete factory
