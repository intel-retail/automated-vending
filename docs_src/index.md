# Automated Checkout Reference Implementation

## Introduction

This guide helps you build and run the Automated Checkout Reference Implementation.

Upon completing the steps in this guide, you will be ready to integrate sensors and services to build your own complete solution.

!!! info
    This guide does not create a complete, ready-to-use solution. Instead, upon completing the steps in this guide, you will be ready to integrate sensors and services to build your own Automated Checkout solution.

    Certain third-party software or hardware identified in this document only may be used upon securing a license directly from the third-party software or hardware owner. The identification of non-Intel software, tools, or services in this document does not constitute a sponsorship, endorsement, or warranty by Intel.

<p align="center">
    <img src="./images/automated-checkout.png">
</p>

### Block diagram

The high-level diagram below shows the sensors and services used with the Automated Checkout Reference Implementation. The diagram shows the sensors and services, and how they communicate through EdgeX. Intel provides the services outlined in blue, the purple services are provided by EdgeX, and the black services are either simulated or can interface with real hardware.

![Automated Checkout Reference Implementation Diagram](./images/automated-checkout-reference-implementation.png)

### Prerequisites

The following items are required to build the Automated Checkout Reference Implementation. You will need additional hardware and software when you are ready to build your own solution.

- **A deep learning model for CV inferencing.** Intel provides a reference inference service using openVINO that produces inventory changes based on preloaded images. See [cv inference service](./automated-checkout-services/device_services.md#cv-inference) for information on this process.
- **A device that allows badging-in to the Automated Checkout.** Intel provides a card reader service that can be simulated or integrated with a physical USB device. See the [Card Reader](./automated-checkout-services/device_services.md#card-reader) device service page for information on this service.
- **A controller device that locks the door to the Automated Checkout**, as well as providing readouts (such as a small text-based LCD screen) to display authorization state, items purchased, and other sensor readings. This could be an Arduino-powered circuit. Intel provides a display service that can run in a simulated mode or with a physical USB/serial interface. See the [Controller Board](./automated-checkout-services/device_services.md#controller-board) device service page for implementation details.

- <a href="https://releases.ubuntu.com/20.04.5/" rel="noopener noreferrer" target="_blank">Ubuntu 20.04</a>
- <a href="https://docs.docker.com/install/" rel="noopener noreferrer" target="_blank">Docker</a>
- <a href="https://docs.docker.com/compose/" rel="noopener noreferrer" target="_blank">Docker Compose</a>
- <a href="https://golang.org/doc/devel/release.html" rel="noopener noreferrer" target="_blank">Go 1.12+</a> for development purposes or running without docker.
- <a href="https://git-scm.com/" rel="noopener noreferrer" target="_blank">Git</a>
- <a href="https://www.gnu.org/software/make/" rel="noopener noreferrer" target="_blank">GNU make</a>
- A REST client such as <a href="https://github.com/curl/curl" rel="noopener noreferrer" target="_blank">curl</a> or <a href="https://www.postman.com/" rel="noopener noreferrer" target="_blank">Postman</a> for running through the phases outlined in the documentation.

Here's how to install `git`, `curl`, and `make` on an Ubuntu 20.04 system - other operating systems may vary:

```bash
sudo apt-get update -y
sudo apt-get install -y git curl build-essential
```

### Hardware

For [**Phase 2 - Add Card Reader Device**](./phases/phase2.md), a USB based RFID card reader or second regular USB keyboard can be used.

Additionally, frequently throughout this documentation, we will refer to a "cooler" or "cooler door". This is referring to a vending machine or refrigerator with a sealed location for cooling. Inventory/stock are intended to be placed inside the cooler, and the temperature and humidity inside the cooler are monitored.

### Recommended domain knowledge

- <a href="https://www.edgexfoundry.org/" rel="noopener noreferrer" target="_blank">EdgeX</a> - the Automated Checkout reference implementation utilizes the EdgeX framework
- <a href="https://mqtt.org" rel="noopener noreferrer" target="_blank">MQTT</a>
- <a href="https://en.wikipedia.org/wiki/Representational_state_transfer" rel="noopener noreferrer" target="_blank">REST</a>
- <a href="https://en.wikipedia.org/wiki/Evdev" rel="noopener noreferrer" target="_blank">evdev</a>, if reading input events from a USB input device, such as a card reader that inputs key strokes upon scanning a card
- Communication over serial on Linux, if using serial devices such as Arduino
- Computer Vision concepts, if using CV inferencing components
- Basic RFID concepts, if using RFID components (i.e. for badge-in card reader)
- <a href="https://www.portainer.io/" rel="noopener noreferrer" target="_blank">Portainer</a> - included with the Automated Checkout reference implementation. Usage is optional, but this is a highly recommended utility for managing Docker containers, and we provide easy ways to run it.

## Getting started

### Step 1: Clone the repository

```bash
git clone https://github.com/intel-iot-devkit/automated-checkout.git && cd ./automated-checkout
```

### Step 2: Build the reference implementation

You must build the provided component services and create local docker images. To do so, run:

```bash
make build
```

!!! note
    This command may take a while to run depending on your internet connection and machine specifications.

#### Check for build success

Make sure the command was successful. To do so, run:

```bash
docker images | grep automated-checkout
```

!!! success
    The results are:

    - `automated-checkout/as-controller-board-status`
    - `automated-checkout/as-vending`
    - `automated-checkout/build` (latest tag)
    - `automated-checkout/ds-card-reader`
    - `automated-checkout/ds-controller-board`
    - `automated-checkout/ds-cv-inference`
    - `automated-checkout/ms-authentication`
    - `automated-checkout/ms-inventory`
    - `automated-checkout/ms-ledger`

!!! failure
    If you do not see all of the above docker images, look through the console output for errors. Sometimes dependencies fail to resolve and must be run again. Address obvious issues. To try again, repeat step 2.

### Step 3: Start the reference implementation suite

Use Docker Compose to start the reference implementation suite. To do so, run:

```bash
make run
```

This command starts the EdgeX Device Services and then starts all the Automated Checkout Reference Implementation Services.

#### Check for success

Make sure the command was successful. To do so, run:

```bash
docker ps --format 'table{{.Image}}\t{{.Status}}'
```

!!! success
    Your output is as follows:

    | IMAGE                                                | STATUS            |
    |------------------------------------------------------|-------------------|
    | automated-checkout/ms-ledger:dev                     | Up 53 seconds     |
    | eclipse-mosquitto:1.6.3                              | Up 52 seconds     |
    | automated-checkout/as-vending:dev                    | Up 52 seconds     |
    | automated-checkout/ms-inventory:dev                  | Up 52 seconds     |
    | automated-checkout/ds-controller-board:dev           | Up 52 seconds     |
    | automated-checkout/ms-authentication:dev             | Up 55 seconds     |
    | edgexfoundry/docker-device-mqtt-go:1.2.0             | Up 53 seconds     |
    | automated-checkout/ds-card-reader:dev                | Up 53 seconds     |
    | automated-checkout/as-controller-board-status:dev    | Up 52 seconds     |
    | edgexfoundry/docker-core-command-go:1.2.0            | Up About a minute |
    | edgexfoundry/docker-core-data-go:1.2.0               | Up About a minute |
    | edgexfoundry/docker-core-metadata-go:1.2.0           | Up About a minute |
    | edgexfoundry/docker-support-notifications-go:1.2.0   | Up About a minute |
    | edgexfoundry/docker-edgex-consul:1.2.0               | Up About a minute |
    | automated-checkout/ds-cv-inference:dev             | Up 51 seconds     |
    | redis:5.0.8-alpine                                   | Up About a minute |

You can also use Portainer to check the status of the services. You must run Portainer service first:

```bash
make run-portainer
```

Then, navigate to the following Portainer URL and create an admin account:

<p>
    <a href="http://127.0.0.1:9000" rel="noopener noreferrer" target="_blank"><code>http://127.0.0.1:9000</code></a>
</p>

After, navigate to the following URL to list all of the containers running under the `automated-checkout` stack:

<p>
    <a href="http://127.0.0.1:9000/#/stacks/automated-checkout?type=2&external=true" rel="noopener noreferrer" target="_blank"><code>http://127.0.0.1:9000/#/stacks/automated-checkout?type=2&external=true</code></a>
</p>

### Step 4: Dive in

All of the core components of Automated Checkout are up and running, and you are ready to begin going through the following phases.

- [Phase 1](./phases/phase1.md) - Simulate data and inferencing, and simulate events
- [Phase 2](./phases/phase2.md) - Add Card Reader Device
- [Phase 3](./phases/phase3.md) - Bring Your Own Hardware and Software

## General Guidance

After completing the steps in the Getting Started section, it may be helpful to read through the remainder of this document for some further guidance on the Automated Checkout reference implementation.

### How to use the Compose Files

The `docker-compose.yml` files are segmented to allow for fine control of physical and simulated devices, as well as allowing you the choice of running Portainer. Use the [`makefile`](https://github.com/intel-iot-devkit/automated-checkout/blob/master/Makefile) to manage the various compose files:

| Compose file              | Purpose                                         | Makefile Command                     |
|---------------------------|-------------------------------------------------|--------------------------------------|
| Portainer                 | Container management                            | `make run-portainer`                 |
| All Services              | Automated Checkout and EdgeX services           | `make run`                           |
| Physical Environment      | Mounts physical devices                         | `make run-physical`                  |
| Physical Card Reader      | Allows just the card reader to be physical      | `make run-physical-card-reader`      |
| Physical Controller Board | Allows just the controller board to be physical | `make run-physical-controller-board` |

### Expanding the Reference Implementation

The reference implementation you created is not a complete solution. It provides the base components for creating a framework to run a CV-powered Automated Checkout. It is your choice on how many and which sensors and devices to include. This section provides information about components you might want to include or replace.

| Component                        | Description                                                                                                                                                                                                                                                                                                                                                         |
|----------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| RFID card reader                 | A card reader device service is provided for a USB based RFID card reader. As an alternative, you can also use a regular USB keyboard to enter 10-digit number. See *[Phase 2 - Add Card Reader Device](./phases/phase2.md)* for more information.                                                                                                                  |
| Micro-controller board           | A controller board device service is provided for an Arduino based micro-controller. This micro-controller is in charge of controlling the locks of the automated checkout door and the LED display. Also, it uses modules such as temperature and humidity. See *[Phase 3 - Bring Your Own Hardware and Software](./phases/phase3.md)* for more information.       |
| Deep learning model | The Automated Checkout reference implementation provides a computer vision inference service using openVINO inference engine and openVINO product detection model for demonstration purposes. See more information [here](./automated-checkout-services/device_services.md#cv-inference). You can create your own service and send events to EdgeX using the [EdgeX MQTT Device Service](https://github.com/edgexfoundry/device-mqtt-go). |
