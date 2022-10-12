// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

// DeltaInventorySKUPost allows a change in inventory (a delta), via HTTP Post
// REST requests to occur
func (c *Controller) DeltaInventorySKUPost(writer http.ResponseWriter, req *http.Request) {
	response := utilities.GetHTTPResponseTemplate()

	// Read request body
	body := make([]byte, req.ContentLength)
	if _, err := io.ReadFull(req.Body, body); err != nil {
		c.lc.Errorf("Failed to process the posted delta inventory item(s): %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Failed to process the posted delta inventory item(s): "+err.Error(), true)
		return
	}

	// Unmarshal the string contents of request into a proper structure
	var deltaInventorySKUList []DeltaInventorySKU
	err := json.Unmarshal(body, &deltaInventorySKUList)
	if err != nil {
		c.lc.Errorf("Failed to process the posted delta inventory item(s): %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Failed to process the posted delta inventory item(s): "+err.Error(), true)
		return
	}

	// load the inventory.json file
	var inventoryItems Products
	err = utilities.LoadFromJSONFile(c.inventoryFileName, &inventoryItems)
	if err != nil {
		c.lc.Errorf("Failed to retrieve all inventory items: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to retrieve all inventory items: "+err.Error(), true)
		return
	}

	// iterate over all deltaInventorySKU's and find their corresponding SKU in inventory
	// then update the inventory with the delta
	var updatedInventoryItems []Product // will return the inventory items that got updated
	performedUpdate := false
	for _, deltaInventorySKU := range deltaInventorySKUList {
		for i, inventoryItem := range inventoryItems.Data {
			if deltaInventorySKU.SKU == inventoryItem.SKU {
				inventoryItems.Data[i].UnitsOnHand += deltaInventorySKU.Delta
				updatedInventoryItems = append(updatedInventoryItems, inventoryItems.Data[i])
				performedUpdate = true
				break
			}
		}
	}

	// Nothing was done, so return "Not Modified" status
	if !performedUpdate {
		c.lc.Info("No change made to inventory")
		utilities.WriteStringHTTPResponse(writer, req, http.StatusNotModified, "", false)
		return
	}

	// Write the updated inventory to the inventory json file
	err = utilities.WriteToJSONFile(c.inventoryFileName, inventoryItems, 0644)
	if err != nil {
		c.lc.Errorf("Failed to write inventory: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to write inventory: "+err.Error(), true)
		return
	}
	// return the new/updated items as JSON, or if for some reason it cannot be processed back into
	// JSON for returning to the user, fallback to a simple string
	updatedInventoryItemsJSON, err := utilities.GetAsJSON(updatedInventoryItems)
	if err != nil {
		c.lc.Info("Updated inventory successfully")
		response.SetStringHTTPResponseFields(http.StatusOK, "Updated inventory successfully", false)
	} else {
		c.lc.Infof("Updated inventory successfully: %s", updatedInventoryItemsJSON)
		response.SetJSONHTTPResponseFields(http.StatusOK, updatedInventoryItemsJSON, false)
	}
	response.WriteHTTPResponse(writer, req)
}

