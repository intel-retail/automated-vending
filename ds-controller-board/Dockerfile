# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM automated-checkout/build:latest AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2022: Intel'

WORKDIR /app
COPY . .

RUN go mod download
# Compile the code
RUN make gobuild

# Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2022: Intel'

RUN apk add --update --no-cache zeromq dumb-init

COPY --from=builder /app/res/profiles/ds-controller-board.yaml /res/profiles/ds-controller-board.yaml
COPY --from=builder /app/res/devices/device-list.toml /res/devices/device-list.toml
COPY --from=builder /app/res/configuration.toml /res/configuration.toml
COPY --from=builder /app/ds-controller-board /ds-controller-board

CMD [ "/ds-controller-board", "-cp=consul.http://edgex-core-consul:8500", "-r" ]
