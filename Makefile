.PHONY: build test fmt vet
build:
	go build -o bin/terraform-provider-html-css-to-image .
test:
	go test -race ./...
fmt:
	gofmt -w .
vet:
	go vet ./...
