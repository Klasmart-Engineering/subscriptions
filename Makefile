
check-docs:
	which swagger || (GO111MODULE=off go get -u github.com/go-swagger/go-swagger/cmd/swagger)

docs: check-docs
	GO111MODULE=off swagger generate spec -o ./docs.yaml --scan-models

serve-docs: check-docs docs
	swagger serve -F=swagger docs.yaml
