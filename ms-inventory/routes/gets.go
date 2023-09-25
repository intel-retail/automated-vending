// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

// InventoryGet allows for the retrieval of the entire inventory
func (c *Controller) InventoryGet(writer http.ResponseWriter, req *http.Request) {
	inventoryItems, err := c.GetInventoryItems()
	c.inventoryItems = inventoryItems
	if err != nil {
		c.lc.Errorf("Failed to retrieve all inventory items: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to retrieve all inventory items: "+err.Error(), true)
		return
	}

	// No logic needs to be done here, since we are just reading the file
	// and writing it back out. Simply marshaling it will validate its structure
	inventoryItemsJSON, err := utilities.GetAsJSON(inventoryItems)
	if err != nil {
		c.lc.Errorf("Failed to process all inventory items: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to process inventory items: "+err.Error(), true)
		return
	}
	c.lc.Infof("Successfully retrieved all inventory items")
	utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, inventoryItemsJSON, false)
}

// InventoryItemGet allows for a single inventory item to be retrieved by SKU
func (c *Controller) InventoryItemGet(writer http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	sku := vars["sku"]
	if sku != "" {
		inventoryItem, _, err := c.GetInventoryItemBySKU(sku)
		if err != nil {
			c.lc.Errorf("Failed to get inventory item by SKU: %s with error: %s", sku, err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to get inventory item by SKU: "+err.Error(), true)
			return
		}
		if inventoryItem.SKU == "" {
			c.lc.Infof("SKU is empty")
			utilities.WriteStringHTTPResponse(writer, req, http.StatusNotFound, "", false)
			return
		}
		outputInventoryItemJSON, err := utilities.GetAsJSON(inventoryItem)
		if err != nil {
			c.lc.Errorf("Failed to process inventory item with SKU: %s with error: %s", sku, err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to process the requested inventory item "+sku+":"+err.Error(), true)
			return
		}
		c.lc.Infof("Succcessfully got inventory item by SKU: %s", sku)
		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, outputInventoryItemJSON, false)
		return
	}
	c.lc.Error("Valid inventory item not in the form of /inventory/{sku}")
	utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Please enter a valid inventory item in the form of /inventory/{sku}", false)
}

// AuditLogGetAll allows all audit log entries to be retrieved
func (c *Controller) AuditLogGetAll(writer http.ResponseWriter, req *http.Request) {
	auditLog, err := c.GetAuditLog()
	c.auditLog = auditLog
	if err != nil {
		c.lc.Errorf("Failed to retrieve all audit log entries: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to retrieve all audit log entries: "+err.Error(), true)
		return
	}

	// No logic needs to be done here, since we are just reading the file
	// and writing it back out. Simply marshaling it will validate its structure
	auditLogJSON, err := utilities.GetAsJSON(auditLog)
	if err != nil {
		c.lc.Errorf("Failed to process audit log entries: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to process audit log entries: "+err.Error(), true)
		return
	}
	c.lc.Infof("Successfully retrieved all audit log entries")
	utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, auditLogJSON, false)
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
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to get audit log entry by ID: "+err.Error(), true)
			return
		}
		if auditLogEntry.AuditEntryID == "" {
			c.lc.Info("Audit log entry is not set")
			utilities.WriteStringHTTPResponse(writer, req, http.StatusNotFound, "", false)
			return
		}
		outputAuditLogEntryJSON, err := utilities.GetAsJSON(auditLogEntry)
		if err != nil {
			c.lc.Errorf("Failed to process the requested audit log entry item: %s with error: %s", entryID, err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to process the requested audit log entry item "+entryID+":"+err.Error(), true)
			return
		}
		c.lc.Info("Successfully retrieved audit log entry with id: %s", entryID)
		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, outputAuditLogEntryJSON, false)
		return
	}
	c.lc.Info("valid entry ID in the form of /auditlog/{entry} not set")
	utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Please enter a valid entry ID in the form of /auditlog/{entry}", false)
}
