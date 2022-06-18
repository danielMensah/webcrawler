install:
	go get -v ./...

lint.fmt:
	go fmt ./...;

lint.vet:
	go vet ./...;

lint.golangci:
	golangci-lint run ./...;

lint.testfmt:
	test -z $(gofmt -s -l -w .);

lint: lint.fmt lint.vet lint.golangci lint.testfmt

clean:
	go clean -testcache;

test: clean
	go test ./...;

build:
	go build -o build/crawler cmd/main.go

make run:
	./build/crawler
