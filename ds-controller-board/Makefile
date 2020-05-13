# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

.PHONY: build gobuild run gorun stop test lint


MICROSERVICE=automated-checkout/ds-controller-board

build:
	docker build --rm \
		--build-arg http_proxy \
		--build-arg https_proxy \
		-f Dockerfile \
		-t $(MICROSERVICE):dev \
		.

gobuild:
	CGO_ENABLED=1 GOOS=linux go build -ldflags='-s -w' -a -installsuffix cgo -o ds-controller-board

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
ifdef GOLANGCI_VERSION
	golangci-lint run
else
	@echo "golangci-lint not found. Please refer to the README documentation for proper installation"
endif
