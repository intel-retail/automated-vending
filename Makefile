# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

.PHONY: clean \
		build \
		docker-rm \
		clean-docker \
		run-portainer \
		run \
		down \
		build-image \

DOCKERS= \
		docker-build-image \
		as-vending \
		as-controller-board-status \
		ds-card-reader \
		ds-controller-board \
		ds-inference-mock \
		ms-authentication \
		ms-inventory \
		ms-ledger

.PHONY: $(DOCKERS)

docker-rm:
	-docker rm $$(docker ps -aq)

clean-docker: docker-rm
	docker volume prune -f && \
	docker network prune -f

run-portainer:
	docker-compose -f docker-compose.portainer.yml up -d

run:
	docker-compose -f docker-compose.yml up -d

run-physical:
	docker-compose -f docker-compose.yml -f docker-compose.physical.card-reader.yml -f docker-compose.physical.controller-board.yml up -d

run-physical-card-reader:
	docker-compose -f docker-compose.yml -f docker-compose.physical.card-reader.yml up -d

run-physical-controller-board:
	docker-compose -f docker-compose.yml -f docker-compose.physical.controller-board.yml up -d

down:
	-docker-compose -f docker-compose.yml stop -t 1
	-docker-compose -f docker-compose.yml down

clean: down docker-rm
	docker rmi -f $$(docker images | grep '<none>' | awk '{print $$3}') && \
	docker rmi -f $$(docker images | grep automated-checkout | awk '{print $$3}') && \
	docker volume prune -f && \
	docker network prune -f

$(DOCKERS):
	cd $@; \
	make build

build : $(DOCKERS)
