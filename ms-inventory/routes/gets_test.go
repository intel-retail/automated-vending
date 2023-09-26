// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/stretchr/testify/require"

	"github.com/gorilla/mux"
)

// TestInventoryGet tests the function InventoryGet
func TestInventoryGet(t *testing.T) {
	// Product slice
	products := getDefaultProductsList()
	c := Controller{
		lc:                logger.NewMockClient(),
		service:           nil,
		inventoryItems:    products,
		inventoryFileName: InventoryFileName,
	}
	tests := []struct {
		Name               string
		BadInventory       bool
		ExpectedStatusCode int
	}{
		{"InventoryGet", false, http.StatusOK},
		{"with invalid inventory json with", true, http.StatusInternalServerError},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			err := c.DeleteInventory()
			require.NoError(t, err)

			if currentTest.BadInventory {
				err := ioutil.WriteFile(c.inventoryFileName, []byte("invalid json test"), 0644)
				require.NoError(t, err)
			} else {
				err := c.WriteInventory()
				require.NoError(t, err)
			}
			defer func() {
				_ = os.Remove(c.inventoryFileName)
			}()

			req := httptest.NewRequest("GET", "http://localhost:48096/inventory", nil)
			w := httptest.NewRecorder()
			c.InventoryGet(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")
		})
	}
}

// TestInventoryItemGet tests the function InventoryGet
func TestInventoryItemGet(t *testing.T) {
	// Product slice
	products := getDefaultProductsList()
	c := Controller{
		lc:                logger.NewMockClient(),
		service:           nil,
		inventoryItems:    products,
		inventoryFileName: InventoryFileName,
	}
	tests := []struct {
		Name               string
		WriteInventory     bool
		BadInventory       bool
		URLPath            string
		ExpectedStatusCode int
	}{
		{"with valid SKU", true, false, products.Data[0].SKU, http.StatusOK},
		{"with missing sku in url", true, false, "", http.StatusBadRequest},
		{"with missing items", false, false, products.Data[0].SKU, http.StatusNotFound},
		{"with invalid inventory json", true, true, products.Data[0].SKU, http.StatusInternalServerError},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			err := c.DeleteInventory()
			require.NoError(t, err)

			if currentTest.WriteInventory {
				if currentTest.BadInventory {
					err := ioutil.WriteFile(c.inventoryFileName, []byte("invalid json test"), 0644)
					require.NoError(t, err)
				} else {
					err := c.WriteInventory()
					require.NoError(t, err)
				}
				defer func() {
					_ = os.Remove(c.inventoryFileName)
				}()
			}

			req := httptest.NewRequest("GET", "http://localhost:48096/inventory/"+test.URLPath, nil)
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"sku": currentTest.URLPath})
			c.InventoryItemGet(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")
		})
	}
}

// TestAuditLogGetAll tests the ability to get all audit logs
// related functions
func TestAuditLogGetAll(t *testing.T) {
	// Audit slice

	audits := getDefaultAuditsList()
	c := Controller{
		lc:               logger.NewMockClient(),
		service:          nil,
		auditLog:         audits,
		auditLogFileName: AuditLogFileName,
	}
	tests := []struct {
		Name               string
		BadAuditLog        bool
		ExpectedStatusCode int
	}{
		{"AuditLogGetAll", false, http.StatusOK},
		{"with invalid audit log json", true, http.StatusInternalServerError},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			err := c.DeleteAuditLog()
			require.NoError(t, err)

			if currentTest.BadAuditLog {
				err := ioutil.WriteFile(c.auditLogFileName, []byte("invalid json test"), 0644)
				require.NoError(t, err)
			} else {
				err := c.WriteAuditLog()
				require.NoError(t, err)
			}
			defer func() {
				_ = os.Remove(c.auditLogFileName)
			}()

			req := httptest.NewRequest("GET", "http://localhost:48096/auditlog", nil)
			w := httptest.NewRecorder()
			c.AuditLogGetAll(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")
		})
	}
}

// TestAuditLogGetEntry tests the ability to get all audit logs
// related functions
func TestAuditLogGetEntry(t *testing.T) {
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

	tests := []struct {
		Name               string
		WriteAuditLog      bool
		BadAuditLog        bool
		URLPath            string
		ExpectedStatusCode int
	}{
		{"with valid Audit ID", true, false, audits.Data[0].AuditEntryID, http.StatusOK},
		{"with missing entry ID in url", true, false, "", http.StatusBadRequest},
		{"with missing items", false, false, audits.Data[0].AuditEntryID, http.StatusNotFound},
		{"with invalid audit log json", true, true, audits.Data[0].AuditEntryID, http.StatusInternalServerError},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			err := c.DeleteAuditLog()
			require.NoError(t, err)

			if currentTest.WriteAuditLog {
				if currentTest.BadAuditLog {
					err := ioutil.WriteFile(c.auditLogFileName, []byte("invalid json test"), 0644)
					require.NoError(t, err)
				} else {
					err := c.WriteAuditLog()
					require.NoError(t, err)
				}
				defer func() {
					_ = os.Remove(c.auditLogFileName)
				}()
			}

			req := httptest.NewRequest("GET", "http://localhost:48096/auditlog/"+test.URLPath, nil)
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"entry": currentTest.URLPath})
			c.AuditLogGetEntry(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")
		})
	}
}
