// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// LedgerAccountGet will get the transaction ledger for a specific account
func (c *Controller) LedgerAccountGet(writer http.ResponseWriter, req *http.Request) {
	//Get all ledgers for all accounts
	accountLedgers, err := c.GetAllLedgers()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve all ledgers for accounts %v", err.Error())
		c.lc.Error(errMsg)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(errMsg))
		return
	}

	// Get the current accountID from the request
	vars := mux.Vars(req)
	accountIDstr := vars["accountid"]
	accountID, err := strconv.Atoi(accountIDstr)
	if err != nil {
		errMsg := fmt.Sprintf("AccountID is invalid %v", err.Error())
		c.lc.Error(errMsg)
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(errMsg))
		return
	}

	if accountID >= 0 {
		for _, account := range accountLedgers.Data {
			if accountID == account.AccountID {
				accountLedger, err := json.Marshal(account)
				if err != nil {
					errMsg := fmt.Sprintf("Failed to retrieve account ledger %v", err.Error())
					c.lc.Error(errMsg)
					writer.WriteHeader(http.StatusInternalServerError)
					writer.Write([]byte(errMsg))
					return
				}
				c.lc.Info("GET ledger account successfully")
				writer.Write(accountLedger)
				return
			}
		}
		errMsg := fmt.Sprintf("AccountID %v not found in ledger", accountID)
		c.lc.Error(errMsg)
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(errMsg))
		return
	}
}

// AllAccountsGet will get the entire ledger with transactions for all accounts
func (c *Controller) AllAccountsGet(writer http.ResponseWriter, req *http.Request) {
	// Get the list of accounts with all ledgers
	accountLedgers, err := c.GetAllLedgers()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve all ledgers for accounts %v", err.Error())
		c.lc.Error(errMsg)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(errMsg))
		return
	}

	// No logic needs to be done here, since we are just reading the file
	// and writing it back out. Simply marshaling it will validate its structure
	accountLedgersJSON, err := json.Marshal(accountLedgers)
	if err != nil {
		errMsg := "Failed to unmarshal accountLedgers"
		c.lc.Errorf("%s: %s", errMsg, err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(errMsg))
		return
	}
	c.lc.Info("GET ALL ledger accounts successfully")
	writer.Write(accountLedgersJSON)
}
