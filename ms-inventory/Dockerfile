# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM automated-checkout/build:latest AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2022: Intel'

RUN mkdir ms-inventory
WORKDIR /usr/local/bin/ms-inventory/
COPY . .

# Compile the code
RUN make gobuild

# Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2022: Intel'

RUN apk --no-cache add zeromq
COPY --from=builder /usr/local/bin/ms-inventory/res/configuration.toml /res/configuration.toml
COPY --from=builder /usr/local/bin/ms-inventory/main /ms-inventory
COPY --from=builder /usr/local/bin/ms-inventory/inventory.json /inventory.json
COPY --from=builder /usr/local/bin/ms-inventory/auditlog.json /auditlog.json

RUN chmod 640 /inventory.json /auditlog.json && \
  chown 2000 /inventory.json /auditlog.json

CMD [ "/ms-inventory", "-cp=consul.http://edgex-core-consul:8500", "-r", "-s"]
