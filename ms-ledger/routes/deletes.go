// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

// LedgerDelete will delete a specific ledger for an account
func (c *Controller) LedgerDelete(writer http.ResponseWriter, req *http.Request) {

	//Get all ledgers for all accounts
	accountLedgers, err := c.GetAllLedgers()
	if err != nil {
		errMsg := fmt.Sprintf("failed to retrieve all ledgers for accounts: %v", err.Error())
		c.lc.Error(errMsg)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(errMsg))
		return
	}

	// Get variables from HTTP request
	vars := mux.Vars(req)
	tidstr := vars["tid"]
	tid, tiderr := strconv.ParseInt(tidstr, 10, 64)
	if tiderr != nil {
		errMsg := "transactionID contains bad data"
		c.lc.Error("%s: %s", errMsg, tiderr.Error())
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(errMsg))
		return
	}

	accountIDstr := vars["accountid"]
	accountID, accountIDerr := strconv.Atoi(accountIDstr)
	if accountIDerr != nil {
		errMsg := "accountID contains bad data"
		c.lc.Error("%s: %s", errMsg, accountIDerr.Error())
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(errMsg))
		return
	}

	//Iterate through accounts
	if tid > 0 && accountID >= 0 {
		for accountIndex, account := range accountLedgers.Data {
			if accountID == account.AccountID {
				for ledgerIndex, ledger := range account.Ledgers {
					if tid == ledger.TransactionID {
						accountLedgers.Data[accountIndex].Ledgers = append(account.Ledgers[:ledgerIndex], account.Ledgers[ledgerIndex+1:]...)

						data, err := json.Marshal(accountLedgers)
						if err != nil {
							errMsg := "marshal failed for update ledger with deleted transaction"
							c.lc.Errorf("%s: %s", errMsg, err.Error())
							writer.WriteHeader(http.StatusInternalServerError)
							writer.Write([]byte(errMsg))
							return
						}

						if err = os.WriteFile(c.ledgerFileName, data, 0644); err != nil {
							errMsg := "write failed for update ledger with deleted transaction"
							c.lc.Errorf("%s: %s", errMsg, err.Error())
							writer.WriteHeader(http.StatusInternalServerError)
							writer.Write([]byte(errMsg))
							return
						}
						c.lc.Info("Deleted ledger successfully")
						writer.Write([]byte("Deleted ledger " + tidstr))
						return
					}
				}
				errMsg := fmt.Sprintf("Could not find Transaction %v", strconv.FormatInt(tid, 10))
				c.lc.Errorf(errMsg)
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(errMsg))
				return
			}
		}

		errMsg := fmt.Sprintf("Could not find account %v", strconv.Itoa(accountID))
		c.lc.Errorf(errMsg)
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(errMsg))
		return
	}
}
