# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause


.PHONY: test lint

test:
	go test -test.v -coverprofile=test_coverage.out ./...

testHTML:
	go test -test.v -coverprofile=test_coverage.out ./... && \
	go tool cover -html=test_coverage.out

lint:
	golangci-lint run
