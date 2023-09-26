// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
	"github.com/stretchr/testify/require"
)

const (
	AuditLogFileName  = "test-auditlog.json"
	InventoryFileName = "test-inventory.json"
)

func getDefaultProductsList() Products {
	return Products{
		Data: []Product{{
			CreatedAt:          1567787309,
			IsActive:           true,
			ItemPrice:          1.99,
			MaxRestockingLevel: 24,
			MinRestockingLevel: 0,
			ProductName:        "Sprite (Lemon-Lime) - 16.9 oz",
			SKU:                "4900002470",
			UnitsOnHand:        0,
			UpdatedAt:          1567787309,
		}, {
			CreatedAt:          1567787309,
			IsActive:           true,
			ItemPrice:          1.99,
			MaxRestockingLevel: 18,
			MinRestockingLevel: 0,
			ProductName:        "Mountain Dew (Low Calorie) - 16.9 oz",
			SKU:                "1200010735",
			UnitsOnHand:        0,
			UpdatedAt:          1567787309,
		}, {
			CreatedAt:          1567787309,
			IsActive:           true,
			ItemPrice:          1.99,
			MaxRestockingLevel: 6,
			MinRestockingLevel: 0,
			ProductName:        "Mountain Dew - 16.9 oz",
			SKU:                "1200050408",
			UnitsOnHand:        0,
			UpdatedAt:          1567787309,
		}}}
}

// TestWriteInventory tests the ability to write inventory items
// related functions
func TestWriteInventory(t *testing.T) {
	// Product slice
	products := getDefaultProductsList()
	c := Controller{
		lc:                logger.NewMockClient(),
		service:           nil,
		inventoryItems:    products,
		inventoryFileName: InventoryFileName,
	}
	err := c.WriteInventory()
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(c.inventoryFileName)
	}()

	// load product from file to validate
	productsFromFile := Products{}
	err = utilities.LoadFromJSONFile(c.inventoryFileName, &productsFromFile)
	require.NoError(t, err)

	// Check to make sure items match
	require.Equal(t, products, productsFromFile, "Products should match")
}

// TestGetInventoryItems tests the ability to get all inventory items
// related functions
func TestGetInventoryItems(t *testing.T) {
	// Product slice
	products := getDefaultProductsList()
	c := Controller{
		lc:                logger.NewMockClient(),
		service:           nil,
		inventoryItems:    products,
		inventoryFileName: InventoryFileName,
	}
	err := c.WriteInventory()
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(c.inventoryFileName)
	}()

	// run GetInventoryItems and get the result as JSON
	productsFromFile, err := c.GetInventoryItems()
	require.NoError(t, err)

	// Check to make sure items do not match
	require.Equal(t, products, productsFromFile, "Products should match")
}

// TestGetInventoryItemBySKU tests the ability to get inventory item based on SKU
// related functions
func TestGetInventoryItemBySKU(t *testing.T) {
	// Product slice
	products := getDefaultProductsList()
	c := Controller{
		lc:                logger.NewMockClient(),
		service:           nil,
		inventoryItems:    products,
		inventoryFileName: InventoryFileName,
	}
	// variables
	missingSKUToReturn := "0000000000"

	tests := []struct {
		Name         string
		InventorySKU string
		FoundProduct bool
	}{
		{"valid SKU", products.Data[0].SKU, true},
		{"missing SKU", missingSKUToReturn, false},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			err := c.DeleteInventory()
			require.NoError(t, err)

			err = c.WriteInventory()
			require.NoError(t, err)
			defer func() {
				_ = os.Remove(c.inventoryFileName)
			}()

			// run GetInventoryItems and get the result as JSON
			productBySKU, productsFromFile, err := c.GetInventoryItemBySKU(currentTest.InventorySKU)
			require.NoError(t, err)

			for _, product := range productsFromFile.Data {
				if product.SKU == currentTest.InventorySKU {
					if currentTest.FoundProduct {
						require.Equal(t, productBySKU, product, "Expected products to match")
					} else {
						require.NotEqual(t, productBySKU, product, "Expected no products to match")
					}
				}
			}
		})
	}
}

// TestDeleteInventoryItem tests the ability to delete inventory items
// related functions
func TestDeleteInventoryItem(t *testing.T) {
	// Product slice
	products := getDefaultProductsList()
	c := Controller{
		lc:                logger.NewMockClient(),
		service:           nil,
		inventoryItems:    products,
		inventoryFileName: InventoryFileName,
	}
	err := c.WriteInventory()
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(c.inventoryFileName)
	}()

	deletedProductID := products.Data[0].SKU
	c.DeleteInventoryItem(products.Data[0])

	for _, product := range products.Data {
		if product.SKU == deletedProductID {
			t.Fatalf("Deleted person with ID " + (product.SKU) + " but it still exists in the test list")
		}
	}
}

