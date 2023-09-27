# Other Microservices

The Automated Vending reference implementation utilizes three services that expose REST API endpoints. These three services handle business logic for the Automated Vending reference implementation, and are somewhat generic in their design patterns, so for the purposes of the reference implementation, we simply refer to them  "microservices".

## List of microservices

- [Authentication](#authentication-service) - Service that takes a card ID number and returns authentication/authorization status for the card number.
- [Inventory](#inventory-service) - Service that manages changes to the Automated Vending's inventory, including storing transactions in an audit log.
- [Ledger](#ledger-service) - Service that stores customer financial transactions.

## Authentication service

### Authentication service description

The `ms-authentication` microservice is a service that works with EdgeX to expose a REST API that takes a simple 10-digit string of digits (presumably corresponding to an RFID card) and responds with a valid or invalid authentication response, as well as the corresponding role for authenticated cards.

This repository contains logic for working within the following schemas:

- _Card/Cards_ - swiping a card is what allows the Automated Vending automation to proceed with its workflow. A card can be associated with one of 3 roles:
  - Consumer - a typical customer; is expected to open the vending machine door, remove an item, close the door and be charged accordingly
  - Stocker - a person that is authorized to re-stock the vending machine with new products
  - Maintainer - a person that is authorized to fix the software/hardware
- _Account/Accounts_ - represents a bank account to charge. Multiple people can be associated with an account, such as a married couple
- _Person/People_ - a person can carry multiple cards but is only associated with one account

The [`ds-card-reader`](https://github.com/intel-retail/automated-vending/tree/main/ds-card-reader) service is responsible for pushing card "swipe" events to the EdgeX framework, which will then feed into the [`as-vending`](https://github.com/intel-retail/automated-vending/tree/main/as-vending) microservice that then performs a REST HTTP API call to this microservice. The response is processed by the [`as-vending`](https://github.com/intel-retail/automated-vending/tree/main/as-vending) microservice and the workflow continues there.

### Authentication service APIs

---

#### `GET`: `/authentication/{cardid}`

The `GET` call will return the user information if the `cardid` URL parameter matches a valid card ID number (according to the file `cards.json`). If the `cardid` is not found, an unauthorized response is returned.

Simple usage example:

```bash
curl -X GET http://localhost:48096/authentication/0003278425
```

Authorized card sample response:

```json
{
    "content": "{\"accountID\":1,\"personID\":1,\"roleID\":1,\"cardID\":\"0003278425\"}",
    "contentType": "json",
    "statusCode": 200,
    "error": false
}
```

Unauthorized card sample response:

```json
  {
      "content": "Card ID is not a valid card",
      "contentType": "string",
      "statusCode": 401,
      "error": false
  }
```

## Inventory service

### Inventory service description

The `ms-inventory` microservice is a service that works with EdgeX to expose a REST API that manages the inventory for the vending machine, and keeps an audit log of all transactions (authorized or not).

This repository contains logic for working within the following schemas:

- _Inventory_ - an inventory item has the following attributes:
  - `sku` - the SKU number of the inventory item
  - `itemPrice` - the price of the inventory item
  - `productName` - the name of the inventory item, will be displayed to users
  - `unitsOnHand` - the number of units stored in the vending machine
  - `maxRestockingLevel` - the maximum allowable number of units of this type to be stored in the vending machine
  - `minRestockingLevel` - the minimum allowable number of units of this type to be stored in the vending machine
  - `createdAt` - the date the inventory item was created and catalogued
  - `updatedAt` - the date the inventory item was last updated (either via a transaction or something else)
  - `isActive` - whether or not the inventory item is "active", which is not currently actively used by the Automated Vending reference implementation for any specific purposes
- _Audit Log_ - an audit log entry contains the following attributes:
  - `cardId` - card number
  - `accountId` - account number
  - `roleId` - the role
  - `personId` - the ID of the person who is associated with the card
  - `inventoryDelta` - what was changed in inventory
  - `createdAt` - the transaction date
  - `auditEntryId` - and a UUID representing the transaction itself uniquely

The `ms-inventory` microservice receives REST API calls from the upstream [`as-vending`](https://github.com/intel-retail/automated-vending/tree/main/as-vending) application service during a typical vending workflow. Typically, an individual will swipe a card, the workflow will start, and the inventory will be manipulated after an individual has removed or added items to the vending machine and an inference has completed. REST API calls to this service are not locked behind any authentication mechanism.

### Inventory service APIs

---

#### `GET`: `/inventory`

The `GET` call will return the entire inventory in JSON format.

Simple usage example:

```bash
curl -X GET http://localhost:48095/inventory
```

Sample response:

```json
{
  "content": "{\"data\":[{\"sku\":\"4900002470\",\"itemPrice\":1.99,\"productName\":\"Sprite (Lemon-Lime) - 16.9 oz\",\"unitsOnHand\":0,\"maxRestockingLevel\":24,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1567787309\",\"isActive\":true},{\"sku\":\"1200010735\",\"itemPrice\":1.99,\"productName\":\"Mountain Dew (Low Calorie) - 16.9 oz\",\"unitsOnHand\":0,\"maxRestockingLevel\":18,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1567787309\",\"isActive\":true},{\"sku\":\"1200050408\",\"itemPrice\":1.99,\"productName\":\"Mountain Dew - 16.9 oz\",\"unitsOnHand\":0,\"maxRestockingLevel\":6,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1567787309\",\"isActive\":true},{\"sku\":\"7800009257\",\"itemPrice\":1.99,\"productName\":\"Water (Dejablue) - 16.9 oz\",\"unitsOnHand\":0,\"maxRestockingLevel\":24,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1567787309\",\"isActive\":true},{\"sku\":\"4900002762\",\"itemPrice\":1.99,\"productName\":\"Dasani Water - 16.9 oz\",\"unitsOnHand\":0,\"maxRestockingLevel\":32,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1567787309\",\"isActive\":true},{\"sku\":\"1200081119\",\"itemPrice\":1.99,\"productName\":\"Pepsi (Wild Cherry) - 16.9 oz\",\"unitsOnHand\":0,\"maxRestockingLevel\":12,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1567787309\",\"isActive\":true},{\"sku\":\"1200018402\",\"itemPrice\":1.99,\"productName\":\"Mountain Dew (blue) - 16.9 oz\",\"unitsOnHand\":0,\"maxRestockingLevel\":6,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1567787309\",\"isActive\":true},{\"sku\":\"4900002469\",\"itemPrice\":1.99,\"productName\":\"Diet Coke - 16.9 oz\",\"unitsOnHand\":0,\"maxRestockingLevel\":24,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1567787309\",\"isActive\":true},{\"sku\":\"490440\",\"itemPrice\":1.99,\"productName\":\"Coca-Cola - 20 oz\",\"unitsOnHand\":0,\"maxRestockingLevel\":72,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1567787309\",\"isActive\":true}]}",
  "contentType": "json",
  "statusCode": 200,
  "error": false
}
```

---

#### `POST`: `/inventory`

The `POST` call will add a list of items into inventory and will return the newly added items as a JSON string in the `content` field of the response. Will also behave like a `PATCH` and supports updating the inventory in accordance with the submitted list of objects, each containing the fields to update based on matched SKU values.

Simple usage example:

```bash
curl -X POST -d '[{"createdAt": "1567787309","isActive": true,"itemPrice": 3.00,"maxRestockingLevel": 24,"minRestockingLevel": 0,"sku": "4900002470","unitsOnHand": 0,"updatedAt": "1567787309"}]' http://localhost:48095/inventory
```

Sample response:

```json
{
  "content": "[{\"sku\":\"4900002470\",\"itemPrice\":3,\"productName\":\"Sprite (Lemon-Lime) - 16.9 oz\",\"unitsOnHand\":0,\"maxRestockingLevel\":24,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1578955062042600972\",\"isActive\":true}]",
  "contentType": "json",
  "statusCode": 200,
  "error": false
}
```

---

#### `POST`: `/inventory/delta`

The `POST` call will increment or decrement inventory item(s) by a provided `delta` that match the given `SKU` numbers, and will return a JSON string containing the updated inventory items in the `content` field of the response.

Simple usage example:

```bash
curl -X POST -d '[{"SKU":"7800009257","delta":-1000},{"SKU":"7800009257","delta":-1000}]' http://localhost:48095/inventory/delta
```

Sample response:

```json
{
  "content": "[{\"sku\":\"7800009257\",\"itemPrice\":1.99,\"productName\":\"Water (Dejablue) - 16.9 oz\",\"unitsOnHand\":-1000,\"maxRestockingLevel\":24,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1567787309\",\"isActive\":true},{\"sku\":\"7800009257\",\"itemPrice\":1.99,\"productName\":\"Water (Dejablue) - 16.9 oz\",\"unitsOnHand\":-2000,\"maxRestockingLevel\":24,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1567787309\",\"isActive\":true}]",
  "contentType": "json",
  "statusCode": 200,
  "error": false
}
```

---

#### `GET`: `/inventory/{sku}`

The `GET` call will return a JSON string of a single inventory item whose SKU matches the URL parameter `{sku}` in the `content` field of the response.

Simple usage example:

```bash
curl -X GET http://localhost:48095/inventory/4900002470
```

Sample response:

```json
{
  "content": "{\"sku\":\"4900002470\",\"itemPrice\":3,\"productName\":\"Sprite (Lemon-Lime) - 16.9 oz\",\"unitsOnHand\":0,\"maxRestockingLevel\":24,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1578955062042600972\",\"isActive\":true}",
  "contentType": "json",
  "statusCode": 200,
  "error": false
}
```

If the `{sku}` does not correspond to a known item in the inventory, the response is:

```json
{
  "content": "",
  "contentType": "string",
  "statusCode": 404,
  "error": false
}
```

---

#### `DELETE`: `/inventory/{sku}`

The `DELETE` call will delete an inventory item whose SKU matches the URL parameter `{sku}` and return the deleted inventory item in the `content` field of the responses.

Simple usage example:

```bash
curl -X DELETE http://localhost:48095/inventory/4900002470
```

Sample response:

```json
{
  "content": "{\"sku\":\"4900002470\",\"itemPrice\":1.99,\"productName\":\"Sprite (Lemon-Lime) - 16.9 oz\",\"unitsOnHand\":0,\"maxRestockingLevel\":24,\"minRestockingLevel\":0,\"createdAt\":\"1567787309\",\"updatedAt\":\"1567787309\",\"isActive\":true}",
  "contentType": "json",
  "statusCode": 200,
  "error": false
}
```

If the provided `{sku}` does not correspond to a known item in the inventory, the response is:

```json
{
  "content": "Item does not exist",
  "contentType": "string",
  "statusCode": 404,
  "error": false
}
```

---

#### `GET`: `/auditlog`

The `GET` call on this API endpoint will return the entire audit log in JSON format.

Simple usage example:

```bash
curl -X GET http://localhost:48095/auditlog
```

Sample response:

```json
{
  "content": "{\"data\":[{\"cardId\":\"0003293374\",\"accountId\":1,\"roleId\":2,\"personId\":1,\"inventoryDelta\":[{\"SKU\":\"4900002470\",\"delta\":24},{\"SKU\":\"1200010735\",\"delta\":18},{\"SKU\":\"1200050408\",\"delta\":6},{\"SKU\":\"7800009257\",\"delta\":24},{\"SKU\":\"4900002762\",\"delta\":32},{\"SKU\":\"1200081119\",\"delta\":12},{\"SKU\":\"1200018402\",\"delta\":6},{\"SKU\":\"4900002469\",\"delta\":24},{\"SKU\":\"490440\",\"delta\":72}],\"createdAt\":\"1588006102888406420\",\"auditEntryId\":\"f944b60b-e389-4054-9643-2a33e4a0b227\"}]}",
  "contentType": "json",
  "statusCode": 200,
  "error": false
}
```

---

#### `POST`: `/auditlog`

The `POST` call on this API endpoint will add one entry into the audit log and will return the added entry as a JSON string in the `content` field of the response.

Simple usage example:

```bash
curl -X POST -d '{"cardId": "0","roleId": 0,"personId": 0,"inventoryDelta": [{"SKU": "000","delta":-1}],"createdAt": "000"}' http://localhost:48095/auditlog
```

Sample response:

```json
{
  "content": "{\"cardId\":\"0\",\"accountId\":0,\"roleId\":0,\"personId\":0,\"inventoryDelta\":[{\"SKU\":\"000\",\"delta\":-1}],\"createdAt\":\"1588006208233972031\",\"auditEntryId\":\"b61bed78-da3b-4862-b548-b4ab16574495\"}",
  "contentType": "json",
  "statusCode": 200,
  "error": false
}
```

---

#### `GET`: `/auditlog/{auditEntryId}`

The `GET` call on this API endpoint will will return a JSON string of a single audit log entry whose `auditEntryId` (which is a UUID) matches the one specified in the URL.

Simple usage example:

```bash
curl -X GET http://localhost:48095/auditlog/b61bed78-da3b-4862-b548-b4ab16574495
```

Sample response:

```json
{
  "content": "{\"cardId\":\"0\",\"accountId\":0,\"roleId\":0,\"personId\":0,\"inventoryDelta\":[{\"SKU\":\"000\",\"delta\":-1}],\"createdAt\":\"1588006208233972031\",\"auditEntryId\":\"b61bed78-da3b-4862-b548-b4ab16574495\"}",
  "contentType": "json",
  "statusCode": 200,
  "error": false
}
```

If the `{auditEntryId}` parameter does not correspond to an existing audit log entry, the response is:

```json
{
  "content": "",
  "contentType": "string",
  "statusCode": 404,
  "error": false
}
```

---

#### `DELETE`: `/auditlog/{auditEntryId}`

The `DELETE` call on this API endpoint will delete an audit log entry whose `auditEntryId` (which is a UUID) matches the URL parameter `{auditEntryId}` and will return a JSON string containing the deleted audit log entry in the `content` field of the response

Simple usage example:

```bash
curl -X DELETE http://localhost:48095/auditlog/b61bed78-da3b-4862-b548-b4ab16574495
```

Sample response:

```json
{
  "content": "{\"cardId\":\"0\",\"accountId\":0,\"roleId\":0,\"personId\":0,\"inventoryDelta\":[{\"SKU\":\"000\",\"delta\":-1}],\"createdAt\":\"1588014993644633617\",\"auditEntryId\":\"c555d26e-9b3d-4ae8-8053-d2270e40ccf0\"}",
  "contentType": "json",
  "statusCode": 200,
  "error": false
}
```

If the `{auditEntryId}` parameter does not correspond to an existing audit log entry, the response is:

```json
{
  "content": "Item does not exist",
  "contentType": "string",
  "statusCode": 404,
  "error": false
}
```

---

## Ledger service

### Ledger service description

The `ms-ledger` microservice updates a ledger with the current transaction information (products purchased, quantity, total price, transaction timestamp). Transactions are added to the consumer's account. Transactions also have an `isPaid` attribute to designate which transactions have been paid/unpaid.

This microservice returns the current transaction to the [`as-vending`](https://github.com/intel-retail/automated-vending/tree/main/as-vending) microservice, which then calls the [`ds-controller-board`](https://github.com/intel-retail/automated-vending/tree/main/ds-controller-board) microservice to display the items purchased and the total price of the transaction on the LCD.

### Ledger service APIs

#### `GET`: `/ledger`

The `GET` call will return the entire ledger in JSON format.

Simple usage example:

```bash
curl -X GET http://localhost:48093/ledger
```

Sample response:

```json
{
  "content": "{\"data\":[{\"accountID\":1,\"ledgers\":[{\"transactionID\":\"1588006480995452968\",\"txTimeStamp\":\"1588006480995453037\",\"lineTotal\":7.96,\"createdAt\":\"1588006480995453110\",\"updatedAt\":\"1588006480995453171\",\"isPaid\":false,\"lineItems\":[{\"sku\":\"1200050408\",\"productName\":\"Mountain Dew - 16.9 oz\",\"itemPrice\":1.99,\"itemCount\":3},{\"sku\":\"7800009257\",\"productName\":\"Water (Dejablue) - 16.9 oz\",\"itemPrice\":1.99,\"itemCount\":1}]}]},{\"accountID\":2,\"ledgers\":[]},{\"accountID\":3,\"ledgers\":[]},{\"accountID\":4,\"ledgers\":[]},{\"accountID\":5,\"ledgers\":[]},{\"accountID\":6,\"ledgers\":[]}]}",
  "contentType": "json",
  "statusCode": 200,
  "error": false
}
```

---

#### `POST`: `/ledger`

The `POST` call will create a transaction and add it to the ledger for the specified `accountId` in the JSON body.

Simple usage example:

```bash
curl -X POST -d '{"accountId":1,"deltaSKUs":[{"sku":"1200050408","delta":-1}]}' http://localhost:48093/ledger
```

Sample response:

```json
{
  "content": "{\"transactionID\":\"1588006579251812793\",\"txTimeStamp\":\"1588006579251812850\",\"lineTotal\":1.99,\"createdAt\":\"1588006579251812909\",\"updatedAt\":\"1588006579251812968\",\"isPaid\":false,\"lineItems\":[{\"sku\":\"1200050408\",\"productName\":\"Mountain Dew - 16.9 oz\",\"itemPrice\":1.99,\"itemCount\":1}]}",
  "contentType": "json",
  "statusCode": 200,
  "error": false
}
```

---

#### `GET`: `/ledger/{accountid}`

The `GET` call will return the ledger for a specified `{accountid}`.

Simple usage example:

```bash
curl -X GET http://localhost:48093/ledger/1
```

Sample response:

```json
{
  "content": "{\"accountID\":1,\"ledgers\":[{\"transactionID\":\"1588006480995452968\",\"txTimeStamp\":\"1588006480995453037\",\"lineTotal\":7.96,\"createdAt\":\"1588006480995453110\",\"updatedAt\":\"1588006480995453171\",\"isPaid\":false,\"lineItems\":[{\"sku\":\"1200050408\",\"productName\":\"Mountain Dew - 16.9 oz\",\"itemPrice\":1.99,\"itemCount\":3},{\"sku\":\"7800009257\",\"productName\":\"Water (Dejablue) - 16.9 oz\",\"itemPrice\":1.99,\"itemCount\":1}]},{\"transactionID\":\"1588006579251812793\",\"txTimeStamp\":\"1588006579251812850\",\"lineTotal\":1.99,\"createdAt\":\"1588006579251812909\",\"updatedAt\":\"1588006579251812968\",\"isPaid\":false,\"lineItems\":[{\"sku\":\"1200050408\",\"productName\":\"Mountain Dew - 16.9 oz\",\"itemPrice\":1.99,\"itemCount\":1}]}]}",
  "contentType": "json",
  "statusCode": 200,
  "error": false
}
```

If the provided `{accountid}` parameter does not correspond to a valid ledger account number, the response is:

```json
{
  "content": "AccountID not found in ledger",
  "contentType": "string",
  "statusCode": 400,
  "error": false
}
```

---

#### `POST`: `/ledger/ledgerPaymentUpdate`

The `POST` call will update the transaction in the ledger of the specified account.

Simple usage example:

```bash
curl -X POST -d '{"accountId":1,"transactionID":"1588006579251812793","isPaid":true}' http://localhost:48093/ledgerPaymentUpdate
```

Sample response:

```json
{
  "content": "Updated Payment Status for transaction 1588006579251812793",
  "contentType": "string",
  "statusCode": 200,
  "error": false
}
```

If the provided `transactionID` does not correspond to an existing transaction in the ledger, the response is:

```json
{
  "content": "Could not find Transaction 1588006579251812793",
  "contentType": "string",
  "statusCode": 400,
  "error": true
}
```

---

#### `DELETE`: `/ledger/{accountid}/{transactionid}`

The `DELETE` call will delete the transaction by its `transactionid` from the ledger for the specified account by its `accountid`.

Simple usage example:

```bash
curl -X DELETE http://localhost:48093/ledger/1/1588006579251812793
```

Sample response:

```json
{
  "content": "Deleted ledger 1588006579251812793",
  "contentType": "string",
  "statusCode": 200,
  "error": false
}
```

If the provided `transactionID` does not correspond to an existing transaction in the ledger, the response is:

```json
{
  "content": "Could not find Transaction 1588006579251812793",
  "contentType": "string",
  "statusCode": 400,
  "error": true
}
```

---

#### `How to add to CORS settings and Enable CORS`

Please refer to [EdgeX kamakura documentation on how to add CORS settings and Enable CORS](https://github.com/edgexfoundry/edgex-docs/blob/kamakura/docs_src/security/Ch-CORS-Settings.md)

---