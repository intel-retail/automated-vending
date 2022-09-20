// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
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
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to retrieve all ledgers for accounts "+err.Error(), true)
			c.lc.Errorf("Failed to retrieve all ledgers for accounts %s", err.Error())
			return
		}

		// Get variables from HTTP request
		vars := mux.Vars(req)
		tidstr := vars["tid"]
		tid, tiderr := strconv.ParseInt(tidstr, 10, 64)
		if tiderr != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "transactionID contains bad data", true)
			c.lc.Errorf("transactionID contains bad data %s", tiderr.Error())
			return
		}

		accountIDstr := vars["accountid"]
		accountID, accountIDerr := strconv.Atoi(accountIDstr)
		if accountIDerr != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "accountID contains bad data", true)
			c.lc.Errorf("accountID contains bad data %s", accountIDerr.Error())
			return
		}

		//Iterate through accounts
		if tid > 0 && accountID >= 0 {
			for accountIndex, account := range accountLedgers.Data {
				if accountID == account.AccountID {
					for ledgerIndex, ledger := range account.Ledgers {
						if tid == ledger.TransactionID {
							accountLedgers.Data[accountIndex].Ledgers = append(account.Ledgers[:ledgerIndex], account.Ledgers[ledgerIndex+1:]...)

							err := utilities.WriteToJSONFile(LedgerFileName, &accountLedgers, 0644)
							if err != nil {
								utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to update ledger with deleted transaction", true)
								c.lc.Errorf("Failed to update ledger with deleted transaction %s", err.Error())
								return
							}
							utilities.WriteStringHTTPResponse(writer, req, http.StatusOK, "Deleted ledger "+tidstr, false)
							c.lc.Info("Deleted ledger successfully")
							return
						}
					}
					utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Could not find Transaction "+strconv.FormatInt(tid, 10), true)
					c.lc.Errorf("Could not find Transaction %s", strconv.FormatInt(tid, 10))
					return
				}
			}
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Could not find account "+strconv.Itoa(accountID), true)
			c.lc.Errorf("Could not find account %s", strconv.Itoa(accountID))
			return
		}
	})
}
