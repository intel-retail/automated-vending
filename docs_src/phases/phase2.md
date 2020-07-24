# Phase 2 - Add Card Reader Device

In [phase 1](./phase1.md), the scenarios presented a breakdown of the various modes, events, and services that are working together within the Automated Checkout reference design. Everything in phase 1 was simulated and all interactions were done via REST API calls.

Phase 2 will be mostly the same, except there will now be a physical card reader device. This device is actually just a keyboard that types 10 digits and then presses enter, which is what a common RFID card reader also might do.

## Setup

If the Automated Checkout services are still running from phase 1, the services can stay running. The only service that will be terminated and re-created is `ds-card-reader` (which will be done as part of this guide).

Start by removing the `ds-card-reader` service. In order to do this, first identify the container's name using this command:

```bash
docker ps | grep -i ds-card-reader
```

Use the output of that command to stop the container for `ds-card-reader` - replace `container_name_or_id_from_previous_command` with the output from above:

```bash
docker rm -f container_name_or_id_from_previous_command
```

!!! note
    If you encounter any issues with the `ds-card-reader` later in this guide, consider cleaning up all of the Automated Checkout services. The best way to guarantee a clean run through of this phase is to tear down any existing Automated Checkout environment and start from fresh. This can be accomplished by running the following steps.

    Navigate to the root of this repository - this can vary based on where you chose to clone the repository:

    ```
    cd <repository_root>
    ```

    Run the following commands to clean things up:

    ```
    make down && make clean-docker
    ```

    This will destroy any existing ledger entries, audit log entries, inventory changes, and EdgeX event readings and data associated with Automated Checkout. However, this will not alter any other non-Automated Checkout Docker images, containers, or volumes. The scope of the above command is limited to only Automated Checkout.

## Plug in the card reader device

