// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

// LedgerAccountGet will get the transaction ledger for a specific account
func (c *Controller) LedgerAccountGet(writer http.ResponseWriter, req *http.Request) {
	//Get all ledgers for all accounts
	accountLedgers, err := c.GetAllLedgers()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve all ledgers for accounts %v", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, errMsg, true)
		c.lc.Error(errMsg)
		return
	}

	// Get the current accountID from the request
	vars := mux.Vars(req)
	accountIDstr := vars["accountid"]
	accountID, err := strconv.Atoi(accountIDstr)
	if err != nil {
		errMsg := fmt.Sprintf("AccountID is invalid %v", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, true)
		c.lc.Error(errMsg)
		return
	}

	if accountID >= 0 {
		for _, account := range accountLedgers.Data {
			if accountID == account.AccountID {
				accountLedger, err := utilities.GetAsJSON(account)
				if err != nil {
					errMsg := fmt.Sprintf("Failed to retrieve account ledger %v", err.Error())
					utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, errMsg, true)
					c.lc.Error(errMsg)
					return
				}
				utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, accountLedger, false)
				c.lc.Info("GET ledger account successfully")
				return
			}
		}
		errMsg := fmt.Sprintf("AccountID %v not found in ledger", accountID)
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, false)
		c.lc.Error(errMsg)
		return
	}
}

// AllAccountsGet will get the entire ledger with transactions for all accounts
func (c *Controller) AllAccountsGet(writer http.ResponseWriter, req *http.Request) {
	// Get the list of accounts with all ledgers
	accountLedgers, err := c.GetAllLedgers()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve all ledgers for accounts %v", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, errMsg, true)
		c.lc.Error(errMsg)
		return
	}

	// No logic needs to be done here, since we are just reading the file
	// and writing it back out. Simply marshaling it will validate its structure
	accountLedgersJSON, err := utilities.GetAsJSON(accountLedgers)
	if err != nil {
		errMsg := "Failed to unmarshal accountLedgers"
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, errMsg, true)
		c.lc.Errorf("%s: %s", errMsg, err.Error())
		return
	}
	utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, accountLedgersJSON, false)
	c.lc.Info("GET ALL ledger accounts successfully")
}
