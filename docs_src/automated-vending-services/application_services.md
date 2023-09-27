
# Application Services

The **Automated Vending** reference implementation utilizes two application services that are used to consume event data from the EdgeX framework.

## List of application services

- Controller Board Status – Handles events coming from the controller board device service.
- Vending – The main business logic for the Automated Vending application. This service handles events directly from the card reader device service and inference engine as well as coordinates data between each of the microservices.

## Controller board status application service

### Controller board status application service description

The `as-controller-board-status` application service checks the status of the controller board for changes in the state of the door, lock, temperature, and humidity, and triggers notifications if the average temperature and humidity are outside the desired ranges.

### Controller board status application service APIs

This service exposes a few REST API endpoints that are either intended to be interacted with via EdgeX's core services or directly.

All exposed HTTP responses are of the format:

```json
{
    "content": "",
    "contentType": "json|string|<any>",
    "statusCode": 200,
    "error": false
}
```

This means that you must parse the `content` field if the `contentType` field is `json`.

This section includes documentation on each API endpoint exposed by this service. Depending on how you've chosen to deploy the service, you may use `localhost`, `as-controller-board-status` or some other hostname. This document assumes `localhost` is used.

---

#### `GET`: `/status`

The `GET` call will return the current controller board status information. This information can be used by other services trying to get metrics from the controller board or by a test suite to validate the controller board behavior.

Simple usage example:

```bash
curl -X GET http://localhost:48094/status
```

Sample response:

```json
{
    "content": "{\"lock1_status\": 0,\"lock2_status\": 0,\"door_closed\": false,\"temperature\": 30.1,\"humidity\": 26.2,\"minTemperatureStatus\": true,\"maxTemperatureStatus\": false}",
    "contentType": "json",
    "statusCode": 200,
    "error": false
}
```

If there is an error marshaling the controller board's state into a JSON response, the error will be an HTTP 500 internal server error in the expected format:

```json
{
    "content": "Failed to serialize the controller board's current state.",
    "contentType": "string",
    "statusCode": 500,
    "error": true
}
```

---

## Vending application service

### Vending application service description

The `as-vending` application service is the central microservice that contains the business logic to handle the following:

- Coordinates unlocking the cooler upon authentication
- Requests inference snap shots (an inventory delta since the cooler was last closed)
- Updates the inventory and ledger
- Displays transaction data to the LCD

This service also implements **_"maintenance mode"_** to manage error handling and recovery due to faulty hardware, temperatures outside the desired ranges, or any other actions that disrupt the normal workflow of the vending machine. The functions that execute this logic can be found in `as-vending/functions/output.go`

### Vending application service APIs

---

#### `POST`: `/boardStatus`

The `POST` call will inform the application service on the current state of the instrumentation (temperature, door, lock, humidity) on the controller board so that it can handle the business logic associated with those states.  The events are typically posted from the controller board status application service. It is important to highlight that the REST API response will not necessarily be a holistic response of all of the actions taken place by the `as-vending` service. Please review the service's logs in order to gain a complete view of all changes that occur when interacting with this API endpoint.

Simple usage example:

```bash
curl -X POST -d '{"lock1_status": 0,"lock2_status": 0,"door_closed": false,"temperature": 30.1,"humidity": 26.2,"minTemperatureStatus": true,"maxTemperatureStatus": false}' http://localhost:48099/boardStatus
```

A `POST` without any new information in the body will return the result:

!!! success
    Response Status Code 200 OK.
    Board status was read

If the `minTemperatureStatus` or `maxTemperatureStatus` values are set to `true`, maintenance mode will be set and the HTTP API response may be:

!!! success
    Response Status Code 200 OK.
    Temperature status received and maintenance mode was set

If the `door_closed` property is different than what `as-vending` currently believes it is, this response may be returned:

!!! success
    Response Status Code 200 OK.
    Door closed change event was received

---

### `POST`: `/resetDoorLock`

The `POST` call will simply reset the Automated Vending's internal `vendingState`. This API endpoint has no logic to process any input data - it just responds to a simple `POST`.

Simple usage example:

```bash
curl -X POST http://localhost:48099/resetDoorLock
```

The response will _always_ be `200 OK`:

```bash
reset the door lock
```

---

### `GET`: `/maintenanceMode`

The `GET` call will simply return the boolean state that represents whether or not the vending state is in maintenance mode.

Simple usage example:

```bash
curl -X GET http://localhost:48099/maintenanceMode
```

The response will _always_ be `200 OK`:

```json
{
    "content": "{\"maintenanceMode\":false}",
    "contentType": "json",
    "statusCode": 200,
    "error": false
}
```
