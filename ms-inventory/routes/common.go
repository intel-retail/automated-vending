// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"errors"

	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

// AuditLogFileName is the name of the file that will store the audit log
var AuditLogFileName = "auditlog.json"

// InventoryFileName is the name of the file that will store the inventory
var InventoryFileName = "inventory.json"

// DeleteAllQueryString is a string used across this module to enable
// CRUD operations on "all" items in inventory or audit log
const DeleteAllQueryString = "all"

// GetInventoryItems returns a list of InventoryItems by reading the inventory
// JSON file
func GetInventoryItems() (inventoryItems Products, err error) {
	err = utilities.LoadFromJSONFile(InventoryFileName, &inventoryItems)
	if err != nil {
		return inventoryItems, errors.New(
			"Failed to load inventory JSON file: " + err.Error(),
		)
	}
	return
}

// GetInventoryItemBySKU returns an inventory item by reading from the
// inventory JSON file
func GetInventoryItemBySKU(SKU string) (inventoryItem Product, inventoryItems Products, err error) {
	inventoryItems, err = GetInventoryItems()
	if err != nil {
		return Product{}, Products{}, errors.New(
			"Failed to get inventory items: " + err.Error(),
		)
	}
	for _, inventoryItem := range inventoryItems.Data {
		if SKU == inventoryItem.SKU {
			return inventoryItem, inventoryItems, nil
		}
	}
	return Product{SKU: ""}, inventoryItems, nil
}

// GetAuditLog returns a list of audit log entries by reading from the
// audit log JSON file
func GetAuditLog() (auditLog AuditLog, err error) {
	err = utilities.LoadFromJSONFile(AuditLogFileName, &auditLog)
	if err != nil {
		return auditLog, errors.New(
			"Failed to load audit log JSON file: " + err.Error(),
		)
	}
	return
}

// GetAuditLogEntryByID returns an audit log entry by reading from the
// audit log JSON file
func GetAuditLogEntryByID(auditEntryID string) (auditLogEntry AuditLogEntry, auditLogEntries AuditLog, err error) {
	auditLogEntries, err = GetAuditLog()
	if err != nil {
		return AuditLogEntry{}, AuditLog{}, errors.New(
			"Failed to get audit log items: " + err.Error(),
		)
	}
	for _, auditLogEntry := range auditLogEntries.Data {
		if auditEntryID == auditLogEntry.AuditEntryID {
			return auditLogEntry, auditLogEntries, nil
		}
	}
	return AuditLogEntry{}, auditLogEntries, nil
}

// DeleteInventory will reset the content of the inventory JSON file
func DeleteInventory() error {
	return WriteJSON(InventoryFileName, Products{Data: []Product{}})
}

// DeleteAuditLog will reset the content of the audit log JSON file
func DeleteAuditLog() error {
	return WriteJSON(AuditLogFileName, AuditLog{Data: []AuditLogEntry{}})
}

// WriteJSON is a shorthand for writing an interface to JSON
func WriteJSON(fileName string, content interface{}) error {
	return utilities.WriteToJSONFile(fileName, content, 0644)
}

// WriteInventory is a shorthand for writing the inventory quickly
func (inventoryItems *Products) WriteInventory() error {
	return WriteJSON(InventoryFileName, inventoryItems)
}

// WriteAuditLog is a shorthand for writing the audit log quickly
func (auditLog *AuditLog) WriteAuditLog() error {
	return WriteJSON(AuditLogFileName, auditLog)
}

// DeleteInventoryItem deletes an inventory item matching the
// specified SKU
func (inventoryItems *Products) DeleteInventoryItem(inventoryItem Product) {
	for i, item := range inventoryItems.Data {
		if item.SKU == inventoryItem.SKU {
			inventoryItems.Data = append(inventoryItems.Data[:i], inventoryItems.Data[i+1:]...)
			break
		}
	}
}

// DeleteAuditLogEntry deletes an audit log entry item matching the
// specified EntryID
func (auditLog *AuditLog) DeleteAuditLogEntry(auditLogEntry AuditLogEntry) {
	for i, item := range auditLog.Data {
		if item.AuditEntryID == auditLogEntry.AuditEntryID {
			auditLog.Data = append(auditLog.Data[:i], auditLog.Data[i+1:]...)
			break
		}
	}
}
