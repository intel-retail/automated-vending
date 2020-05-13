# Phase 3 - Bring Your Own Hardware and Software

## Overview

After [phase 2](../phases/phase2.md) has been completed, the next step is to integrate physical hardware. This guide will assist you in understanding the pieces of hardware that are needed in the Automated Checkout reference design.

## Getting Started

### Step 1: Arduino micro-controller board

The first piece of hardware that will control the vast majority of functionality will be an Arduino micro-controller board with multiple sensors.

Specifically, it handles the following integrations:

| Module                       | Description                                                                                                               |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| Door Sensor (open/close)     | Instrumentation used to understand if the cooler door is open or closed and take appropriate action.                      |
| Maglock (lock/unlock)        | Instrumentation is used to lock or unlock the automated checkout to allow for vending of a product.                       |
| LCD Display                  | Instrumentation is used for providing feedback to the person that is using the automated checkout.                        |
| LED Panel                    | Instrumentation is used to provide feedback to the developer as to the state of the board lock and door status.           |
| Temperature/Humidity Sensor  | Instrumentation is used to understand the physical environment inside the automated checkout and take appropriate action. |

We have created a reference design service that will interact with the controller board and EdgeX core services [here](../automated-checkout-services/device_services.md#card-reader).

### Step 2: Integrate your own computer vision inference hardware

Next, in order to be able to provide computer vision capabilities to the Automated Checkout it is necessary to bring your own set of cameras and deep learning model. It is expected by the application services to receive an inventory delta using an MQTT broker. For development and testing purposes, we have created a mock inference device service that mimics what a real computer vision inference engine would do. Please see more details in the device services section [here](../automated-checkout-services/device_services.md#inference-mock).
