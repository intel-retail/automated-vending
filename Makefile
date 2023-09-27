# Copyright © 2022-2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

.PHONY: clean \
		build \
		docker-rm \
		clean-docker \
		run-portainer \
		run \
		down \
		build-image \

GOREPOS= \
		as-vending \
		as-controller-board-status \
		ds-card-reader \
		ds-controller-board \
		ds-cv-inference \
		ms-authentication \
		ms-inventory \
		ms-ledger \


.PHONY: $(GOREPOS)

getlatest:
	git submodule update --init --recursive --remote

docker-rm:
	-docker rm -f $$(docker ps -aq)

clean-docker: docker-rm
	docker volume prune -f && \
	docker network prune -f

run-portainer:
	docker compose -f docker-compose.portainer.yml up -d

run:
	docker compose -f docker-compose.av.yml -f docker-compose.edgex.yml up -d

run-edgex:
	docker compose -f docker-compose.edgex.yml up -d

run-physical:
	docker compose -f docker-compose.av.yml -f docker-compose.edgex.yml -f docker-compose.physical.card-reader.yml -f docker-compose.physical.controller-board.yml up -d

run-physical-card-reader:
	docker compose -f docker-compose.av.yml -f docker-compose.edgex.yml -f docker-compose.physical.card-reader.yml up -d

run-physical-controller-board:
	docker compose -f docker-compose.av.yml -f docker-compose.edgex.yml -f docker-compose.physical.controller-board.yml up -d

down:
	-docker compose -f docker-compose.av.yml -f docker-compose.edgex.yml stop -t 1
	-docker compose -f docker-compose.av.yml -f docker-compose.edgex.yml down

clean: down docker-rm
	docker rmi -f $$(docker images | grep 'automated-vending' | awk '{print $$3}') && \
	docker volume prune -f && \
	docker network prune -f

docker: 
	for repo in ${GOREPOS}; do \
		echo $$repo; \
		cd $$repo; \
		make docker || exit 1; \
		cd ..; \
	done

go-test: 
	for repo in ${GOREPOS}; do \
		echo $$repo; \
		cd $$repo; \
		make test || exit 1; \
		cd ..; \
	done

go-lint: go-tidy
	@which golangci-lint >/dev/null || echo "WARNING: go linter not installed. To install, run make install-lint"
	@which golangci-lint >/dev/null ;  echo "running golangci-lint"; golangci-lint version; go version; 
	for repo in ${GOREPOS}; do \
		echo $$repo; \
		cd $$repo; \
		golangci-lint run --config ../.github/.golangci.yml --out-format=line-number >> ../goLintResults.txt ; \
		cd ..; \
	done

go-tidy: 
	for repo in ${GOREPOS}; do \
		echo $$repo; \
		cd $$repo; \
		make tidy || exit 1; \
		cd ..; \
	done

install-go-lint:
	sudo curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b $$(go env GOPATH)/bin v1.51.2

hadolint: 
	docker run --rm -v $(pwd):/repo -i hadolint/hadolint:latest-alpine sh -c "cd /repo && hadolint -f json ./**/Dockerfile" > go-hadolint.json
