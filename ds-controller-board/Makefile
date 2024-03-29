# Copyright © 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

.PHONY: build gobuild run gorun stop test lint


MICROSERVICE=automated-vending/ds-controller-board

ARCH=$(shell uname -m)

tidy:
	go mod tidy

docker: 
	docker build --rm \
		--build-arg http_proxy \
		--build-arg https_proxy \
		-f Dockerfile \
		-t $(MICROSERVICE):dev \
		.

gobuild: tidy
	CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -a -o ds-controller-board

run:
	docker run \
		--rm \
		-p 48097:48097 \
		$(MICROSERVICE):dev

gorun:
	./main

stop:
	docker rm -f $(MICROSERVICE):dev

test:
	go test -test.v \
		-cover \
		./...

testHTML:
	go test \
		-test.v \
		-coverprofile=test_coverage.out \
		./... && \
	go tool cover -html=test_coverage.out

testPhysical:
	sudo -E go test -test.v \
		-cover \
		-tags=physical \
		./...

testPhysicalHTML:
	sudo -E go test \
		-tags=physical \
		-test.v \
		-coverprofile=test_coverage_physical.out \
		./... && \
	go tool cover -html=test_coverage_physical.out

testAll:
	sudo -E go test \
		-test.v \
		-cover \
		-tags=all \
		./...

testAllHTML:
	sudo -E go test \
		-tags=all \
		-test.v \
		-coverprofile=test_coverage_all.out \
		./... && \
	go tool cover \
		-html=test_coverage_all.out


GOLANGCI_VERSION := $(shell golangci-lint --version 2>/dev/null)

lint:
	@which golangci-lint >/dev/null || echo "WARNING: go linter not installed. To install, run\n  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.47.3"
	@if [ "z${ARCH}" = "zx86_64" ] && which golangci-lint >/dev/null ; then golangci-lint run --config .golangci.yml ; else echo "WARNING: Linting skipped (not on x86_64 or linter not installed)"; fi