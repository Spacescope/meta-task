COMMIT_ID = $(shell git rev-parse --short HEAD)
ifeq ($(COMMIT_ID),)
COMMIT_ID = 'latest'
endif

.PHONY: test
IMAGE_PREFIX ?= 129862287110.dkr.ecr.us-east-2.amazonaws.com/data-infra
REGISTRY_SERVER ?= 129862287110.dkr.ecr.us-east-2.amazonaws.com/

help:
	@echo
	@echo "  binary - build binary"
	@echo "  build-meta-task - build docker images for centos"
	@echo "  swag - regenerate swag"
	@echo "  build-all - build docker images for centos"
	@echo "  push images to docker hub"

swag:
	swag init -g cmd/meta-task/main.go

binary:
	go build -o bin/meta-task cmd/meta-task/main.go

test:
	go clean -testcache
	gotestsum --format pkgname

build-meta-task:
	docker build -t $(IMAGE_PREFIX)/data-meta-task:$(COMMIT_ID) -f build/Dockerfile .

build-all: build-meta-task

push:
	docker push $(IMAGE_PREFIX)/meta-task:$(COMMIT_ID)