// InventoryPost allows new items to be added to inventory, as well as updating
// existing items
func (c *Controller) InventoryPost(writer http.ResponseWriter, req *http.Request) {
	// Use the ProcessCORS "decorator" function to reduce code repetition
	response := utilities.GetHTTPResponseTemplate()

	// Read request body
	body := make([]byte, req.ContentLength)
	if _, err := io.ReadFull(req.Body, body); err != nil {
		c.lc.Errorf("Failed to process the posted inventory item(s): %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Failed to process the posted inventory item(s): "+err.Error(), true)
		return
	}

	// deltaInventoryList is created as a map[string]interface{} because we are being
	// passed in _fields_ to update, not entire structs. If we attempt to use structs,
	// golang will automatically populate fields that were not modified by the user.
	// You can play with this using this Go playground URL:
	// https://play.golang.org/p/XJ0wiE629z8
	deltaInventoryList := make([]map[string]interface{}, 0)

	err := json.Unmarshal(body, &deltaInventoryList)
	if err != nil {
		c.lc.Errorf("Failed to process the posted inventory item(s): %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Failed to process the posted inventory item(s): "+err.Error(), true)
		return
	}

	// load the inventory.json file
	var inventoryItems Products
	err = utilities.LoadFromJSONFile(c.inventoryFileName, &inventoryItems)
	if err != nil {
		c.lc.Errorf("Failed to retrieve all inventory items: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to retrieve all inventory items: "+err.Error(), true)
		return
	}

	// Keep track of the items that get added so that the user can be informed of them in our response
	var newInventoryItems []Product

	// Loop through the posted inventory item list to find matching SKUs
	inventoryChanged := false
	for _, postedInventoryItem := range deltaInventoryList {
		postedInventoryItemFound := false
		for i := range inventoryItems.Data {
			// If the SKU matches update that item
			if postedInventoryItem["sku"] == inventoryItems.Data[i].SKU {
				postedInventoryItemFound = true
				if postedInventoryItem["itemPrice"] != nil {
					switch postedInventoryItem["itemPrice"].(type) {
					case float64:
						inventoryItems.Data[i].ItemPrice = postedInventoryItem["itemPrice"].(float64)
					}
				}
				if postedInventoryItem["maxRestockingLevel"] != nil {
					switch postedInventoryItem["maxRestockingLevel"].(type) {
					case float64:
						inventoryItems.Data[i].MaxRestockingLevel = int(postedInventoryItem["maxRestockingLevel"].(float64))
					}
				}
				if postedInventoryItem["minRestockingLevel"] != nil {
					switch postedInventoryItem["minRestockingLevel"].(type) {
					case float64:
						inventoryItems.Data[i].MinRestockingLevel = int(postedInventoryItem["minRestockingLevel"].(float64))
					}
				}
				if postedInventoryItem["isActive"] != nil {
					switch postedInventoryItem["isActive"].(type) {
					case bool:
						inventoryItems.Data[i].IsActive = postedInventoryItem["isActive"].(bool)
					}
				}
				if postedInventoryItem["unitsOnHand"] != nil {
					switch postedInventoryItem["unitsOnHand"].(type) {
					case float64:
						inventoryItems.Data[i].UnitsOnHand = inventoryItems.Data[i].UnitsOnHand + int(postedInventoryItem["unitsOnHand"].(float64))
					}

					// Need to send an error if the product units on hand is below 0
					if inventoryItems.Data[i].UnitsOnHand < 0 {
						c.lc.Infof("Product %s on hand is less than 0 which was caused by a bad delta value", postedInventoryItem["sku"])
					}
					// Item is under minimum stock level. Send notification
					if inventoryItems.Data[i].UnitsOnHand <= inventoryItems.Data[i].MinRestockingLevel {
						c.lc.Infof("Product %s needs to be restocked", postedInventoryItem["sku"])
					}
					// Item is under maximum stock level. Send notification
					if inventoryItems.Data[i].UnitsOnHand > inventoryItems.Data[i].MaxRestockingLevel {
						c.lc.Infof("Product %s is overstocked", postedInventoryItem["sku"])
					}
				}
				inventoryItems.Data[i].UpdatedAt = time.Now().UnixNano()
				inventoryChanged = true
				newInventoryItems = append(newInventoryItems, inventoryItems.Data[i])
				response.SetStringHTTPResponseFields(http.StatusOK, "Updated inventory", false)
			}
		}
		if !postedInventoryItemFound {
			newProduct := Product{
				SKU:       postedInventoryItem["sku"].(string),
				CreatedAt: time.Now().UnixNano(),
				UpdatedAt: time.Now().UnixNano(),
				IsActive:  true,
			}
			// Set the ItemPrice. If the ItemPrice isn't provided set a default value
			if postedInventoryItem["itemPrice"] != nil {
				switch postedInventoryItem["itemPrice"].(type) {
				case float64:
					newProduct.ItemPrice = postedInventoryItem["itemPrice"].(float64)
				default:
					newProduct.ItemPrice = 0
				}
			} else {
				newProduct.ItemPrice = 0
			}
			// Set the UnitsOnHand. If the UnitsOnHand isn't provided set a default value
			if postedInventoryItem["unitsOnHand"] != nil {
				switch postedInventoryItem["unitsOnHand"].(type) {
				case float64:
					newProduct.UnitsOnHand = int(postedInventoryItem["unitsOnHand"].(float64))
				default:
					newProduct.UnitsOnHand = 0
				}
			} else {
				newProduct.UnitsOnHand = 0
			}
			// Set the maxRestockingLevel. If the maxRestockingLevel isn't provided set a default value
			if postedInventoryItem["maxRestockingLevel"] != nil {
				switch postedInventoryItem["maxRestockingLevel"].(type) {
				case float64:
					newProduct.MaxRestockingLevel = int(postedInventoryItem["maxRestockingLevel"].(float64))
				default:
					newProduct.MaxRestockingLevel = 5
				}
			} else {
				newProduct.MaxRestockingLevel = 5
			}
			// Set the minRestockingLevel. If the minRestockingLevel isn't provided set a default value
			if postedInventoryItem["minRestockingLevel"] != nil {
				switch postedInventoryItem["minRestockingLevel"].(type) {
				case float64:
					newProduct.MinRestockingLevel = int(postedInventoryItem["minRestockingLevel"].(float64))
				default:
					newProduct.MinRestockingLevel = 0
				}
			} else {
				newProduct.MinRestockingLevel = 0
			}
			// Add new product to the product List
			inventoryItems.Data = append(inventoryItems.Data, newProduct)
			inventoryChanged = true
			newInventoryItems = append(newInventoryItems, newProduct)
			response.SetStringHTTPResponseFields(http.StatusOK, "Updated inventory", false)
		}
	}

	if inventoryChanged {
		// Write the updated audit log to the audit log json file
		err = utilities.WriteToJSONFile(c.inventoryFileName, inventoryItems, 0644)
		if err != nil {
			c.lc.Errorf("Failed to write inventory: %s", err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to write inventory: "+err.Error(), true)
			return
		}
		// return the new/updated items as JSON, or if for some reason it cannot be processed back into
		// JSON for returning to the user, fallback to a simple string
		newInventoryItemsJSON, err := utilities.GetAsJSON(newInventoryItems)
		if err != nil {
			c.lc.Info("Updated inventory successfully")
			response.SetStringHTTPResponseFields(http.StatusOK, "Updated inventory successfully", false)
		} else {
			c.lc.Infof("Updated inventory successfully: %s", newInventoryItemsJSON)
			response.SetJSONHTTPResponseFields(http.StatusOK, newInventoryItemsJSON, false)
		}
	}
	response.WriteHTTPResponse(writer, req)
}

// AuditLogPost allows for a new audit log entry to be added
func (c *Controller) AuditLogPost(writer http.ResponseWriter, req *http.Request) {
	// Use the ProcessCORS "decorator" function to reduce code repetition
	response := utilities.GetHTTPResponseTemplate()

	// Read request body
	body := make([]byte, req.ContentLength)
	if _, err := io.ReadFull(req.Body, body); err != nil {
		c.lc.Errorf("Failed to process the posted audit log entry: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Failed to process the posted audit log entry: "+err.Error(), true)
		return
	}

	// Unmarshal the string contents of request into an AuditLogEntry struct
	var postedAuditLogEntry AuditLogEntry
	err := json.Unmarshal(body, &postedAuditLogEntry)
	if err != nil {
		c.lc.Errorf("Failed to process the posted audit log entry: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Failed to process the posted audit log entry: "+err.Error(), true)
		return
	}

	// Check if an existing UUID has been specified,
	// and reject it if it has
	if postedAuditLogEntry.AuditEntryID != "" {
		c.lc.Error("The posted audit log entry is not defined")
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "The submitted audit log entry must not have an auditEntryId defined.", true)
		return
	}

	// assign a new UUID to our new audit log entry
	postedAuditLogEntry.AuditEntryID = utilities.GenUUID()

	// assign the CreatedAt value to right now, if the user hasn't passed it
	if postedAuditLogEntry.CreatedAt == 0 {
		postedAuditLogEntry.CreatedAt = time.Now().UnixNano()
	}

	// load the auditlog.json file
	var auditLog AuditLog
	err = utilities.LoadFromJSONFile(c.auditLogFileName, &auditLog)
	if err != nil {
		c.lc.Errorf("Failed to retrieve all audit log entries: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to retrieve all audit log entries: "+err.Error(), true)
		return
	}

	// the odds of matching a UUID are either:
	//   * nearly impossible, mathematically
	//   * high due to developer / user error
	// So we have to check for it.
	foundEntry := false
	for _, auditLogEntry := range auditLog.Data {
		if postedAuditLogEntry.AuditEntryID == auditLogEntry.AuditEntryID {
			foundEntry = true
			c.lc.Errorf("Failed to process the requested audit log entry: %s", postedAuditLogEntry.AuditEntryID)
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to process the requested audit log entry "+postedAuditLogEntry.AuditEntryID+" as it already exists", true)
			break
		}
	}
	// Happy path
	// If we haven't found any conflicting UUID's and there have been no errors,
	// proceed to update the list
	if !foundEntry && !response.Error {
		// write the result
		auditLog.Data = append(auditLog.Data, postedAuditLogEntry)

		err = utilities.WriteToJSONFile(c.auditLogFileName, auditLog, 0644)
		if err != nil {
			c.lc.Errorf("Failed to write the audit log entry: %s : %s", postedAuditLogEntry.AuditEntryID, err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to write the audit log entry "+postedAuditLogEntry.AuditEntryID+": "+err.Error(), true)
			return
		}

		// return the posted audit log entry to the user once added
		result, err := utilities.GetAsJSON(postedAuditLogEntry)
		if err != nil {
			c.lc.Errorf("Failed to return the requested audit log entry to the user: %s : %s", postedAuditLogEntry.AuditEntryID, err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to return the requested audit log entry to the user "+postedAuditLogEntry.AuditEntryID+": "+err.Error(), true)
			return
		}

		// Happy path HTTP response
		c.lc.Infof("Successfully added new entry to audit log: %s", postedAuditLogEntry.AuditEntryID)
		response.SetJSONHTTPResponseFields(http.StatusOK, result, false)
	}
	response.WriteHTTPResponse(writer, req)
}
