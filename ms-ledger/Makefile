# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause


.PHONY: build gobuild run gorun stop test lint

MICROSERVICE=automated-checkout/ms-ledger

build:
	docker build --rm \
		--build-arg http_proxy \
		--build-arg https_proxy \
		-f Dockerfile \
		-t $(MICROSERVICE):dev \
		.

gobuild:
	CGO_ENABLED=1 GOOS=linux go build -ldflags='-s -w' -a -installsuffix cgo main.go

run:
	docker run \
		--rm \
		-p 48093:48093 \
		$(MICROSERVICE):dev

gorun:
	./main

stop:
	docker rm -f $(MICROSERVICE):dev

test:
	go test -test.v -cover ./...

testHTML:
	go test -test.v -coverprofile=test_coverage.out ./... && \
	go tool cover -html=test_coverage.out

GOLANGCI_VERSION := $(shell golangci-lint --version 2>/dev/null)

lint:
ifdef GOLANGCI_VERSION
	golangci-lint run
else
	@echo "golangci-lint not found. Please refer to the README documentation for proper installation"
endif
