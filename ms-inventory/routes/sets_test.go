// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/stretchr/testify/require"
)

// TestInventoryPost tests the function InventoryPost
func TestInventoryPost(t *testing.T) {
	// Product slice

	products := getDefaultProductsList()
	c := Controller{
		lc:             logger.NewMockClient(),
		service:        nil,
		inventoryItems: products,
	}

	tests := []struct {
		Name                string
		BadInventory        bool
		ProductUpdateString string
		ExpectedStatusCode  int
		ProductsMatch       bool
	}{
		{"modify first inventory item price", false, `[{"sku": "4900002470","itemPrice": 10.5,"unitsOnHand": 2,"maxRestockingLevel": 9,"minRestockingLevel": 1,"isActive": false}]`, http.StatusOK, false},
		{"add new inventory item", false, `[{"sku": "9999999999","itemPrice": 10.5,"unitsOnHand": 2,"maxRestockingLevel": 9,"minRestockingLevel": 1,"isActive": false}]`, http.StatusOK, false},
		{"add new inventory item with default items", false, `[{"sku": "8888888888","isActive": false}]`, http.StatusOK, false},
		{"modify inventory item with strings instead of float values", false, `[{"sku": "7777777777","itemPrice": "zero","unitsOnHand": "zero","maxRestockingLevel": "zero","minRestockingLevel": "zero","isActive": false}]`, http.StatusOK, false},
		{"reduce inventory below 0", false, `[{"sku": "4900002470","itemPrice": 10.5,"unitsOnHand": -10,"maxRestockingLevel": 9,"minRestockingLevel": 1,"isActive": false}]`, http.StatusOK, false},
		{"raise inventory above max threshold", false, `[{"sku": "4900002470","itemPrice": 10.5,"unitsOnHand": 20,"maxRestockingLevel": 9,"minRestockingLevel": 1,"isActive": false}]`, http.StatusOK, false},
		{"invalid inventory item", false, `invalid item`, http.StatusBadRequest, true},
		{"invalid inventory item", true, `[{"sku": "4900002470","itemPrice": 10.5,"unitsOnHand": 2,"maxRestockingLevel": 9,"minRestockingLevel": 1,"isActive": false}]`, http.StatusInternalServerError, true},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			err := c.DeleteInventory()
			require.NoError(t, err)

			if currentTest.BadInventory {
				err := ioutil.WriteFile(InventoryFileName, []byte("invalid json test"), 0644)
				require.NoError(t, err)
			} else {
				err := c.WriteInventory()
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "http://localhost:48096/inventory", bytes.NewBuffer([]byte(currentTest.ProductUpdateString)))
			w := httptest.NewRecorder()
			req.Header.Set("Content-Type", "application/json")
			c.InventoryPost(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")

			if !test.BadInventory {
				// run GetInventoryItems and get the result as JSON
				productsFromFile, err := c.GetInventoryItems()
				require.NoError(t, err)

				if currentTest.ProductsMatch {
					require.Equal(t, products, productsFromFile, "Products should match")
				} else {
					require.NotEqual(t, products, productsFromFile, "Products should not match")
				}
			}
		})
	}
}

