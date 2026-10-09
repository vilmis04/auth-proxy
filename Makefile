.PHONY: build run test vet fmt check docker

build:
	go build -o bin/app ./cmd/app

run:
	go run ./cmd/app

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

# What CI runs.
check: vet test
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:"; gofmt -l .; exit 1)

docker:
	docker build -t auth-proxy .
