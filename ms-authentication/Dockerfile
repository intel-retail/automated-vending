# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM automated-checkout/build:latest AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2020: Intel'

RUN mkdir ms-authentication
WORKDIR /usr/local/bin/ms-authentication/
COPY . .

# Compile the code
RUN make gobuild

# Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
  copyright='Copyright (c) 2020: Intel'

RUN apk --no-cache add zeromq
COPY --from=builder /usr/local/bin/ms-authentication/res/docker/configuration.toml /res/docker/configuration.toml
COPY --from=builder /usr/local/bin/ms-authentication/main /ms-authentication
COPY --from=builder /usr/local/bin/ms-authentication/cards.json /cards.json
COPY --from=builder /usr/local/bin/ms-authentication/accounts.json /accounts.json
COPY --from=builder /usr/local/bin/ms-authentication/people.json /people.json

CMD [ "/ms-authentication","--profile=docker","--confdir=/res", "-r"]
