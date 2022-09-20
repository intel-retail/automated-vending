// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

// LedgerAccountGet will get the transaction ledger for a specific account
func (c *Controller) LedgerAccountGet(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {
		//Get all ledgers for all accounts
		accountLedgers, err := c.GetAllLedgers()
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to retrieve all ledgers for accounts "+err.Error(), true)
			c.lc.Errorf("Failed to retrieve all ledgers for accounts %s", err.Error())
			return
		}

		// Get the current accountID from the request
		vars := mux.Vars(req)
		accountIDstr := vars["accountid"]
		accountID, err := strconv.Atoi(accountIDstr)
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "AccountID is invalid "+err.Error(), true)
			c.lc.Errorf("AccountID is invalid %s", err.Error())
			return
		}

		if accountID >= 0 {
			for _, account := range accountLedgers.Data {
				if accountID == account.AccountID {
					accountLedger, err := utilities.GetAsJSON(account)
					if err != nil {
						utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to retrieve account ledger "+err.Error(), true)
						c.lc.Errorf("Failed to retrieve account ledger %s", err.Error())
						return
					}
					utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, accountLedger, false)
					c.lc.Info("GET ledger account successfully")
					return
				}
			}
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "AccountID not found in ledger", false)
			c.lc.Errorf("AccountID %s not found in ledger", accountID)
			return
		}
	})
}

// AllAccountsGet will get the entire ledger with transactions for all accounts
func (c *Controller) AllAccountsGet(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {

		// Get the list of accounts with all ledgers
		accountLedgers, err := c.GetAllLedgers()
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to retrieve all ledgers for accounts "+err.Error(), true)
			c.lc.Errorf("Failed to retrieve all ledgers for accounts %s", err.Error())
			return
		}

		// No logic needs to be done here, since we are just reading the file
		// and writing it back out. Simply marshaling it will validate its structure
		accountLedgersJSON, err := utilities.GetAsJSON(accountLedgers)
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to unmarshal accountLedgers", true)
			c.lc.Errorf("Failed to unmarshal accountLedgers %s", err.Error())
			return
		}
		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, accountLedgersJSON, false)
		c.lc.Info("GET ALL ledger accounts successfully")
	})
}
