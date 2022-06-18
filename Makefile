.PHONY: lint

lint.fmt:
	go fmt ./...;

lint.vet:
	go vet ./...;

lint.golangci:
	golangci-lint run ./...;

lint.testfmt:
	test -z $(gofmt -s -l -w .);

lint: lint.fmt lint.vet lint.golangci lint.testfmt
