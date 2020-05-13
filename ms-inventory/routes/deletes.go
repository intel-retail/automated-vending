// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"net/http"
	"fmt"

	"github.com/gorilla/mux"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

// InventoryDelete allows deletion of an inventory item or multiple items
func InventoryDelete(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {
		// find the requested SKU and exit if it's invalid
		vars := mux.Vars(req)
		SKU := vars["sku"]
		if SKU == "" {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Please enter a valid inventory item in the form of /inventory/{sku}", true)
			return
		}
		// if the user wants to delete all inventory, do it
		if SKU == DeleteAllQueryString {
			err := DeleteInventory()
			if err != nil {
				utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to properly reset inventory: "+err.Error(), true)
				return
			}
			emptyInventoryResponseJSON, err := utilities.GetAsJSON(Products{Data: []Product{}})
			if err != nil {
				utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to serialize empty inventory response: "+err.Error(), true)
				return
			}

			utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, emptyInventoryResponseJSON, false)
			return
		}
		// look up the requested inventory item by SKU
		inventoryItemToDelete, inventoryItems, err := GetInventoryItemBySKU(SKU)
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to get requested inventory item by SKU: "+err.Error(), true)
			return
		}

		// check if GetInventoryItemBySKU found an inventory item to delete
		if inventoryItemToDelete.SKU == "" {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusNotFound, "Item does not exist", false)
			return
		}
		// delete the inventory item & write the modified inventory
		inventoryItems.DeleteInventoryItem(inventoryItemToDelete)
		err = inventoryItems.WriteInventory()
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to write updated inventory", true)
			return
		}
		inventoryItemToDeleteJSON, err := utilities.GetAsJSON(inventoryItemToDelete)
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, fmt.Sprintf("Successfully deleted the item from inventory, but failed to serialize it so that it could be sent back to the requester: %v", err.Error()), true)
		}
		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, inventoryItemToDeleteJSON, false)
	})
}

// AuditLogDelete allows deletion of one or more audit log entry items
func AuditLogDelete(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {
		// find the requested SKU and exit if it's invalid
		vars := mux.Vars(req)
		entryID := vars["entry"]
		if entryID == "" {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Please enter a valid audit log entry ID in the form of /auditlog/{entryId}", true)
			return
		}
		// if the user wants to delete all inventory, do it
		if entryID == DeleteAllQueryString {
			err := DeleteAuditLog()
			if err != nil {
				utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to properly reset audit log: "+err.Error(), true)
				return
			}
			emptyAuditLogResponseJSON, err := utilities.GetAsJSON(AuditLog{Data: []AuditLogEntry{}})
			if err != nil {
				utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, fmt.Sprintf("Failed to serialize empty audit log response: %v", err.Error()), true)
			}
			utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, emptyAuditLogResponseJSON, false)
			return
		}
		// look up the requested audit log entry by EntryID
		auditLogEntryToDelete, auditLog, err := GetAuditLogEntryByID(entryID)
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to get requested audit log entry by ID: "+err.Error(), true)
			return
		}

		// check if GetAuditLogEntryByID found an audit log entry to delete
		if auditLogEntryToDelete.AuditEntryID == "" {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusNotFound, "Item does not exist", false)
			return
		}
		// delete the audit log entry & write the modified audit log
		auditLog.DeleteAuditLogEntry(auditLogEntryToDelete)
		err = auditLog.WriteAuditLog()
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to write updated audit log", true)
			return
		}
		auditLogEntryToDeleteJSON, err := utilities.GetAsJSON(auditLogEntryToDelete)
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, fmt.Sprintf("Successfully deleted item from audit log, but failed to serialize information back to the requester: %v", err.Error()), true)
		}
		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, auditLogEntryToDeleteJSON, false)
	})
}
