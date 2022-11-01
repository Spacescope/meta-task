BUILD_DIR = ./build
.PHONY: build test

build:
	go build -o bin/observatory-task cmd/observatorytask/main.go

test:
	@echo ">> make test"
	go clean -testcache
	gotestsum --format pkgname
