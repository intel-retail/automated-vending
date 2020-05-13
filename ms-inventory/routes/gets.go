// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

// InventoryGet allows for the retrieval of the entire inventory
func InventoryGet(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {
		inventoryItems, err := GetInventoryItems()
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to retrieve all inventory items: "+err.Error(), true)
			return
		}

		// No logic needs to be done here, since we are just reading the file
		// and writing it back out. Simply marshaling it will validate its structure
		inventoryItemsJSON, err := utilities.GetAsJSON(inventoryItems)
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to process inventory items: "+err.Error(), true)
			return
		}

		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, inventoryItemsJSON, false)
	})
}

// InventoryItemGet allows for a single inventory item to be retrieved by SKU
func InventoryItemGet(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		sku := vars["sku"]
		if sku != "" {
			inventoryItem, _, err := GetInventoryItemBySKU(sku)
			if err != nil {
				utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to get inventory item by SKU: "+err.Error(), true)
				return
			}
			if inventoryItem.SKU == "" {
				utilities.WriteStringHTTPResponse(writer, req, http.StatusNotFound, "", false)
				return
			}
			outputInventoryItemJSON, err := utilities.GetAsJSON(inventoryItem)
			if err != nil {
				utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to process the requested inventory item "+sku+":"+err.Error(), true)
				return
			}
			utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, outputInventoryItemJSON, false)
			return
		}

		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Please enter a valid inventory item in the form of /inventory/{sku}", false)
	})
}

// AuditLogGetAll allows all audit log entries to be retrieved
func AuditLogGetAll(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {
		auditLog, err := GetAuditLog()
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to retrieve all audit log entries: "+err.Error(), true)
			return
		}

		// No logic needs to be done here, since we are just reading the file
		// and writing it back out. Simply marshaling it will validate its structure
		auditLogJSON, err := utilities.GetAsJSON(auditLog)
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to process audit log entries: "+err.Error(), true)
			return
		}

		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, auditLogJSON, false)
	})
}

// AuditLogGetEntry allows a single audit log entry to be retrieved by its
// UUID
func AuditLogGetEntry(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		entryID := vars["entry"]
		if entryID != "" {
			auditLogEntry, _, err := GetAuditLogEntryByID(entryID)
			if err != nil {
				utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to get audit log entry by ID: "+err.Error(), true)
				return
			}
			if auditLogEntry.AuditEntryID == "" {
				utilities.WriteStringHTTPResponse(writer, req, http.StatusNotFound, "", false)
				return
			}
			outputAuditLogEntryJSON, err := utilities.GetAsJSON(auditLogEntry)
			if err != nil {
				utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to process the requested audit log entry item "+entryID+":"+err.Error(), true)
				return
			}
			utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, outputAuditLogEntryJSON, false)
			return
		}

		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Please enter a valid entry ID in the form of /auditlog/{entry}", false)
	})
}
