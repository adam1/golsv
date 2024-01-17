
date=$(shell date '+%Y%m%d-%H%M%S')

build all:
	go build .
	mkdir -p bin
	go build -o bin/align ./cmd/align
	go build -o bin/automorphism ./cmd/automorphism
	go build -o bin/calg-cayley ./cmd/calg-cayley
	go build -o bin/cayley ./cmd/cayley
	go build -o bin/cycles ./cmd/cycles
	go build -o bin/dim ./cmd/dim
	go build -o bin/lift ./cmd/lift
	go build -o bin/menum-cayley ./cmd/menum-cayley
	go build -o bin/multiply ./cmd/multiply
	go build -o bin/shorten ./cmd/shorten
	go build -o bin/smith ./cmd/smith
	go build -o bin/systole ./cmd/systole
	go build -o bin/transpose ./cmd/transpose
	go build -o bin/weight ./cmd/weight

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
	rm -rf bin
	rm -f coverage.out coverage.html
