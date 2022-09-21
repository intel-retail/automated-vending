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

// LedgerDelete will delete a specific ledger for an account
func (c *Controller) LedgerDelete(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {

		//Get all ledgers for all accounts
		accountLedgers, err := c.GetAllLedgers()
		if err != nil {
			errMsg := fmt.Sprintf("failed to retrieve all ledgers for accounts: %v", err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, errMsg, true)
			c.lc.Error(errMsg)
			return
		}

		// Get variables from HTTP request
		vars := mux.Vars(req)
		tidstr := vars["tid"]
		tid, tiderr := strconv.ParseInt(tidstr, 10, 64)
		if tiderr != nil {
			errMsg := "transactionID contains bad data"
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, true)
			c.lc.Error("%s: %s", errMsg, tiderr.Error())
			return
		}

		accountIDstr := vars["accountid"]
		accountID, accountIDerr := strconv.Atoi(accountIDstr)
		if accountIDerr != nil {
			errMsg := "accountID contains bad data"
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, true)
			c.lc.Error("%s: %s", errMsg, accountIDerr.Error())
			return
		}

		//Iterate through accounts
		if tid > 0 && accountID >= 0 {
			for accountIndex, account := range accountLedgers.Data {
				if accountID == account.AccountID {
					for ledgerIndex, ledger := range account.Ledgers {
						if tid == ledger.TransactionID {
							accountLedgers.Data[accountIndex].Ledgers = append(account.Ledgers[:ledgerIndex], account.Ledgers[ledgerIndex+1:]...)

							err := utilities.WriteToJSONFile(c.ledgerFileName, &accountLedgers, 0644)
							if err != nil {
								errMsg := "Failed to update ledger with deleted transaction"
								utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, errMsg, true)
								c.lc.Errorf("%s: %s", errMsg, err.Error())
								return
							}
							utilities.WriteStringHTTPResponse(writer, req, http.StatusOK, "Deleted ledger "+tidstr, false)
							c.lc.Info("Deleted ledger successfully")
							return
						}
					}
					errMsg := fmt.Sprintf("Could not find Transaction %v", strconv.FormatInt(tid, 10))
					utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, true)
					c.lc.Errorf(errMsg)
					return
				}
			}

			errMsg := fmt.Sprintf("Could not find account %v", strconv.Itoa(accountID))
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, true)
			c.lc.Errorf(errMsg)
			return
		}
	})
}
