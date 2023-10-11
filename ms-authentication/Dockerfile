# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM golang:1.18-alpine3.16 AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2022: Intel'

# add git for go modules
RUN apk update && apk add --no-cache make git gcc libc-dev zeromq-dev linux-headers

ENV GO111MODULE=on
WORKDIR /usr/local/bin/

COPY go.mod .
RUN go mod tidy
RUN go mod download

RUN mkdir ms-authentication
WORKDIR /usr/local/bin/ms-authentication/
COPY . .

# Compile the code
RUN make gobuild-authentication

# Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2022: Intel'

RUN apk --no-cache add zeromq
COPY --from=builder /usr/local/bin/ms-authentication/res/configuration.toml /res/configuration.toml
COPY --from=builder /usr/local/bin/ms-authentication/main /ms-authentication
COPY --from=builder /usr/local/bin/ms-authentication/cards.json /cards.json
COPY --from=builder /usr/local/bin/ms-authentication/accounts.json /accounts.json
COPY --from=builder /usr/local/bin/ms-authentication/people.json /people.json

CMD [ "/ms-authentication", "-cp=consul.http://edgex-core-consul:8500", "-r", "-s"]