// TestAuditLogPost tests the function AuditLogPost
// related functions
func TestAuditLogPost(t *testing.T) {
	// Audit slice
	audits := getDefaultAuditsList()
	c := Controller{
		lc:       logger.NewMockClient(),
		service:  nil,
		auditLog: audits,
	}
	tests := []struct {
		Name               string
		BadAuditLog        bool
		AuditLogUpdate     string
		ExpectedStatusCode int
		AuditsMatch        bool
	}{
		{"Test AuditLogPost add new audit log entry with no CreatdAt field", false, `{"CardID":"0003292356","AccountID":1,"RoleID":1,"PersonID":1,"InventoryDelta":[{"SKU":"4900002470","Delta": 5},{"SKU":"1200010735","Delta": 5}]}`, http.StatusOK, false},
		{"Test AuditLogPost with an entry id", false, `{"CardID":"0003292356","AccountID":1,"RoleID":1,"PersonID":1,"AuditEntryID":"1","InventoryDelta":[{"SKU":"4900002470","Delta": 5},{"SKU":"1200010735","Delta": 5}]}`, http.StatusBadRequest, true},
		{"Test InventoryPost invalid audit log item", false, "This is an invalid string", http.StatusBadRequest, true},
		{"Test InventoryPost invalid audit log list json", true, `{"CardID":"0003292356","AccountID":1,"RoleID":1,"PersonID":1,"InventoryDelta":[{"SKU":"4900002470","Delta": 5},{"SKU":"1200010735","Delta": 5}]}`, http.StatusInternalServerError, false},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			err := c.DeleteAuditLog()
			require.NoError(t, err)

			if currentTest.BadAuditLog {
				err := ioutil.WriteFile(AuditLogFileName, []byte("invalid json test"), 0644)
				require.NoError(t, err)
			} else {
				err := c.WriteAuditLog()
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "http://localhost:48096/auditlog", bytes.NewBuffer([]byte(currentTest.AuditLogUpdate)))
			w := httptest.NewRecorder()
			req.Header.Set("Content-Type", "application/json")
			c.AuditLogPost(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")

			if !currentTest.BadAuditLog {
				// run GetAuditLog and get the result as JSON
				auditsFromFile, err := c.GetAuditLog()
				require.NoError(t, err)

				if currentTest.AuditsMatch {
					require.Equal(t, audits, auditsFromFile, "Audits should match")
				} else {
					require.NotEqual(t, audits, auditsFromFile, "Audits should not match")
				}
			}
		})
	}
}

// TestDeltaInventorySKUPost tests the function DeltaInventorySKUPost
func TestDeltaInventorySKUPost(t *testing.T) {
	// Product slice
	products := Products{
		Data: []Product{{
			CreatedAt:          1567787309,
			IsActive:           true,
			ItemPrice:          1.99,
			MaxRestockingLevel: 24,
			MinRestockingLevel: 0,
			ProductName:        "Sprite (Lemon-Lime) - 16.9 oz",
			SKU:                "4900002470",
			UnitsOnHand:        5,
			UpdatedAt:          1567787309,
		}, {
			CreatedAt:          1567787309,
			IsActive:           true,
			ItemPrice:          1.99,
			MaxRestockingLevel: 18,
			MinRestockingLevel: 0,
			ProductName:        "Mountain Dew (Low Calorie) - 16.9 oz",
			SKU:                "1200010735",
			UnitsOnHand:        5,
			UpdatedAt:          1567787309,
		}, {
			CreatedAt:          1567787309,
			IsActive:           true,
			ItemPrice:          1.99,
			MaxRestockingLevel: 6,
			MinRestockingLevel: 0,
			ProductName:        "Mountain Dew - 16.9 oz",
			SKU:                "1200050408",
			UnitsOnHand:        5,
			UpdatedAt:          1567787309,
		}}}
	c := Controller{
		lc:             logger.NewMockClient(),
		service:        nil,
		inventoryItems: products,
	}

	tests := []struct {
		Name               string
		BadInventory       bool
		DeltaUpdateString  string
		ExpectedStatusCode int
		ProductsMatch      bool
	}{
		{"subtracting 1 item from existing SKU", false, `[{"SKU": "4900002470","Delta": -1}]`, http.StatusOK, false},
		{"missing SKU and no delta change", false, `[{"SKU": "0000000000","Delta": 0}]`, http.StatusNotModified, true},
		{"invalid delta json", false, `This is an invalid string`, http.StatusBadRequest, true},
		{"subtracting 1 item from existing SKU with invalid inventory", true, `[{"SKU": "4900002470","Delta": -1}]`, http.StatusInternalServerError, false},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			err := c.DeleteInventory()
			require.NoError(t, err)

			if currentTest.BadInventory {
				err := ioutil.WriteFile(InventoryFileName, []byte("invalid json test"), 0644)
				require.NoError(t, err)
			} else {
				err := c.WriteInventory()
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "http://localhost:48096/inventory/delta", bytes.NewBuffer([]byte(currentTest.DeltaUpdateString)))
			w := httptest.NewRecorder()
			req.Header.Set("Content-Type", "application/json")
			c.DeltaInventorySKUPost(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")

			if !currentTest.BadInventory {
				// run GetInventoryItems and get the result as JSON
				productsFromFile, err := c.GetInventoryItems()
				require.NoError(t, err)

				if currentTest.ProductsMatch {
					require.Equal(t, products, productsFromFile, "Products should match")
				} else {
					require.NotEqual(t, products, productsFromFile, "Products should not match")
				}
			}
		})
	}
}
