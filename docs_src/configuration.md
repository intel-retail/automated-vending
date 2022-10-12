# Configuration

This page lists all of the relevant configuration parameters for each service in the Automated Checkout reference implementation.

!!!info
    Note that this document likely does not cover EdgeX-specific configuration parameters. Application and device service SDK documentation can be found in the [EdgeX Foundry GitHub repositories](https://github.com/edgexfoundry) or in the [official EdgeX documentation here](https://docs.edgexfoundry.org/1.2/).

## Environment overrides

The simplest way to change one of the configuration values described below is via the use of environment variable overrides in the docker compose file. The value of each configuration item in a service's configuration can be overridden with an environment variable specific to that item. The name of the environment variable is the path to the item in the configuration tree with underscores separating the nodes. The character case of each node in the environment variable name must match that found in the service's configuration. Here are a few examples for the `Driver` section:

```toml
[Driver]
  VID = "65535" # 0xFFFF
  PID = "53"    # 0x0035
```

```yaml
Driver_VID: "256" ** Good **
Driver_PID: "26"  ** Good **

DRIVER_VID: "256" ** BAD **
driver_pid: "26"  ** BAD **
```

These overrides are placed in the target service's environment section of the compose file. Here is an example:

```yaml
  ds-card-reader:
    user: "2000:2000"
    image: "automated-checkout/ds-card-reader:dev"
    container_name: automated-checkout_ds-card-reader
    environment:
      <<: *common-variables
      Driver_VID: "256"
      Driver_PID: "26"
    <<: *logging
    depends_on:
      - data
      - command
    restart: always
    ipc: none
```

## Card reader device service

The following items can be configured via the `Driver` section of the service's [configuration.toml](https://github.com/intel-iot-devkit/automated-checkout/blob/master/ds-card-reader/res/configuration.toml) file. All values are strings.

- `DeviceName` - the name of the device to be associated with EdgeX events and readings originating from this service, if unsure leave default `ds-card-reader`
- `DeviceSearchPath` - the bash globstar expression to use when searching for the raw input device, default is `/dev/input/event*`
- `VID` - the `uint16` value (as a base-10 string) corresponding to the Vendor ID of the USB device (run `lsusb` to list VID and PID values of connected USB devices). For example, if the VID is `ffff` in the output of `lsusb`, it is `"65535"` in the configuration file
- `PID` - the `uint16` value (as a base-10 string) corresponding to the Product ID of the USB device (run `lsusb` to list VID and PID values of connected USB devices). For example, if the PID is `0035` in the output of `lsusb`, it is `"53"` in the configuration file
- `SimulateDevice` - the boolean value that tells this device service to expect an input device to dictate inputs (`false`), or if a simulated device will be used (and REST API calls will control it) (`true`) - if `true`

## Controller board device service

The following items can be configured via the `Driver` section of the service's [configuration.toml](https://github.com/intel-iot-devkit/automated-checkout/blob/master/ds-controller-board/res/configuration.toml) file. All values are strings.

- `DisplayTimeout` - The value in seconds corresponding to the display timeout length before resetting the display to the status display.
- `LockTimeout` - The value in seconds corresponding to the lock timeout used to automatically lock the door in case no lock command was sent
- `VID` - the `string` value corresponding to the Vendor ID hexadecimal (base-16) of the USB device (run `lsusb` to list VID and PID values of connected USB devices). For example, if the VID is `2341` in the output of `lsusb`, it is `"2341"` in the configuration file
- `PID` - the `string` value corresponding to the Product ID hexadecimal (base-16) of the USB device (run `lsusb` to list VID and PID values of connected USB devices). For example, if the PID is `8037` in the output of `lsusb`, it is `"8037"` in the configuration file
- `VirtualControllerBoard` - the boolean value that tells this device service to expect an input device to dictate inputs (`false`), or if a simulated device will be used (and REST API calls will control it) (`true`) - if `true`

## EdgeX MQTT device service

This reference implementation uses the [MQTT Device Service](https://github.com/edgexfoundry/device-mqtt-go) from EdgeX with custom device profiles. These device profiles YAML files are located [here](https://github.com/intel-iot-devkit/automated-checkout/blob/master/res/device-mqtt/docker) and are volume mounted into the device service's running Docker container.

The following items can be configured via the `DeviceList` and `Driver` section of the service's [configuration.toml](https://github.com/intel-iot-devkit/automated-checkout/blob/master/res/device-mqtt/docker/configuration.toml) file. All values are strings.

`DeviceList`

- `IncomingSchema` - Data schema type, aka protocol
- `IncomingHost` - Host name of the incoming MQTT Broker
- `IncomingPort` - Port number of the incoming MQTT Broker
- `IncomingUser` - Username for the incoming MQTT Broker
- `IncomingPassword` - Password for the incoming MQTT Broker
- `IncomingQos` - Quality of service agreement between sender and receiver
- `IncomingKeepAlive` - Keep alive duration for the incoming MQTT Broker
- `IncomingClientId` - Client ID for the incoming MQTT Broker
- `IncomingTopic` - Subscribe topic for the incoming MQTT Broker

`Driver`

- `ResponseSchema` - Data schema type, aka protocol
- `ResponseHost` - Host name of the response MQTT Broker
- `ResponsePort` - Port number of the response MQTT Broker
- `ResponseUser` - Username for the response MQTT Broker
- `ResponsePassword` - Password for the response MQTT Broker
- `IncomingQos` - Quality of service agreement between sender and receiver
- `ResponseKeepAlive` - Keep alive duration for the response MQTT Broker
- `ResponseClientId` - Client ID for the response MQTT Broker
- `ResponseTopic` - Subscribe topic for the response MQTT Broker

## CV inference device service

If you run the CV inference device service via source code, the following items can be configured via arguments:

```text
  -confidence float
        Confidence threshold. (default 0.85)
  -config string
        XML model config file path. (default "product-detection-0001/FP32/product-detection-0001.xml")
  -dir string
        Images directory. (default "./images")
  -model string
        Model file path. (default "product-detection-0001/FP32/product-detection-0001.bin")
  -mqtt string
        Mqtt address. (default "localhost:1883")
  -skuMapping string
        SKU Mapping JSON file path (default "skumapping.json")
```

For simplification, if you use the docker version, we have created an entrypoint script that has some defaults for model and config. For the rest of the arguments, they must be passed in the following order:

`command: ["/path/to/images","mqtt URL","confidence value","/path/to/skumapping.json"]`

Example:

`command: ["/go/src/ds-cv-inference/images","mqtt-broker:1883","0.85","/go/src/ds-cv-inference/skumapping.json"]`

## Controller board status application service

The following items can be configured via the `ApplicationSettings` section of the service's [configuration.toml](https://github.com/intel-iot-devkit/automated-checkout/blob/master/as-controller-board-status/res/configuration.toml) file. All values are strings.

- `AverageTemperatureMeasurementDuration` - The time-duration string (i.e. `-15s`, `-10m`) value of how long to process temperature measurements for calculating an average temperature. This calculation determines how quickly a "temperature threshold exceeded" notification is sent
- `DeviceName` - The string name of the upstream EdgeX device that will be pushing events & readings to this application service
- `MaxTemperatureThreshold` - The float64 value of the maximum temperature threshold, if the average temperature over the sample `AverageTemperatureMeasurementDuration` exceeds this value, a notification is sent
- `MinTemperatureThreshold` - The float64 value of the minimum temperature threshold, if the average temperature over the sample `AverageTemperatureMeasurementDuration` exceeds this value, a notification is sent
- `DoorStatusCommandEndpoint` - A string containing the full EdgeX core command REST API endpoint corresponding to the `inferenceDoorStatus` command, registered by the MQTT device service in the cv inference service
- `NotificationCategory` - The category for notifications as a string
- `NotificationEmailAddresses` - A comma-separated values (CSV) string of emails to send notifications to
- `NotificationHost` - The full string URL of the EdgeX notifications service API that allows notifications to be sent by submitting an HTTP Post request
- `NotificationLabels` - A comma-separated values (CSV) string of labels to apply to notifications, which are handled by EdgeX
- `NotificationReceiver` - The human-readable string name of the person/entity receiving the notification, such as `System Administrator`
- `NotificationSender` - The human-readable string name of the person/entity sending the notification, such as `Automated Checkout Maintenance Notification`
- `NotificationSeverity` - A string tag indicating the severity of the notification, such as `CRITICAL`
- `NotificationSlug` - A string that is a short label that may be used as part of a URL to delineate the notification subscription, such as `sys-admin`. The EdgeX official documentation says, _"Effectively a name or key that labels the notification"_. This service creates an EdgeX subscription with a `slug` value of `NotificationSlug`.
- `NotificationSlugPrefix` - A string similar to `NotificationSlug`, except that the `NotificationSlugPrefix` is appended with the current system time and the actual notification message's slug value is set to that value.
- `NotificationSubscriptionMaxRESTRetries` - The integer value that represents the maximum number of times to try creating a subscription in the EdgeX notification service, such as `10`
- `NotificationSubscriptionRESTRetryInterval` - The time-duration string (i.e. `10s`) representing how long to wait between each attempt at trying to create a subscription in the EdgeX notification service,
- `NotificationThrottleDuration` - The time-duration string corresponding to how long to snooze notification alerts after sending an alert, such as `1m`. Note that this value is stored in memory at runtime and if the service restarts, the time between notifications is not kept.
- `RESTCommandTimeout` - The time-duration string representing how long to wait for any command to an EdgeX command API response before considering it a timed-out request, such as `15s`
- `SubscriptionHost` - The URL (as a string) of the EdgeX notification service's subscription API
- `VendingEndpoint` - The URL (as a string) corresponding to the central vending endpoint's `/boardStatus` API endpoint, which is where events will be Posted when there is a door open/close change event, or a "temperature threshold exceeded" event.

## Vending application service

The following items can be configured via the `ApplicationSettings` section of the service's [configuration.toml](https://github.com/intel-iot-devkit/automated-checkout/blob/master/as-vending/res/configuration.toml) file. All values are strings.

- `AuthenticationEndpoint` - Endpoint for authentication microservice
- `DeviceControllerBoarddisplayReset` - EdgeX Command service endpoint for Resetting the LCD text
- `DeviceControllerBoarddisplayRow0` - EdgeX Command service endpoint for Row 0 on LCD
- `DeviceControllerBoarddisplayRow1` - EdgeX Command service endpoint for Row 1 on LCD
- `DeviceControllerBoarddisplayRow2` - EdgeX Command service endpoint for Row 2 on LCD
- `DeviceControllerBoarddisplayRow3` - EdgeX Command service endpoint for Row 3 on LCD
- `DeviceControllerBoardLock1` - EdgeX Command service endpoint for lock 1 events
- `DeviceControllerBoardLock2` - EdgeX Command service endpoint for lock 2 events
- `DeviceNames` - String value, containing comma-separated device names, as registered in EdgeX for each service. Incoming events/readings that do not match one of the device names will likely be ignored by this service.
- `DoorCloseStateTimeout` - The time-duration string (i.e. `-15s`, `-10m`) used for Door Close lockout time delay, in seconds
- `DoorOpenStateTimeout` - The time-duration string (i.e. `-15s`, `-10m`) used for Door Open lockout time delay, in seconds
- `InferenceDoorStatus` - EdgeX Command service endpoint for Inference Door status
- `InferenceHeartbeat` - EdgeX Command service endpoint for Inference Heartbeat
- `InferenceTimeout` - The time-duration string (i.e. `-15s`, `-10m`) used for Inference message time delay, in seconds
- `InventoryAuditLogService` - Endpoint for Inventory Audit Log Micro Service
- `InventoryService` - Endpoint for Inventory Micro Service
- `LCDRowLength` - Max number of characters for LCD Rows
- `LedgerService` - Endpoint for Ledger Micro Service

## Authentication microservice

For this particular microservice, there are no specific configuration options. Future settings would be added under the `[ApplicationSettings]` section.

## Inventory microservice

For this particular microservice, there are no specific configuration options. Future settings would be added under the `[ApplicationSettings]` section.

## Ledger microservice

The following items can be configured via the `ApplicationSettings` section of the service's [configuration.toml](https://github.com/intel-iot-devkit/automated-checkout/blob/master/ms-ledger/res/configuration.toml) file. All values are strings.

- `InventoryEndpoint` - Endpoint that correlates to the Inventory microservice. This is used to query Inventory data used to generate the ledgers.
