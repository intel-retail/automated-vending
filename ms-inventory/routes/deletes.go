// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// InventoryDelete allows deletion of an inventory item or multiple items
func (c *Controller) InventoryDelete(writer http.ResponseWriter, req *http.Request) {
	// find the requested SKU and exit if it's invalid
	vars := mux.Vars(req)
	SKU := vars["sku"]
	if SKU == "" {
		c.lc.Errorf("Empty inventory item SKU")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Please enter a valid inventory item in the form of /inventory/{sku}"))
		return
	}
	// if the user wants to delete all inventory, do it
	if SKU == DeleteAllQueryString {
		err := c.DeleteInventory()
		if err != nil {
			c.lc.Errorf("Failed to properly reset inventory: %s", err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte("Failed to properly reset inventory: " + err.Error()))
			return
		}
		emptyInventoryResponseJSON, err := json.Marshal(Products{Data: []Product{}})
		if err != nil {
			c.lc.Errorf("Failed to serialize empty inventory response: %s", err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte("Failed to serialize empty inventory response: " + err.Error()))
			return
		}
		writer.Write(emptyInventoryResponseJSON)
		return
	}
	// look up the requested inventory item by SKU
	inventoryItemToDelete, inventoryItems, err := c.GetInventoryItemBySKU(SKU)
	c.inventoryItems = inventoryItems
	if err != nil {
		c.lc.Errorf("Failed to get requested inventory item by SKU: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Failed to get requested inventory item by SKU: " + err.Error()))
		return
	}

	// check if GetInventoryItemBySKU found an inventory item to delete
	if inventoryItemToDelete.SKU == "" {
		c.lc.Info("Item does not exist")
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte("Item does not exist"))
		return
	}
	// delete the inventory item & write the modified inventory
	c.DeleteInventoryItem(inventoryItemToDelete)
	err = c.WriteInventory()
	if err != nil {
		c.lc.Errorf("Failed to write updated inventory: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Failed to write updated inventory"))
		return
	}
	inventoryItemToDeleteJSON, err := json.Marshal(inventoryItemToDelete)
	if err != nil {
		c.lc.Errorf("Successfully deleted the item from inventory, but failed to serialize it so that it could be sent back to the requester: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf("Successfully deleted the item from inventory, but failed to serialize it so that it could be sent back to the requester: %v", err.Error())))
	}
	c.lc.Infof("Successfully deleted the item: %s from inventory", inventoryItemToDelete.SKU)
	writer.Write(inventoryItemToDeleteJSON)
}

// AuditLogDelete allows deletion of one or more audit log entry items
func (c *Controller) AuditLogDelete(writer http.ResponseWriter, req *http.Request) {
	// find the requested SKU and exit if it's invalid
	vars := mux.Vars(req)
	entryID := vars["entry"]
	if entryID == "" {
		c.lc.Error("EntryID is empty")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Please enter a valid audit log entry ID in the form of /auditlog/{entryId}"))
		return
	}
	// if the user wants to delete all inventory, do it
	if entryID == DeleteAllQueryString {
		err := c.DeleteAuditLog()
		if err != nil {
			c.lc.Errorf("Failed to reset audit log: %s", err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte("Failed to properly reset audit log: " + err.Error()))
			return
		}
		emptyAuditLogResponseJSON, err := json.Marshal(AuditLog{Data: []AuditLogEntry{}})
		if err != nil {
			c.lc.Errorf("Failed to serialize empty audit log response: %s", err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(fmt.Sprintf("Failed to serialize empty audit log response: %v", err.Error())))
		}
		c.lc.Info("Successfully deleted audit log")
		writer.Write(emptyAuditLogResponseJSON)
		return
	}
	// look up the requested audit log entry by EntryID
	auditLogEntryToDelete, auditLog, err := c.GetAuditLogEntryByID(entryID)
	c.auditLog = auditLog
	if err != nil {
		c.lc.Errorf("Failed to get audit log entry ID: %s with error: %s", entryID, err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Failed to get requested audit log entry by ID: " + err.Error()))
		return
	}

	// check if GetAuditLogEntryByID found an audit log entry to delete
	if auditLogEntryToDelete.AuditEntryID == "" {
		c.lc.Errorf("Item with entry ID: %s does not exist", auditLogEntryToDelete.AuditEntryID)
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte("Item does not exist"))
		return
	}
	// delete the audit log entry & write the modified audit log
	c.DeleteAuditLogEntry(auditLogEntryToDelete)
	err = c.WriteAuditLog()
	if err != nil {
		c.lc.Error("Failed to write updated audit log")
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Failed to write updated audit log"))
		return
	}
	auditLogEntryToDeleteJSON, err := json.Marshal(auditLogEntryToDelete)
	if err != nil {
		c.lc.Errorf("Successfully deleted item: %s from audit log, but failed to serialize information back to the requester: %s", auditLogEntryToDelete.AuditEntryID, err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf("Successfully deleted item from audit log, but failed to serialize information back to the requester: %v", err.Error())))
	}
	c.lc.Infof("Succssfully deleted item: %s from audit log", auditLogEntryToDelete.AuditEntryID)
	writer.Write(auditLogEntryToDeleteJSON)
}