// TestGetInventoryItemErrors tests the error checking on the GetInventoryItems function
// related functions
func TestGetInventoryItemErrors(t *testing.T) {
	// Product slice
	products := getDefaultProductsList()
	c := Controller{
		lc:                logger.NewMockClient(),
		service:           nil,
		inventoryItems:    products,
		inventoryFileName: InventoryFileName,
	}
	err := c.WriteInventory()
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(c.inventoryFileName)
	}()

	t.Run("Test GetInventoryItems Error", func(t *testing.T) {
		err := ioutil.WriteFile(c.inventoryFileName, []byte("invalid json test"), 0644)
		require.NoError(t, err)

		_, err = c.GetInventoryItems()
		require.NotNil(t, err, "Expected inventory get to fail")
	})
	t.Run("Test GetInventoryItemBySKU Error", func(t *testing.T) {
		err := ioutil.WriteFile(c.inventoryFileName, []byte("invalid json test"), 0644)
		require.NoError(t, err)

		_, _, err = c.GetInventoryItemBySKU(products.Data[0].SKU)
		require.NotNil(t, err, "Expected inventory get to fail")
	})
	t.Run("Test DeleteInventory", func(t *testing.T) {
		err := c.DeleteInventory()
		require.NoError(t, err)

		productsFromFile, err := c.GetInventoryItems()
		require.NoError(t, err)
		require.LessOrEqual(t, len(productsFromFile.Data), 0, "Expected products list to be empty but it contained 1 or more entry")
	})
}

func getDefaultAuditsList() AuditLog {
	return AuditLog{
		Data: []AuditLogEntry{{
			CardID:    "0003292356",
			AccountID: 1,
			RoleID:    1,
			PersonID:  1,
			InventoryDelta: []DeltaInventorySKU{
				{
					SKU:   "4900002470",
					Delta: 5,
				}, {
					SKU:   "1200010735",
					Delta: 5,
				},
			},
			CreatedAt:    1567787309,
			AuditEntryID: "1",
		}, {
			CardID:    "0003292371",
			AccountID: 2,
			RoleID:    2,
			PersonID:  2,
			InventoryDelta: []DeltaInventorySKU{
				{
					SKU:   "4900002470",
					Delta: -1,
				}, {
					SKU:   "1200010735",
					Delta: -1,
				},
			},
			CreatedAt:    1567787309,
			AuditEntryID: "2",
		}, {
			CardID:    "0003621873",
			AccountID: 3,
			RoleID:    3,
			PersonID:  3,
			InventoryDelta: []DeltaInventorySKU{
				{
					SKU:   "4900002470",
					Delta: -2,
				}, {
					SKU:   "1200010735",
					Delta: -2,
				},
			},
			CreatedAt:    1567787309,
			AuditEntryID: "3",
		}}}
}

// TestWriteAuditLog tests the ability to get all audit logs
// related functions
func TestWriteAuditLog(t *testing.T) {
	// Audit slice
	audits := getDefaultAuditsList()
	c := Controller{
		lc:               logger.NewMockClient(),
		service:          nil,
		auditLog:         audits,
		auditLogFileName: AuditLogFileName,
	}
	err := c.WriteAuditLog()
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(c.auditLogFileName)
	}()

	// load audits from file to validate
	auditsFromFile := AuditLog{}
	err = utilities.LoadFromJSONFile(AuditLogFileName, &auditsFromFile)
	require.NoError(t, err)

	// Check to make sure audit entries match
	require.Equal(t, audits, auditsFromFile, "Audits should match")
}

// TestGetAuditLog tests the ability to get all inventory items
// related functions
func TestGetAuditLog(t *testing.T) {
	// Audit slice
	audits := getDefaultAuditsList()
	c := Controller{
		lc:               logger.NewMockClient(),
		service:          nil,
		auditLog:         audits,
		auditLogFileName: AuditLogFileName,
	}
	err := c.WriteAuditLog()
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(c.auditLogFileName)
	}()

	// run GetInventoryItems and get the result as JSON
	auditsFromFile, err := c.GetAuditLog()
	require.NoError(t, err)

	// Check to make sure audit entries match
	require.Equal(t, audits, auditsFromFile, "Audits should match")
}

