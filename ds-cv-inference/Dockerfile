# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

ARG OPENVINO_BASE=ubuntu20_dev
ARG OPENVINO_VERSION=2022.2.0
ARG OPENCV_VERSION=4.6.0

FROM gocv/opencv:${OPENCV_VERSION} AS gocv

FROM openvino/${OPENVINO_BASE}:${OPENVINO_VERSION}

# required for installing dependencies
USER root

COPY --from=gocv /usr/local /usr/local

RUN cd /opt/intel/openvino && \
      install_dependencies/install_openvino_dependencies.sh -c=opencv_req -c=opencv_opt -y

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

RUN apt-get update && apt-get install libgtk2.0-dev libgtk-3-dev libjpeg62 -y
WORKDIR /go/src/ds-cv-inference

COPY go.* /go/src/ds-cv-inference/
RUN go mod download

COPY . /go/src/ds-cv-inference

RUN /bin/bash -c "source /opt/intel/openvino/setupvars.sh && echo building ds-cv-inference... && go build -o ds-cv-inference" && \
      chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
