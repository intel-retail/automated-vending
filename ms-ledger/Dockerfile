# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM automated-checkout/build:latest AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2019: Intel'

RUN mkdir ms-ledger
WORKDIR /usr/local/bin/ms-ledger/
COPY . .

# Compile the code
RUN make gobuild

# Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2019: Intel'

RUN apk --no-cache add zeromq
COPY --from=builder /usr/local/bin/ms-ledger/res/docker/configuration.toml /res/docker/configuration.toml
COPY --from=builder /usr/local/bin/ms-ledger/main /ms-ledger
COPY --from=builder /usr/local/bin/ms-ledger/ledger.json /ledger.json

RUN chmod 640 /ledger.json && \
    chown 2000 /ledger.json

CMD [ "/ms-ledger","--profile=docker","--confdir=/res", "-r" ]
