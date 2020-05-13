# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM automated-checkout/build:latest AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2019: Intel'

WORKDIR /app
COPY . .

# Compile the code
RUN make gobuild

# Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2019: Intel'

COPY --from=builder /app/res/ds-controller-board.yaml /res/ds-controller-board.yaml
COPY --from=builder /app/res/docker/configuration.toml /res/docker/configuration.toml
COPY --from=builder /app/ds-controller-board /ds-controller-board

CMD [ "/ds-controller-board","--profile=docker","--confdir=/res", "-r", "consul://edgex-core-consul:8500"]
