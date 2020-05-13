// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

// Products is the schema for the data that will be returned to the user
// when hitting the inventory endpoint
type Products struct {
	Data []Product `json:"data"`
}

// Product is the schema for a single inventory item
type Product struct {
	SKU                string  `json:"sku"`
	ItemPrice          float64 `json:"itemPrice"`
	ProductName        string  `json:"productName"`
	UnitsOnHand        int     `json:"unitsOnHand"`
	MaxRestockingLevel int     `json:"maxRestockingLevel"`
	MinRestockingLevel int     `json:"minRestockingLevel"`
	CreatedAt          int64   `json:"createdAt,string"`
	UpdatedAt          int64   `json:"updatedAt,string"`
	IsActive           bool    `json:"isActive"`
}

// DeltaInventorySKU is required because we cannot unmarshal a delta
// into Product struct, and the API endpoints needs to accept a delta
type DeltaInventorySKU struct {
	SKU   string `json:"SKU"`
	Delta int    `json:"delta"`
}

// AuditLog is similar to Products in that it is the schema for the data
// that will be returned to the user when hitting the audit log endpoint
type AuditLog struct {
	Data []AuditLogEntry `json:"data"`
}

// AuditLogEntry represents the schema for a single audit log entry
type AuditLogEntry struct {
	CardID         string              `json:"cardId"`
	AccountID      int                 `json:"accountId"`
	RoleID         int                 `json:"roleId"`
	PersonID       int                 `json:"personId"`
	InventoryDelta []DeltaInventorySKU `json:"inventoryDelta"`
	CreatedAt      int64               `json:"createdAt,string"`
	AuditEntryID   string              `json:"auditEntryId"`
}
