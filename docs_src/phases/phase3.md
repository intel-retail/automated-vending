# Phase 3 - Bring Your Own Hardware and Software

## Overview

After [phase 2](../phases/phase2.md) has been completed, the next step is to integrate physical hardware. This guide will assist you in understanding the pieces of hardware that are needed in the Automated Vending reference implementation.

## Getting Started

### Step 1: Arduino micro-controller board

The first piece of hardware that will control the vast majority of functionality will be an Arduino micro-controller board with multiple sensors.

Specifically, it handles the following integrations:

| Module                       | Description                                                                                                               |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| Door Sensor (open/close)     | Instrumentation used to understand if the cooler door is open or closed and take appropriate action.                      |
| Maglock (lock/unlock)        | Instrumentation is used to lock or unlock the automated vending to allow for vending of a product.                       |
| LCD Display                  | Instrumentation is used for providing feedback to the person that is using the automated vending.                        |
| LED Panel                    | Instrumentation is used to provide feedback to the developer as to the state of the board lock and door status.           |
| Temperature/Humidity Sensor  | Instrumentation is used to understand the physical environment inside the automated vending and take appropriate action. |

We have created a reference implementation service that will interact with the controller board and EdgeX core services [here](../automated-vending-services/device_services.md#card-reader).

### Step 2: Integrate your own computer vision hardware and software

Next, in order to be able to provide computer vision capabilities in production environment to the Automated Vending it is necessary to bring your own set of cameras and deep learning model. It is expected by the application services to receive an inventory delta using an MQTT broker. For demonstration and evaluation purposes, we have created a CV inference device service using [Intel openVINO](https://docs.openvinotoolkit.org/) inference engine and [openVINO product detection model](https://docs.openvinotoolkit.org/latest/_models_intel_product_detection_0001_description_product_detection_0001.html) that calculates inventory delta based on preloaded images to simulate stocking and purchasing products.

 Please see more details in the device services section [here](../automated-vending-services/device_services.md#cv-inference).
