# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM automated-checkout/build:latest AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2022: Intel'

RUN mkdir as-controller-board-status
WORKDIR /usr/local/bin/as-controller-board-status/
COPY . .

# Compile the code
RUN make gobuild

# Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2022: Intel'

RUN apk --no-cache add zeromq
COPY --from=builder /usr/local/bin/as-controller-board-status/res/configuration.toml /res/configuration.toml
COPY --from=builder /usr/local/bin/as-controller-board-status/main /as-controller-board-status

CMD [ "/as-controller-board-status", "-cp=consul.http://edgex-core-consul:8500", "-r"]
