// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// DeleteAllQueryString is a string used across this module to enable
// CRUD operations on "all" items in inventory or audit log
const (
	DeleteAllQueryString = "all"
)

// GetInventoryItems returns a list of InventoryItems by reading the inventory
// JSON file
func (c *Controller) GetInventoryItems() (inventoryItems Products, err error) {
	data, err := os.ReadFile(c.inventoryFileName)
	if err != nil {
		return inventoryItems, fmt.Errorf("failed to read from inventory file: %s", err.Error())
	}
	if err := json.Unmarshal(data, &inventoryItems); err != nil {
		return inventoryItems, fmt.Errorf("failed to unmarshal inventory file: %s", err.Error())
	}

	return
}

// GetInventoryItemBySKU returns an inventory item by reading from the
// inventory JSON file
func (c *Controller) GetInventoryItemBySKU(SKU string) (inventoryItem Product, inventoryItems Products, err error) {
	inventoryItems, err = c.GetInventoryItems()
	if err != nil {
		c.lc.Errorf("Failed to get inventory items: %s", err.Error())
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
func (c *Controller) GetAuditLog() (auditLog AuditLog, err error) {
	data, err := os.ReadFile(c.auditLogFileName)
	if err != nil {
		return auditLog, fmt.Errorf("failed to read from audit log JSON file: %s", err.Error())
	}

	if err := json.Unmarshal(data, &auditLog); err != nil {
		return auditLog, fmt.Errorf("failed to unmarshal audit log JSON file: %s", err.Error())
	}

	return
}

// GetAuditLogEntryByID returns an audit log entry by reading from the
// audit log JSON file
func (c *Controller) GetAuditLogEntryByID(auditEntryID string) (auditLogEntry AuditLogEntry, auditLogEntries AuditLog, err error) {
	auditLogEntries, err = c.GetAuditLog()
	if err != nil {
		c.lc.Errorf("Failed to get audit log items: %s" + err.Error())
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
func (c *Controller) DeleteInventory() error {
	c.lc.Debug("Inventory JSON content reset")
	return c.WriteJSON(c.inventoryFileName, Products{Data: []Product{}})
}

// DeleteAuditLog will reset the content of the audit log JSON file
func (c *Controller) DeleteAuditLog() error {
	c.lc.Debug("Audit Log JSON content reset")
	return c.WriteJSON(c.auditLogFileName, AuditLog{Data: []AuditLogEntry{}})
}

// WriteJSON is a shorthand for writing an interface to JSON
func (c *Controller) WriteJSON(fileName string, content interface{}) error {
	c.lc.Debugf("Writing: %s to Inventory JSON: %s", content, fileName)
	data, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %s", err.Error())
	}
	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %s", err.Error())
	}
	return nil
}

// WriteInventory is a shorthand for writing the inventory quickly
func (c *Controller) WriteInventory() error {
	c.lc.Debugf("Wrote: %s to Inventory: %s", c.inventoryItems, c.inventoryFileName)
	return c.WriteJSON(c.inventoryFileName, c.inventoryItems)
}

// WriteAuditLog is a shorthand for writing the audit log quickly
func (c *Controller) WriteAuditLog() error {
	c.lc.Debugf("Wrote: %s to Inventory: %s", c.auditLog, c.auditLogFileName)
	return c.WriteJSON(c.auditLogFileName, c.auditLog)
}

// DeleteInventoryItem deletes an inventory item matching the
// specified SKU
func (c *Controller) DeleteInventoryItem(inventoryItem Product) {
	for i, item := range c.inventoryItems.Data {
		if item.SKU == inventoryItem.SKU {
			c.inventoryItems.Data = append(c.inventoryItems.Data[:i], c.inventoryItems.Data[i+1:]...)
			c.lc.Debugf("Deleted: %s from inventory", inventoryItem.SKU)
			break
		}
	}
}

// DeleteAuditLogEntry deletes an audit log entry item matching the
// specified EntryID
func (c *Controller) DeleteAuditLogEntry(auditLogEntry AuditLogEntry) {
	for i, item := range c.auditLog.Data {
		if item.AuditEntryID == auditLogEntry.AuditEntryID {
			c.auditLog.Data = append(c.auditLog.Data[:i], c.auditLog.Data[i+1:]...)
			c.lc.Debugf("Deleted: %s from audit log", auditLogEntry.AuditEntryID)
			break
		}
	}
}
