# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause


.PHONY: build gobuild run gorun stop test lint

MICROSERVICE=automated-checkout/ms-ledger

ARCH=$(shell uname -m)

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

test: lint
	go test -test.v -cover ./...

testHTML:
	go test -test.v -coverprofile=test_coverage.out ./... && \
	go tool cover -html=test_coverage.out

lint:
	@which golangci-lint >/dev/null || echo "WARNING: go linter not installed. To install, run\n  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.47.3"
	@if [ "z${ARCH}" = "zx86_64" ] && which golangci-lint >/dev/null ; then golangci-lint run --config .golangci.yml ; else echo "WARNING: Linting skipped (not on x86_64 or linter not installed)"; fi