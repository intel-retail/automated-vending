// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"as-vending/config"
	"fmt"
	"time"

	clientInterfaces "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces"
)

// VendingState is a representation of the entire state of vending workflow.
// The information stored in this is shared across this application service.
// Information about the state of the vending workflow should generally
// be stored in this struct.
type VendingState struct {
	CVWorkflowStarted              bool       `json:"cvWorkflowStarted"`
	MaintenanceMode                bool       `json:"MaintenanceMode"`
	CurrentUserData                OutputData `json:"personID"`
	DoorClosed                     bool       `json:"doorClosed"`
	ThreadStopChannel              chan int   `json:"threadStopChannel"`            // global stop channel for threads
	DoorOpenedDuringCVWorkflow     bool       `json:"doorOpenedDuringCVWorkflow  "` // door open event
	DoorOpenWaitThreadStopChannel  chan int   `json:"doorOpenWaitThreadStopChannel"`
	DoorClosedDuringCVWorkflow     bool       `json:"doorClosedDuringCVWorkflow  "` //door close event
	DoorCloseWaitThreadStopChannel chan int   `json:"doorCloseWaitThreadStopChannel"`
	InferenceDataReceived          bool       `json:"inferenceDataReceived"` // inference event
	InferenceWaitThreadStopChannel chan int   `json:"inferenceWaitThreadStopChannel"`
	Configuration                  *config.VendingConfig
	CommandClient                  clientInterfaces.CommandClient
	DoorCloseStateTimeout          time.Duration
	DoorOpenStateTimeout           time.Duration
	InferenceTimeout               time.Duration
}

// MaintenanceMode is a simple structure used to return the state of
// maintenance mode to REST API consumers.
type MaintenanceMode struct {
	MaintenanceMode bool `json:"maintenanceMode"`
}

// ControllerBoardStatus represents the status of the controller board,
// which is pushed into this application service from the
// as-controller-board-status service as a REST request.
type ControllerBoardStatus struct {
	Lock1                int     `json:"lock1_status"`
	Lock2                int     `json:"lock2_status"`
	DoorClosed           bool    `json:"door_closed"` // true means the door is closed and false means the door is open
	Temperature          float64 `json:"temperature"`
	Humidity             float64 `json:"humidity"`
	MinTemperatureStatus bool    `json:"minTemperatureStatus"`
	MaxTemperatureStatus bool    `json:"maxTemperatureStatus"`
}

// Ledger is the data structure that represents financial ledger transactions,
// and comes from the ledger service.
type Ledger struct {
	TransactionID int64      `json:"transactionID,string"`
	TxTimeStamp   int64      `json:"txTimeStamp,string"`
	LineTotal     float64    `json:"lineTotal"`
	CreatedAt     int64      `json:"createdAt,string"`
	UpdatedAt     int64      `json:"updatedAt,string"`
	IsPaid        bool       `json:"isPaid"`
	LineItems     []LineItem `json:"lineItems"`
}

// LineItem is a single item contained in the Ledger.
type LineItem struct {
	SKU         string  `json:"sku"`
	ProductName string  `json:"productName"`
	ItemPrice   float64 `json:"itemPrice"`
	ItemCount   int     `json:"itemCount"`
}

// deltaLedger is a representation of a set of deltaSKUs from an upstream
// inference service.
type deltaLedger struct {
	AccountID int        `json:"accountId"`
	DeltaSKUs []deltaSKU `json:"deltaSKUs"`
}

// deltaSKU is a single representation of an integer quantity change of a
// specific SKU in inventory. An inference will produce a list of deltaSKUs
// when someone removes items from inventory.
type deltaSKU struct {
	SKU   string `json:"SKU"`
	Delta int    `json:"delta"`
}

// OutputData represents the authentication information associated with
// a person that has been authenticated to open the vending machine and
// remove items from inventory for purchase. This information is pushed to
// the vending state and shared throughout this application service.
type OutputData struct {
	AccountID int    `json:"accountID"`
	PersonID  int    `json:"personID"`
	RoleID    int    `json:"roleID"`
	CardID    string `json:"cardID"`
}

// AuditLogEntry is the representation of an inventory transaction that
// occurs when someone opens the vending machine. Regardless of how many
// items have been taken, an audit log transaction will always be created.
type AuditLogEntry struct {
	CardID         string     `json:"cardId"`
	AccountID      int        `json:"accountId"`
	RoleID         int        `json:"roleId"`
	PersonID       int        `json:"personId"`
	InventoryDelta []deltaSKU `json:"inventoryDelta"`
	CreatedAt      int64      `json:"createdAt,string"`
	AuditEntryID   string     `json:"auditEntryId"`
}

func (vs *VendingState) ParseDurationFromConfig() error {
	var err error
	vs.DoorCloseStateTimeout, err = time.ParseDuration(vs.Configuration.DoorCloseStateTimeoutDuration)
	if err != nil {
		return fmt.Errorf("failed to parse DoorCloseStateTimeoutDuration configuration: %v", err)
	}

	vs.DoorOpenStateTimeout, err = time.ParseDuration(vs.Configuration.DoorOpenStateTimeoutDuration)
	if err != nil {
		return fmt.Errorf("failed to parse DoorOpenStateTimeoutDuration configuration: %v", err)
	}

	vs.InferenceTimeout, err = time.ParseDuration(vs.Configuration.InferenceTimeoutDuration)
	if err != nil {
		return fmt.Errorf("failed to parse InferenceTimeoutDuration configuration: %v", err)
	}
	return nil
}