The first step is to simply plug in the card reader device. This can be a regular USB keyboard or a dedicated RFID card reader, or any other HID keyboard-like input device that works with [`evdev`](https://en.wikipedia.org/wiki/Evdev) and can be identified via the Linux command `lsusb`.

!!! note
    If you are going to use a keyboard as a card reader device, we suggest plugging in a second keyboard for this purpose.

Once it's plugged in, proceed to identify the device by running the command:

```bash
lsusb
```

The output may look like this (this is the output from a virtual machine):

```text
Bus 002 Device 001: ID 1d6b:0003 Linux Foundation 3.0 root hub
Bus 001 Device 003: ID ffff:0035
Bus 001 Device 002: ID 80ee:0021 VirtualBox USB Tablet
Bus 001 Device 001: ID 1d6b:0002 Linux Foundation 2.0 root hub
```

The particular _vendor ID_ and _product ID_ are spelled out clearly for each USB device. The card reader input device itself has been plugged in and has the vendor ID `ffff` and the product ID `0035`.

!!! note
    The VID and PID values are hexadecimal, base 16. A value of `ffff` is equal to `65535` in decimal, base 10, and `0035` in base 16 is equal to `53` in base 10. The configuration files in the Automated Checkout reference design device services may require some conversion between the two. If needed, consider searching online for a hexadecimal to decimal conversion calculator to make the process easier.

Once the VID and PID have been identified, the next step is to configure the `ds-card-reader` device service to grab that device and listen for input events.

## Configure the `ds-card-reader` service to use the card reader device

From the root of this repository, with the card reader device plugged in and its VID and PID identified, navigate to the `ds-card-reader/res/docker` directory:

```bash
cd <repository_root>
```

Then, modify the `docker-compose.physical.card-reader.yml` file in your text editor of choice:

```bash
nano docker-compose.physical.card-reader.yml
```

In this file, you'll see a section that looks like this:

```yaml
ds-card-reader:
  user: "0:0"
  devices:
    - /dev/input:/dev/input
  environment:
    Driver_SimulateDevice: "false"
```

We will be adding three environment variables to this service:

- `DeviceSearchPath` - the path to search for `evdev` input devices
- `VID` - the base-10 value of the vendor ID associated with the input device
- `PID` - the base-10 value of the product ID associated with the input device

First, verify that the default value of `DeviceSearchPath="/dev/input/event*"` corresponds to an actual path on your Linux system - the vast majority of Linux systems should automatically handle everything in this directory, but it helps to check.

```bash
ls -al /dev/input/event
```

<details>
  <summary><i>(Click to Expand)</i> Example Output</summary>

<p>
The output of <code>ls -al /dev/input/event</code> may look like this:
</p>

```text
crw-rw---- 1 root input 13, 64 Apr  9 09:10 /dev/input/event0
crw-rw---- 1 root input 13, 65 Apr  9 09:10 /dev/input/event1
crw-rw---- 1 root input 13, 66 Apr  9 09:10 /dev/input/event2
crw-rw---- 1 root input 13, 67 Apr  9 09:10 /dev/input/event3
crw-rw---- 1 root input 13, 68 Apr  9 09:10 /dev/input/event4
crw-rw---- 1 root input 13, 69 Apr  9 09:10 /dev/input/event5
crw-rw---- 1 root input 13, 70 Apr  9 09:10 /dev/input/event6
crw-rw---- 1 root input 13, 71 Apr  9 09:58 /dev/input/event7
```

<p>
If you do not see input devices under this path, your Linux kernel or operating system may be configured differently. Consult your operating system's documentation for information regarding the behavior of input devices if possible.
</p>

</details>

The resulting section in the configuration file will look something like this:

```yaml
ds-card-reader:
  user: "0:0"
  devices:
    - /dev/input:/dev/input
  environment:
    Driver_SimulateDevice: "false"
    Driver_DeviceSearchPath: "/dev/input/event*"
    Driver_VID: "65535"
    Driver_PID: "53"
```

## Run the Automated Checkout reference design

Run the Automated Checkout reference design with the physical card reader component included:

```bash
make run-physical-card-reader-dev
```

After about a minute or so, the card reader device service (ds-card-reader) will be configured to accept inputs from an external card reader device. Follow the steps outlined in [phase 1](./phase1.md#walkthrough-of-scenarios) again, **except** instead of performing REST API calls to simulate badge swipe events, replace them with a keyboard inputs that correspond to the same cards ID, press `enter`, and then continue forward with the REST API calls that simulate door open/closure events, temperature changes, etc.

!!! note
    You do not need to type the card ID number anywhere specifically. The card reader device service is configured in such a way that it will listen to any inputs from the keyboard at any time.

    The input device is globally grabbed using [evdev's `GrabDevice`](https://www.x.org/releases/X11R7.6/doc/man/man4/evdev.4.xhtml) mechanism, summarized here:

    > **GrabDevice**: ... Doing so will ensure that no other driver can initialise the same device and it will also stop the device from sending events to /dev/kbd or /dev/input/mice. Events from this device will not be sent to virtual devices (e.g. rfkill or the Macintosh mouse button emulation).

For example, in [phase 1](./phase1.md#walkthrough-of-scenarios), the card number for the stocker role is `0003293374`. This card number can be simulated by typing the digits and pressing the enter key, if you're using a keyboard as your input device.

## Dive deeper

Now that the card reader is working with a physical device, it may be time to make some changes to the underlying authentication data to allow your own cards to authenticate. The following sections illustrate the steps needed in order to extend the Automated Checkout reference design to work with your cards.

### Extending the card reader

If the behavior of your particular card reader device's cards does not match the behavior that's been incorporated into this service, you'll need to get your hands dirty with the source code of the `ds-card-reader` device service.

In the service, take a look at the file `ds-card-reader/device/physical.go`. This file contains this function:

```go
func (reader *CardReaderPhysical) Listen() {
    // ...
    reader.processDevReadEvents(events)
    // ...
}
```

`Listen` is a Go routine that loops and processes `evdev` events as they come in from the physical input device. Carefully inspect this section of code as well as the functions called within this Go routine in order to gain an understanding of how to change the behavior of the card reader device service.

!!! warning
    Changing the source code may break unit tests and other functionality across services. Ensure that the software development processes used to make code changes include updating unit and integration tests to work with new changes.

### Adding new cards

The `ms-authentication` microservice contains an index of all cards, accounts, and authorized individuals (people). To add a new card, person, or account, follow these steps.

First, navigate to the `ms-authentication` directory in the root of the repository. These three `.json` files dictate the `ms-authentication` service's behavior:

- [`people.json`](https://github.com/intel-iot-devkit/automated-checkout/blob/master/ms-authentication/people.json)
- [`accounts.json`](https://github.com/intel-iot-devkit/automated-checkout/blob/master/ms-authentication/accounts.json)
- [`cards.json`](https://github.com/intel-iot-devkit/automated-checkout/blob/master/ms-authentication/cards.json)

In this case, we're only going to add a new card and associate it with the person with ID 1. A typical card will look like this:

```json
{
    "cardId": "0003621892",
    "roleId": 1,
    "isValid": true,
    "personId": 3,
    "createdAt": "1560815799",
    "updatedAt": "1560815799"
}
```

Let's add a new card to the `cards.json` file - replace `1234567899` with a 10-digit card that corresponds to one of your cards:

```json
...,
{
    "cardId": "1234567899",
    "roleId": 1,
    "isValid": true,
    "personId": 3,
    "createdAt": "1560815799",
    "updatedAt": "1560815799"
}
...
```

!!! info
    The `createdAt` and `updatedAt` dates do not particularly matter, but should be kept as a Unix timestamp.

Now that we've added the card to the `cards.json` file, the service's Docker image must be rebuilt. Navigate to the root of this repository, which should be up a single directory:

```bash
cd ..
```

Then, use the `Makefile` to build the `ds-card-reader` image:

```bash
make ds-card-reader
```

!!! info
    At this time, the three `.json` files are built in to the Docker image for the `ms-authentication` service. They are not mounted at runtime. This is why image rebuilding is necessary.

### Running the updated service

If the Automated Checkout services are already running, the best way to update the running `ms-authentication` service is to remove the `ms-authentication` container and then re-run the command to bring up the whole stack.

First, remove the `ms-authentication` container:

```bash
docker ps -a | grep -i ms-authentication
```

Use the output of the last command to delete the `ms-authentication` container. Replace `container_name` in the below command with the name from the output from the above command to delete it:

```bash
docker rm -f container_name
```

Then, bring up the services using the same command from before:

```bash
make run-physical-card-reader-dev
```

!!! note
    If this fails to properly update the image, it may be worth running

    ```
    make down
    ```

    And then re-running

    ```
    make run-physical-card-reader-dev
    ```

    If there are still issues, consider completely cleaning the Automated Checkout containers and volumes by running

    ```
    make down && make clean-docker
    ```

    And then running

    ```
    make run-physical-card-reader-dev
    ```

The `ds-card-reader` service should be listening for input events. If your card reader device is a proper RFID USB card reader, swipe the card that corresponds to the card we added, or if it's a USB keyboard, type out the keys and press enter when done, and follow the steps in phase 1 while replacing card reader badge-in events with this method.

!!! info
    For more generalized information on modifying source code, please review the [Modifying source code](../modifying_source_code.md) page.

## Summary

The usage of a physical card reader device only requires a few changes from the simulated mode. In the `ds-card-reader` device service, the device's VID and PID are configured, the service's image gets rebuilt, and the service itself gets updated to use the new image. The device's interactions are captured by Go routines running in the device service itself, and EdgeX event readings are propagated throughout a handful of services to ensure a smooth Automated Checkout workflow.
