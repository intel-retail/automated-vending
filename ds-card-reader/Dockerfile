# Copyright © 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM golang:1.20-alpine3.17 AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2023: Intel'


# add git for go modules
RUN apk update && apk add --no-cache make git gcc libc-dev linux-headers

ENV GO111MODULE=on
WORKDIR /usr/local/bin/

# This caches the packages for use when building the other services.
# Update the go.mod file in this repo when a new package is added to one of the services.
# This will be obvious when building a service and the un-cached package it loaded every build.
COPY go.mod .
RUN go mod tidy
RUN go mod download

RUN mkdir ds-card-reader
WORKDIR /usr/local/bin/ds-card-reader/
COPY . .

# Compile the code
RUN make gobuild

# Next image - Copy built Go binary into new workspace
FROM alpine:3.17

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2023: Intel'

RUN apk add --update --no-cache dumb-init

COPY --from=builder /usr/local/bin/ds-card-reader/res/profiles/ds-card-reader.yaml /res/profiles/ds-card-reader.yaml
COPY --from=builder /usr/local/bin/ds-card-reader/res/devices/device-list.yaml /res/devices/device-list.yaml
COPY --from=builder /usr/local/bin/ds-card-reader/res/configuration.yaml /res/configuration.yaml
COPY --from=builder /usr/local/bin/ds-card-reader/main /ds-card-reader

CMD [ "/ds-card-reader", "-cp=consul.http://edgex-core-consul:8500", "-r"]
