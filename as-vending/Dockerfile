# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM automated-checkout/build:latest AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2022: Intel'

RUN mkdir as-vending
WORKDIR /usr/local/bin/as-vending/
COPY . .

RUN go mod download
# Compile the code
RUN make gobuild

# Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2022: Intel'

RUN apk --no-cache add zeromq
COPY --from=builder /usr/local/bin/as-vending/res/configuration.toml /res/configuration.toml
COPY --from=builder /usr/local/bin/as-vending/main /as-vending

CMD [ "/as-vending", "-cp=consul.http://edgex-core-consul:8500", "-r"]
