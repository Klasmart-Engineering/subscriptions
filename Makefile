
check-docs:
	which swagger || (go get github.com/go-swagger/go-swagger/cmd/swagger@latest)

docs: check-docs
	swagger generate spec -o ./docs.yaml --scan-models

serve-docs: check-docs docs
	swagger serve -F=swagger docs.yaml
