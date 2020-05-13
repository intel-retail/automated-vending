# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

SHELL := /bin/bash

.PHONY: build

MICROSERVICE= automated-checkout/build

build:
	docker build \
		--build-arg http_proxy \
		--build-arg https_proxy \
		-f Dockerfile \
		-t $(MICROSERVICE):latest \
		.; \
