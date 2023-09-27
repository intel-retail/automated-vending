# Configuration

This page lists all of the relevant configuration parameters for each service in the Automated Vending reference implementation.

!!!info
    Note that this document likely does not cover EdgeX-specific configuration parameters. Application and device service SDK documentation can be found in the [EdgeX Foundry GitHub repositories](https://github.com/edgexfoundry) or in the [official EdgeX documentation here](https://docs.edgexfoundry.org/3.0/).

## Environment overrides

The simplest way to change one of the configuration values described below is via the use of environment variable overrides in the docker compose file. The value of each configuration item in a service's configuration can be overridden with an environment variable specific to that item. The name of the environment variable is the path to the item in the configuration tree with underscores separating the nodes. The character case of each node in the environment variable name must match that found in the service's configuration. These overrides are placed in the target service's environment section of the compose file. Here is an example:

```yaml
  ds-card-reader:
    user: "0:0"
    devices:
      - /dev/input:/dev/input
    environment:
      DRIVER_SIMULATEDEVICE: "false"
      DRIVER_DEVICESEARCHPATH: "/dev/input/event*"
      DRIVER_VID: 65535 # 0xFFFF
      DRIVER_PID: 53    # 0x0035
```

## Card reader device service

The following items can be configured via the `DriverConfig` section of the service's [configuration.yaml](https://github.com/intel-retail/automated-vending/blob/Edgex-3.0/ds-card-reader/res/configuration.yaml) file. All values are strings.

- `DeviceSearchPath` - the bash globstar expression to use when searching for the raw input device, default is `/dev/input/event*`
- `VID` - the `uint16` value (as a base-10 string) corresponding to the Vendor ID of the USB device (run `lsusb` to list VID and PID values of connected USB devices). For example, if the VID is `ffff` in the output of `lsusb`, it is `"65535"` in the configuration file
- `PID` - the `uint16` value (as a base-10 string) corresponding to the Product ID of the USB device (run `lsusb` to list VID and PID values of connected USB devices). For example, if the PID is `0035` in the output of `lsusb`, it is `"53"` in the configuration file
- `SimulateDevice` - the boolean value that tells this device service to expect an input device to dictate inputs (`false`), or if a simulated device will be used (and REST API calls will control it) (`true`) - if `true`

## Controller board device service

The following items can be configured via the `DriverConfig` section of the service's [configuration.yaml](https://github.com/intel-retail/automated-vending/blob/Edgex-3.0/ds-controller-board/res/configuration.yaml) file. All values are strings.

- `DisplayTimeout` - The value in seconds corresponding to the display timeout length before resetting the display to the status display.
- `LockTimeout` - The value in seconds corresponding to the lock timeout used to automatically lock the door in case no lock command was sent
- `VID` - the `string` value corresponding to the Vendor ID hexadecimal (base-16) of the USB device (run `lsusb` to list VID and PID values of connected USB devices). For example, if the VID is `2341` in the output of `lsusb`, it is `"2341"` in the configuration file
- `PID` - the `string` value corresponding to the Product ID hexadecimal (base-16) of the USB device (run `lsusb` to list VID and PID values of connected USB devices). For example, if the PID is `8037` in the output of `lsusb`, it is `"8037"` in the configuration file
- `VirtualControllerBoard` - the boolean value that tells this device service to expect an input device to dictate inputs (`false`), or if a simulated device will be used (and REST API calls will control it) (`true`) - if `true`

## EdgeX MQTT device service

This reference implementation uses the [MQTT Device Service](https://github.com/edgexfoundry/device-mqtt-go) from EdgeX with custom device profiles. These device profiles YAML files are located [here](https://github.com/intel-retail/automated-vending/tree/main/res/device-mqtt/profiles/inference-mqtt-device-profile.yml) and are volume mounted into the device service's running Docker container.

The following items can be configured via `device-mqtt.environment` section of the service's [docker-compose.edgex.yml](https://github.com/intel-retail/automated-vending/tree/main/docker-compose.edgex.yml) file.

`device-mqtt.environment`

- `MQTTBROKERINFO_HOST` - Host name of the response MQTT Broker
- `MQTTBROKERINFO_INCOMINGTOPIC` - Subscribe topic for incoming data from MQTT Broker
- `MQTTBROKERINFO_RESPONSETOPIC` - Subscribe topic for the response MQTT Broker

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

The following items can be configured via the `ControllerBoardStatus` section of the service's [configuration.yaml](https://github.com/intel-retail/automated-vending/blob/Edgex-3.0/as-controller-board-status/res/configuration.yaml) file. All values are strings.

- `AverageTemperatureMeasurementDuration` - The time-duration string (i.e. `-15s`, `-10m`) value of how long to process temperature measurements for calculating an average temperature. This calculation determines how quickly a "temperature threshold exceeded" notification is sent
- `DeviceName` - The string name of the upstream EdgeX device that will be pushing events & readings to this application service
- `MaxTemperatureThreshold` - The float64 value of the maximum temperature threshold, if the average temperature over the sample `AverageTemperatureMeasurementDuration` exceeds this value, a notification is sent
- `MinTemperatureThreshold` - The float64 value of the minimum temperature threshold, if the average temperature over the sample `AverageTemperatureMeasurementDuration` exceeds this value, a notification is sent
- `DoorStatusCommandEndpoint` - A string containing the full EdgeX core command REST API endpoint corresponding to the `inferenceDoorStatus` command, registered by the MQTT device service in the cv inference service
- `NotificationCategory` - The category for notifications as a string
- `NotificationEmailAddresses` - A comma-separated values (CSV) string of emails to send notifications to
- `NotificationLabels` - A comma-separated values (CSV) string of labels to apply to notifications, which are handled by EdgeX
- `NotificationReceiver` - The human-readable string name of the person/entity receiving the notification, such as `System Administrator`
- `NotificationSender` - The human-readable string name of the person/entity sending the notification, such as `Automated Vending Maintenance Notification`
- `NotificationSeverity` - A string tag indicating the severity of the notification, such as `CRITICAL`
- `NotificationName` - A string that is a short label that may be used as part of a URL to delineate the notification subscription, such as `sys-admin`. The EdgeX official documentation says, _"Effectively a name or key that labels the notification"_.
- `NotificationSubscriptionMaxRESTRetries` - The integer value that represents the maximum number of times to try creating a subscription in the EdgeX notification service, such as `10`
- `NotificationSubscriptionRESTRetryIntervalDuration` - The time-duration string (i.e. `10s`) representing how long to wait between each attempt at trying to create a subscription in the EdgeX notification service,
- `NotificationThrottleDuration` - The time-duration string corresponding to how long to snooze notification alerts after sending an alert, such as `1m`. Note that this value is stored in memory at runtime and if the service restarts, the time between notifications is not kept.
- `RESTCommandTimeoutDuration` - The time-duration string representing how long to wait for any command to an EdgeX command API response before considering it a timed-out request, such as `15s`
- `SubscriptionAdminState` - The URL (as a string) of the EdgeX notification service's subscription API
- `VendingEndpoint` - The URL (as a string) corresponding to the central vending endpoint's `/boardStatus` API endpoint, which is where events will be Posted when there is a door open/close change event, or a "temperature threshold exceeded" event.

## Vending application service

The following items can be configured via the `ApplicationSettings` section of the service's [configuration.yaml](https://github.com/intel-retail/automated-vending/blob/Edgex-3.0/as-vending/res/configuration.yaml) file. All values are strings.

- `AuthenticationEndpoint` - Endpoint for authentication microservice
- `ControllerBoarddisplayResetCmd` - EdgeX Command service command for Resetting the LCD text
- `ControllerBoarddisplayRow0Cmd` - EdgeX Command service command for Row 0 on LCD
- `ControllerBoarddisplayRow1Cmd` - EdgeX Command service command for Row 1 on LCD
- `ControllerBoarddisplayRow2Cmd` - EdgeX Command service command for Row 2 on LCD
- `ControllerBoarddisplayRow3Cmd` - EdgeX Command service command for Row 3 on LCD
- `ControllerBoardLock1Cmd` - EdgeX Command service command for lock 1 events
- `ControllerBoardLock2Cmd` - EdgeX Command service command for lock 2 events
- `CardReaderDeviceName` - String value, a Card reader device name. Incoming events/readings that do not match this card reader device name will likely be ignored by this service.
- `InferenceDeviceName` - String value, a Inference device name. Incoming events/readings that do not match this device name will likely be ignored by this service.
- `ControllerBoardDeviceName` - String value, a Controller board device name. Incoming events/readings that do not match this device name will likely be ignored by this service.
- `DoorCloseStateTimeoutDuration` - The time-duration string (i.e. `-15s`, `-10m`) used for Door Close lockout time delay, in seconds
- `DoorOpenStateTimeoutDuration` - The time-duration string (i.e. `-15s`, `-10m`) used for Door Open lockout time delay, in seconds
- `InferenceDoorStatusCmd` - EdgeX Command service command for Inference Door status
- `InferenceHeartbeatCmd` - EdgeX Command service command for Inference Heartbeat
- `InferenceTimeoutDuration` - The time-duration string (i.e. `-15s`, `-10m`) used for Inference message time delay, in seconds
- `InventoryAuditLogService` - Endpoint for Inventory Audit Log Micro Service
- `InventoryService` - Endpoint for Inventory Micro Service
- `LCDRowLength` - Max number of characters for LCD Rows
- `LedgerService` - Endpoint for Ledger Micro Service

## Authentication microservice

For this particular microservice, there are no specific configuration options. Future settings would be added under the `[ApplicationSettings]` section.

## Inventory microservice

For this particular microservice, there are no specific configuration options. Future settings would be added under the `[ApplicationSettings]` section.

## Ledger microservice

The following items can be configured via the `[ApplicationSettings]` section of the service's [configuration.yaml](https://github.com/intel-retail/automated-vending/blob/Edgex-3.0/ms-ledger/res/configuration.yaml) file. All values are strings.

- `InventoryEndpoint` - Endpoint that correlates to the Inventory microservice. This is used to query Inventory data used to generate the ledgers.
