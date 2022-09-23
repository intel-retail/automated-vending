// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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
		{"with valid SKU", false, products.Data[0].SKU, http.StatusOK, false, InventoryFileName},
		{"with invalid SKU", false, "0000000000", http.StatusNotFound, true, InventoryFileName},
		{"with missing SKU", false, "", http.StatusBadRequest, true, InventoryFileName},
		{"with all parameter", false, "all", http.StatusOK, false, InventoryFileName},
		{"with invalid inventory json", true, products.Data[0].SKU, http.StatusInternalServerError, true, InventoryFileName},
		{"with all parameter and invalid inventory json path", false, "all", http.StatusInternalServerError, true, "tests/inventory.json"},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			products := getDefaultProductsList()

			c := Controller{
				lc:                logger.NewMockClient(),
				service:           nil,
				inventoryItems:    products,
				inventoryFileName: InventoryFileName,
			}
			err := c.DeleteInventory()
			require.NoError(t, err)
			defer func() {
				_ = os.Remove(c.inventoryFileName)
			}()

			if currentTest.BadInventory {
				err := ioutil.WriteFile(c.inventoryFileName, []byte("invalid json test"), 0644)
				require.NoError(t, err)
			} else {
				err := c.WriteInventory()
				require.NoError(t, err)
			}
			c.inventoryFileName = currentTest.InventoryPath

			req := httptest.NewRequest("DELETE", "http://localhost:48096/inventory/", bytes.NewBuffer([]byte(currentTest.InventorySKU)))
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"sku": currentTest.InventorySKU})
			req.Header.Set("Content-Type", "application/json")
			c.InventoryDelete(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")

			c.inventoryFileName = InventoryFileName
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
		{"with valid Entry ID", false, audits.Data[0].AuditEntryID, http.StatusOK, false, AuditLogFileName},
		{"with invalid Entry ID", false, "0000000000", http.StatusNotFound, true, AuditLogFileName},
		{"with missing Entry ID", false, "", http.StatusBadRequest, true, AuditLogFileName},
		{"with all parameter", false, "all", http.StatusOK, false, AuditLogFileName},
		{"with invalid auditlog json", true, audits.Data[0].AuditEntryID, http.StatusInternalServerError, true, AuditLogFileName},
		{"with all parameter and invalid auditlog json path", false, "all", http.StatusInternalServerError, true, "tests/auditlog.json"},
	}

	for _, test := range tests {
		currentTest := test
		audits := getDefaultAuditsList()
		c := Controller{
			lc:               logger.NewMockClient(),
			service:          nil,
			auditLog:         audits,
			auditLogFileName: AuditLogFileName,
		}
		t.Run(currentTest.Name, func(t *testing.T) {
			err := c.DeleteAuditLog()
			require.NoError(t, err)

			if currentTest.BadAuditID {
				err := ioutil.WriteFile(c.auditLogFileName, []byte("invalid json test"), 0644)
				require.NoError(t, err)
			} else {
				err := c.WriteAuditLog()
				require.NoError(t, err)
			}
			defer func() {
				_ = os.Remove(c.auditLogFileName)
			}()
			c.auditLogFileName = currentTest.AuditLogPath

			req := httptest.NewRequest("DELETE", "http://localhost:48096/auditlog/", bytes.NewBuffer([]byte(currentTest.AuditEntryID)))
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"entry": currentTest.AuditEntryID})
			req.Header.Set("Content-Type", "application/json")
			c.AuditLogDelete(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")

			c.auditLogFileName = AuditLogFileName
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
