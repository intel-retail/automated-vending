# Automated Checkout Events

The following are the different events used in the Automated Checkout reference implementation.

## Card reader events

[`ds-card-reader`](./automated-checkout-services/device_services.md#card-reader) uses the EdgeX events pattern to send the card information into EdgeX Core Data as well as maintain a healthy state.

`card-number` is an asynchronous event sent when a card is scanned by the card reader device. The event reading value is a string containing a 10-digit number and is placed into EdgeX Core Data as an event reading.

``` json
"0003293374"
```

`status` is an auto-event used to check the status of card reader connection. Every x seconds the event will check to see if the card reader is accessible. If the service is unable to verify the connection to the card reader then the service will restart. This event produces no database entry in EdgeX Core Data.

## Controller board events

[`ds-controller-board`](./automated-checkout-services/device_services.md#controller-board) uses the EdgeX events pattern to send the card information into EdgeX Core Data.

`controller-board-status` is an auto-event used to send the current state of the controller board and all of its periferals to EdgeX Core Data. This data is used by the as-controller-board-status application service to determine the state of the system. The information included in the status are the door lock states, door state, temperature, and humidity. The EdgeX Reading value is a string containing the following JSON:

``` json
{
    "lock1_status":1,
    "lock2_status":1,
    "door_closed":true,
    "temperature":78,
    "humidity":10
}
```

The following commands are also available to the ds-controller-board. More details can be found [`here`](./automated-checkout-services/device_services.md#controller-board).

- `lock1`
- `lock2`
- `displayRow0`
- `displayRow1`
- `displayRow2`
- `displayRow3`
- `setTemperature`
- `setHumidity`
- `setDoorClosed`

## Computer vision inference events

The [`ds-cv-inference`](./automated-checkout-services/device_services.md#cv-inference) uses the EdgeX MQTT Device Service to send status updates and inference deltas to the EdgeX Core Data. The MQTT device service profile can be found in the file `./res/device-mqtt/inference.mqtt.device.profile.yml`, in the root of this GitHub repository.

`inferenceHeartbeat` is an asynchronous event that is pushed to EdgeX Core Data when the inference is pinged by another service to verify it is functioning. If the inference is working properly the `inferencePong` response is sent as the event reading.

``` json
{
    "inferencePong"
}
```

`inferenceSkuDelta` is an asynchronous event that pushes the delta data from the inference engine into EdgeX Core Data. The delta data can be used to update the inventory and create ledgers when appropriate. The EdgeX Event Reading contains a string value which is represented by the following JSON example:

``` json
[
    {"SKU": "4900002470", "delta": -1},
    {"SKU": "1200010735", "delta": -2}
]
```

Finally the `inferenceDoorStatus` command is defined by the custom device profile for the EdgeX MQTT Device Service which sends the ping request to the CV inference service. More details can be found [here](./automated-checkout-services/device_services.md#cv-inference).
