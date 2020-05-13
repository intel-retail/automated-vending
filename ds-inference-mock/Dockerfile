# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM python:alpine3.9

RUN pip install paho-mqtt
COPY inference_mock.py /inference_mock.py

ENTRYPOINT ["/inference_mock.py"]
