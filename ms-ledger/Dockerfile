# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM automated-checkout/build:latest AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2022: Intel'

RUN mkdir ms-ledger
WORKDIR /usr/local/bin/ms-ledger/
COPY . .

# Compile the code
RUN make gobuild

# Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2022: Intel'

RUN apk --no-cache add zeromq
COPY --from=builder /usr/local/bin/ms-ledger/res/configuration.toml /res/configuration.toml
COPY --from=builder /usr/local/bin/ms-ledger/main /ms-ledger
COPY --from=builder /usr/local/bin/ms-ledger/ledger.json /ledger.json

RUN chmod 640 /ledger.json && \
  chown 2000 /ledger.json

CMD [ "/ms-ledger","--confdir=/res", "-cp=consul.http://edgex-core-consul:8500", "-r", "-s" ]
