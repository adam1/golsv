
date=$(shell date '+%Y%m%d-%H%M%S')

build all:
	go build
	mkdir -p bin
	go build -o bin ./cmd/...

test t: build
	go test

benchmark bench:
	go test -bench=.

watch w:
	bash scripts/watch.bash

coverage:
	go test -cover

coverage-report:
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

lines:
	find . -name '*.go' | xargs wc -l

clean:
	go clean
	rm -rf bin
	rm -f coverage.out coverage.html
