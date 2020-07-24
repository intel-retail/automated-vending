# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM automated-checkout/build:latest AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2020: Intel'

RUN mkdir ds-card-reader
WORKDIR /usr/local/bin/ds-card-reader/
COPY . .

# Compile the code
RUN make gobuild

# Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2020: Intel'

COPY --from=builder /usr/local/bin/ds-card-reader/res/ds-card-reader.yaml /res/ds-card-reader.yaml
COPY --from=builder /usr/local/bin/ds-card-reader/res/configuration.toml /res/configuration.toml
COPY --from=builder /usr/local/bin/ds-card-reader/main /ds-card-reader

CMD [ "/ds-card-reader", "--confdir=/res", "-cp=consul.http://edgex-core-consul:8500", "-r"]
