// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

// InventoryDelete allows deletion of an inventory item or multiple items
func (c *Controller) InventoryDelete(writer http.ResponseWriter, req *http.Request) {
	// find the requested SKU and exit if it's invalid
	vars := mux.Vars(req)
	SKU := vars["sku"]
	if SKU == "" {
		c.lc.Errorf("Empty inventory item SKU")
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Please enter a valid inventory item in the form of /inventory/{sku}", true)
		return
	}
	// if the user wants to delete all inventory, do it
	if SKU == DeleteAllQueryString {
		err := c.DeleteInventory()
		if err != nil {
			c.lc.Errorf("Failed to properly reset inventory: %s", err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to properly reset inventory: "+err.Error(), true)
			return
		}
		emptyInventoryResponseJSON, err := utilities.GetAsJSON(Products{Data: []Product{}})
		if err != nil {
			c.lc.Errorf("Failed to serialize empty inventory response: %s", err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to serialize empty inventory response: "+err.Error(), true)
			return
		}

		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, emptyInventoryResponseJSON, false)
		return
	}
	// look up the requested inventory item by SKU
	inventoryItemToDelete, inventoryItems, err := c.GetInventoryItemBySKU(SKU)
	c.inventoryItems = inventoryItems
	if err != nil {
		c.lc.Errorf("Failed to get requested inventory item by SKU: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to get requested inventory item by SKU: "+err.Error(), true)
		return
	}

	// check if GetInventoryItemBySKU found an inventory item to delete
	if inventoryItemToDelete.SKU == "" {
		c.lc.Info("Item does not exist")
		utilities.WriteStringHTTPResponse(writer, req, http.StatusNotFound, "Item does not exist", false)
		return
	}
	// delete the inventory item & write the modified inventory
	c.DeleteInventoryItem(inventoryItemToDelete)
	err = c.WriteInventory()
	if err != nil {
		c.lc.Errorf("Failed to write updated inventory: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to write updated inventory", true)
		return
	}
	inventoryItemToDeleteJSON, err := utilities.GetAsJSON(inventoryItemToDelete)
	if err != nil {
		c.lc.Errorf("Successfully deleted the item from inventory, but failed to serialize it so that it could be sent back to the requester: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, fmt.Sprintf("Successfully deleted the item from inventory, but failed to serialize it so that it could be sent back to the requester: %v", err.Error()), true)
	}
	c.lc.Infof("Successfully deleted the item: %s from inventory", inventoryItemToDelete.SKU)
	utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, inventoryItemToDeleteJSON, false)
}

// AuditLogDelete allows deletion of one or more audit log entry items
func (c *Controller) AuditLogDelete(writer http.ResponseWriter, req *http.Request) {
	// find the requested SKU and exit if it's invalid
	vars := mux.Vars(req)
	entryID := vars["entry"]
	if entryID == "" {
		c.lc.Error("EntryID is empty")
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Please enter a valid audit log entry ID in the form of /auditlog/{entryId}", true)
		return
	}
	// if the user wants to delete all inventory, do it
	if entryID == DeleteAllQueryString {
		err := c.DeleteAuditLog()
		if err != nil {
			c.lc.Errorf("Failed to reset audit log: %s", err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to properly reset audit log: "+err.Error(), true)
			return
		}
		emptyAuditLogResponseJSON, err := utilities.GetAsJSON(AuditLog{Data: []AuditLogEntry{}})
		if err != nil {
			c.lc.Errorf("Failed to serialize empty audit log response: %s", err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, fmt.Sprintf("Failed to serialize empty audit log response: %v", err.Error()), true)
		}
		c.lc.Info("Successfully deleted audit log")
		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, emptyAuditLogResponseJSON, false)
		return
	}
	// look up the requested audit log entry by EntryID
	auditLogEntryToDelete, auditLog, err := c.GetAuditLogEntryByID(entryID)
	c.auditLog = auditLog
	if err != nil {
		c.lc.Errorf("Failed to get audit log entry ID: %s with error: %s", entryID, err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to get requested audit log entry by ID: "+err.Error(), true)
		return
	}

	// check if GetAuditLogEntryByID found an audit log entry to delete
	if auditLogEntryToDelete.AuditEntryID == "" {
		c.lc.Errorf("Item with entry ID: %s does not exist", auditLogEntryToDelete.AuditEntryID)
		utilities.WriteStringHTTPResponse(writer, req, http.StatusNotFound, "Item does not exist", false)
		return
	}
	// delete the audit log entry & write the modified audit log
	c.DeleteAuditLogEntry(auditLogEntryToDelete)
	err = c.WriteAuditLog()
	if err != nil {
		c.lc.Error("Failed to write updated audit log")
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to write updated audit log", true)
		return
	}
	auditLogEntryToDeleteJSON, err := utilities.GetAsJSON(auditLogEntryToDelete)
	if err != nil {
		c.lc.Errorf("Successfully deleted item: %s from audit log, but failed to serialize information back to the requester: %s", auditLogEntryToDelete.AuditEntryID, err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, fmt.Sprintf("Successfully deleted item from audit log, but failed to serialize information back to the requester: %v", err.Error()), true)
	}
	c.lc.Infof("Succssfully deleted item: %s from audit log", auditLogEntryToDelete.AuditEntryID)
	utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, auditLogEntryToDeleteJSON, false)
}
