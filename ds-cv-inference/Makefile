# Copyright © 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

.PHONY: build run down

MICROSERVICE=automated-vending/ds-cv-inference

tidy:
	go mod tidy

docker:
	docker build --rm \
		--build-arg http_proxy \
		--build-arg https_proxy \
		-f Dockerfile \
		-t $(MICROSERVICE):dev \
		.

run:
	docker-compose up -d

down:
	docker-compose down

test:
	echo "Pass"