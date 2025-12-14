.PHONY: build test lint bench clean generate install

build:
	go build -o bin/resetgen .

install:
	go install .

test:
	go test ./... -race -cover

bench:
	go test ./... -bench=. -benchmem

lint:
	golangci-lint run

generate:
	go run . testdata/basic/user.go testdata/embedded/models.go testdata/complex/models.go

clean:
	rm -rf bin/
	rm -f testdata/**/*.gen.go

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

all: lint test build
