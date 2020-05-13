# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

# HTTPS_PROXY=http://proxy-chain.site.com:912
# HTTP_PROXY=http://proxy-chain.site.com:911

MICROSERVICE=automated-checkout/ds-inference-mock

DOCKERS=ds-inference-mock

.PHONY: $(DOCKERS)

build: $(DOCKERS)

build:
	docker build \
	--build-arg http_proxy \
	--build-arg https_proxy \
	-f Dockerfile \
	-t $(MICROSERVICE):dev \
	.

run: $(DOCKERS)
	docker run \
	-ti --rm \
	--name inference-mock \
	--mount type=bind,source="$(PWD)"/sku_delta_sequence.json,target=/sku_delta_sequence.json \
	$(MICROSERVICE):dev --broker mqtt-broker --port 1883 --test_sequence sku_delta_sequence.json
