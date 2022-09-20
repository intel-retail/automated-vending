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
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// TestInventoryDelete tests the function InventoryDelete
func TestInventoryDelete(t *testing.T) {
	// Product slice
	products := getDefaultProductsList()

	tests := []struct {
		Name               string
		BadInventory       bool
		InventorySKU       string
		ExpectedStatusCode int
		ProductsMatch      bool
		InventoryPath      string
	}{
		{"with valid SKU", false, products.Data[0].SKU, http.StatusOK, false, "inventory.json"},
		{"with invalid SKU", false, "0000000000", http.StatusNotFound, true, "inventory.json"},
		{"with missing SKU", false, "", http.StatusBadRequest, true, "inventory.json"},
		{"with all parameter", false, "all", http.StatusOK, false, "inventory.json"},
		{"with invalid inventory json", true, products.Data[0].SKU, http.StatusInternalServerError, true, "inventory.json"},
		{"with all parameter and invalid inventory json path", false, "all", http.StatusInternalServerError, true, "tests/inventory.json"},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			products := getDefaultProductsList()

			c := Controller{
				lc:             logger.NewMockClient(),
				service:        nil,
				inventoryItems: products,
			}
			err := c.DeleteInventory()
			require.NoError(t, err)

			if currentTest.BadInventory {
				err := ioutil.WriteFile(InventoryFileName, []byte("invalid json test"), 0644)
				require.NoError(t, err)
			} else {
				err := c.WriteInventory()
				require.NoError(t, err)
			}
			InventoryFileName = currentTest.InventoryPath

			req := httptest.NewRequest("DELETE", "http://localhost:48096/inventory/", bytes.NewBuffer([]byte(currentTest.InventorySKU)))
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"sku": currentTest.InventorySKU})
			req.Header.Set("Content-Type", "application/json")
			c.InventoryDelete(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")

			InventoryFileName = "inventory.json"
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

// TestAuditLogDelete tests the function InventoryDelete
func TestAuditLogDelete(t *testing.T) {
	// Audit slice
	audits := getDefaultAuditsList()

	tests := []struct {
		Name               string
		BadAuditID         bool
		AuditEntryID       string
		ExpectedStatusCode int
		ProductsMatch      bool
		AuditLogPath       string
	}{
		{"with valid Entry ID", false, audits.Data[0].AuditEntryID, http.StatusOK, false, "auditlog.json"},
		{"with invalid Entry ID", false, "0000000000", http.StatusNotFound, true, "auditlog.json"},
		{"with missing Entry ID", false, "", http.StatusBadRequest, true, "auditlog.json"},
		{"with all parameter", false, "all", http.StatusOK, false, "auditlog.json"},
		{"with invalid auditlog json", true, audits.Data[0].AuditEntryID, http.StatusInternalServerError, true, "auditlog.json"},
		{"with all parameter and invalid auditlog json path", false, "all", http.StatusInternalServerError, true, "tests/auditlog.json"},
	}

	for _, test := range tests {
		currentTest := test
		audits := getDefaultAuditsList()
		c := Controller{
			lc:       logger.NewMockClient(),
			service:  nil,
			auditLog: audits,
		}
		t.Run(currentTest.Name, func(t *testing.T) {
			err := c.DeleteAuditLog()
			require.NoError(t, err)

			if currentTest.BadAuditID {
				err := ioutil.WriteFile(AuditLogFileName, []byte("invalid json test"), 0644)
				require.NoError(t, err)
			} else {
				err := c.WriteAuditLog()
				require.NoError(t, err)
			}
			AuditLogFileName = currentTest.AuditLogPath

			req := httptest.NewRequest("DELETE", "http://localhost:48096/auditlog/", bytes.NewBuffer([]byte(currentTest.AuditEntryID)))
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"entry": currentTest.AuditEntryID})
			req.Header.Set("Content-Type", "application/json")
			c.AuditLogDelete(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")

			AuditLogFileName = "auditlog.json"
			if !currentTest.BadAuditID {
				// run GetAuditLog and get the result as JSON
				auditsFromFile, err := c.GetAuditLog()
				require.NoError(t, err)

				if currentTest.ProductsMatch {
					require.Equal(t, audits, auditsFromFile, "Products should match")
				} else {
					require.NotEqual(t, audits, auditsFromFile, "Products should not match")
				}
			}
		})
	}
}
