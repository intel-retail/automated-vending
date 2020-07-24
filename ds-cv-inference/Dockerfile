# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM hybridgroup/gocv:latest as builder

# Install openVINO 2020.3.194
ARG DOWNLOAD_LINK=http://registrationcenter-download.intel.com/akdlm/irc_nas/16670/l_openvino_toolkit_p_2020.3.194.tgz
ARG INSTALL_DIR=/opt/intel/openvino_2020.3.194
ARG TEMP_DIR=/tmp/openvino_installer
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget \
    git \
    cpio \
    sudo \
    lsb-release && \
    rm -rf /var/lib/apt/lists/*
RUN mkdir -p $TEMP_DIR && cd $TEMP_DIR && \
    wget -c $DOWNLOAD_LINK --no-check-certificate && \
    tar xf l_openvino_toolkit*.tgz && \
    cd l_openvino_toolkit* && \
    sed -i 's/decline/accept/g' silent.cfg && \
    ./install.sh -s silent.cfg  && \
    rm -rf $TEMP_DIR && \
    apt autoremove -y

RUN $INSTALL_DIR/install_dependencies/install_openvino_dependencies.sh

WORKDIR /go/src/ds-cv-inference

ADD go.* /go/src/ds-cv-inference/
RUN go mod download

ADD . /go/src/ds-cv-inference
RUN go build -o ds-cv-inference

ENTRYPOINT ["./entrypoint.sh"]