// TestGetAuditLogEntryByID tests the ability to get all audit logs
// related functions
func TestGetAuditLogEntryByID(t *testing.T) {
	// Audit slice
	audits := getDefaultAuditsList()
	c := Controller{
		lc:               logger.NewMockClient(),
		service:          nil,
		auditLog:         audits,
		auditLogFileName: AuditLogFileName,
	}
	err := c.WriteAuditLog()
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(c.auditLogFileName)
	}()

	// variables
	entryIDToReturn := "1"
	invalidEntryIDToReturn := "100"

	t.Run("Test GetAuditLogEntryByID", func(t *testing.T) {
		// run GetAuditLogEntryByID and get the result as JSON
		auditLogEntry, auditsFromFile, err := c.GetAuditLogEntryByID(entryIDToReturn)
		require.NoError(t, err)

		// Check to make sure audit entries match
		require.Equal(t, audits, auditsFromFile, "Audits should match")

		foundAudit := false
		for _, audit := range auditsFromFile.Data {
			if audit.AuditEntryID == entryIDToReturn {
				if auditLogEntry.AuditEntryID == audit.AuditEntryID {
					foundAudit = true
					// Check to make sure audit entries match
				}
			}
		}
		require.True(t, foundAudit, "Did not find expected audit")
	})
	t.Run("Test GetAuditLogEntryByID invalid entry ID", func(t *testing.T) {
		// run GetAuditLogEntryByID and get the result as JSON
		auditLogEntry, auditsFromFile, err := c.GetAuditLogEntryByID(invalidEntryIDToReturn)
		require.NoError(t, err)

		// Check to make sure audit entries match
		require.Equal(t, audits, auditsFromFile, "Audits should match")

		foundAudit := false
		for _, audit := range auditsFromFile.Data {
			if audit.AuditEntryID == entryIDToReturn {
				if auditLogEntry.AuditEntryID == audit.AuditEntryID {
					foundAudit = true
					// Check to make sure audit entries match
				}
			}
		}
		require.False(t, foundAudit, "Found unexpected audit")
	})
}

// TestDeleteAuditLogEntry tests the ability to get all audit logs
// related functions
func TestDeleteAuditLogEntry(t *testing.T) {
	// Audit slice
	audits := getDefaultAuditsList()
	c := Controller{
		lc:               logger.NewMockClient(),
		service:          nil,
		auditLog:         audits,
		auditLogFileName: AuditLogFileName,
	}
	err := c.WriteAuditLog()
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(c.auditLogFileName)
	}()

	deletedAuditID := audits.Data[0].AuditEntryID
	c.DeleteAuditLogEntry(audits.Data[0])

	for _, audit := range audits.Data {
		if audit.AuditEntryID == deletedAuditID {
			t.Fatalf("Deleted person with ID " + (audit.AuditEntryID) + " but it still exists in the test list")
		}
	}
}

// TestGetAuditLogErrors tests the ability to get all audit logs
// related functions
func TestGetAuditLogErrors(t *testing.T) {
	// Audit slice
	audits := getDefaultAuditsList()
	c := Controller{
		lc:               logger.NewMockClient(),
		service:          nil,
		auditLog:         audits,
		auditLogFileName: AuditLogFileName,
	}
	err := c.WriteAuditLog()
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(c.auditLogFileName)
	}()

	// variables
	entryIDToReturn := "1"

	t.Run("Test GetAuditLog Error", func(t *testing.T) {
		err := ioutil.WriteFile(AuditLogFileName, []byte("invalid json test"), 0644)
		require.NoError(t, err)

		_, err = c.GetAuditLog()
		require.NotNil(t, err, "Expected audit log get to fail")
	})
	t.Run("Test GetAuditLogEntryByID Error", func(t *testing.T) {
		err := ioutil.WriteFile(AuditLogFileName, []byte("invalid json test"), 0644)
		require.NoError(t, err)

		_, _, err = c.GetAuditLogEntryByID(entryIDToReturn)
		require.NotNil(t, err, "Expected audit log get to fail")
	})
	t.Run("Test DeleteAuditLog", func(t *testing.T) {
		err := c.DeleteAuditLog()
		require.NoError(t, err)

		auditsFromFile, err := c.GetAuditLog()
		require.NoError(t, err)
		require.LessOrEqual(t, len(auditsFromFile.Data), 0, "Expected audits list to be empty but it contained 1 or more entry")
	})
}
