// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// InventoryGet allows for the retrieval of the entire inventory
func (c *Controller) InventoryGet(writer http.ResponseWriter, req *http.Request) {
	inventoryItems, err := c.GetInventoryItems()
	c.inventoryItems = inventoryItems
	if err != nil {
		c.lc.Errorf("Failed to retrieve all inventory items: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Failed to retrieve all inventory items: " + err.Error()))
		return
	}

	// No logic needs to be done here, since we are just reading the file
	// and writing it back out. Simply marshaling it will validate its structure
	inventoryItemsJSON, err := json.Marshal(inventoryItems)
	if err != nil {
		c.lc.Errorf("Failed to process all inventory items: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Failed to process inventory items: " + err.Error()))
		return
	}
	c.lc.Infof("Successfully retrieved all inventory items")
	writer.Write(inventoryItemsJSON)
}

// InventoryItemGet allows for a single inventory item to be retrieved by SKU
func (c *Controller) InventoryItemGet(writer http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	sku := vars["sku"]
	if sku != "" {
		inventoryItem, _, err := c.GetInventoryItemBySKU(sku)
		if err != nil {
			c.lc.Errorf("Failed to get inventory item by SKU: %s with error: %s", sku, err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte("Failed to get inventory item by SKU: " + err.Error()))
			return
		}
		if inventoryItem.SKU == "" {
			c.lc.Infof("SKU is empty")
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte(""))
			return
		}
		outputInventoryItemJSON, err := json.Marshal(inventoryItem)
		if err != nil {
			c.lc.Errorf("Failed to process inventory item with SKU: %s with error: %s", sku, err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte("Failed to process the requested inventory item " + sku + ":" + err.Error()))
			return
		}
		c.lc.Infof("Succcessfully got inventory item by SKU: %s", sku)
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(outputInventoryItemJSON))
		return
	}
	c.lc.Error("Valid inventory item not in the form of /inventory/{sku}")
	writer.WriteHeader(http.StatusBadRequest)
	writer.Write([]byte("Please enter a valid inventory item in the form of /inventory/{sku}"))
}

// AuditLogGetAll allows all audit log entries to be retrieved
func (c *Controller) AuditLogGetAll(writer http.ResponseWriter, req *http.Request) {
	auditLog, err := c.GetAuditLog()
	c.auditLog = auditLog
	if err != nil {
		c.lc.Errorf("Failed to retrieve all audit log entries: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Failed to retrieve all audit log entries: " + err.Error()))
		return
	}

	// No logic needs to be done here, since we are just reading the file
	// and writing it back out. Simply marshaling it will validate its structure
	auditLogJSON, err := json.Marshal(auditLog)
	if err != nil {
		c.lc.Errorf("Failed to process audit log entries: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Failed to process audit log entries: " + err.Error()))
		return
	}
	c.lc.Infof("Successfully retrieved all audit log entries")
	writer.Write(auditLogJSON)
}

// AuditLogGetEntry allows a single audit log entry to be retrieved by its
// UUID
func (c *Controller) AuditLogGetEntry(writer http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	entryID := vars["entry"]
	if entryID != "" {
		auditLogEntry, _, err := c.GetAuditLogEntryByID(entryID)
		if err != nil {
			c.lc.Errorf("Failed to get audit log entry ID: %s with error: %s", entryID, err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte("Failed to get audit log entry by ID: " + err.Error()))
			return
		}
		if auditLogEntry.AuditEntryID == "" {
			c.lc.Info("Audit log entry is not set")
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		outputAuditLogEntryJSON, err := json.Marshal(auditLogEntry)
		if err != nil {
			c.lc.Errorf("Failed to process the requested audit log entry item: %s with error: %s", entryID, err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte("Failed to process the requested audit log entry item " + entryID + ":" + err.Error()))
			return
		}
		c.lc.Info("Successfully retrieved audit log entry with id: %s", entryID)
		writer.Write(outputAuditLogEntryJSON)
		return
	}
	c.lc.Info("valid entry ID in the form of /auditlog/{entry} not set")
	writer.WriteHeader(http.StatusBadRequest)
	writer.Write([]byte("Please enter a valid entry ID in the form of /auditlog/{entry}"))
}
