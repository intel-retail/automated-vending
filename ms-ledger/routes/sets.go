// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"time"

	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

// SetPaymentStatus sets the `isPaid` field for a transaction to true/false
func (c *Controller) SetPaymentStatus(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {

		// Read request body
		body := make([]byte, req.ContentLength)
		_, err := io.ReadFull(req.Body, body)
		if err != nil {
			errMsg := "Failed to parse request body"
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, true)
			c.lc.Errorf("%s: %s", errMsg, err.Error())
			return
		}

		// Unmarshal the string contents of request into a proper structure
		var paymentStatus paymentInfo
		if err := json.Unmarshal(body, &paymentStatus); err != nil {
			errMsg := "Failed to unmarshal body"
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, true)
			c.lc.Errorf("%s: %s", errMsg, err.Error())
			return
		}

		//Get all ledgers for all accounts
		accountLedgers, err := c.GetAllLedgers()
		if err != nil {
			errMsg := fmt.Sprintf("Failed to retrieve all ledgers for accounts: %v", err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, errMsg, true)
			c.lc.Error(errMsg)
			return
		}

		for accountIndex, account := range accountLedgers.Data {
			if paymentStatus.AccountID == account.AccountID {
				for transactionIndex, transaction := range account.Ledgers {
					if paymentStatus.TransactionID == transaction.TransactionID {
						accountLedgers.Data[accountIndex].Ledgers[transactionIndex].IsPaid = paymentStatus.IsPaid

						err := utilities.WriteToJSONFile(LedgerFileName, &accountLedgers, 0644)
						if err != nil {
							errMsg := "Failed to update ledger"
							utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, errMsg, true)
							c.lc.Errorf("%s: %s", errMsg, err.Error())
							return
						}

						errMsg := fmt.Sprintf("Updated Payment Status for transaction %v", strconv.FormatInt(paymentStatus.TransactionID, 10))
						utilities.WriteStringHTTPResponse(writer, req, http.StatusOK, errMsg, false)
						c.lc.Info(errMsg)
						return
					}
				}
				errMsg := fmt.Sprintf("Could not find Transaction %v", strconv.FormatInt(paymentStatus.TransactionID, 10))
				utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, true)
				c.lc.Error(errMsg)
				return
			}
		}
		errMsg := fmt.Sprintf("Could not find account %v", strconv.Itoa(paymentStatus.AccountID))
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, true)
		c.lc.Error(errMsg)
	})
}

// LedgerAddTransaction adds a new transaction to the Account Ledger
func (c *Controller) LedgerAddTransaction(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {

		response := utilities.GetHTTPResponseTemplate()

		// Read request body (this is the inference data)
		body := make([]byte, req.ContentLength)
		_, err := io.ReadFull(req.Body, body)
		if err != nil {
			errMsg := "Failed to parse request body"
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, true)
			c.lc.Errorf("%s: %s", errMsg, err.Error())
			return
		}

		// Unmarshal the string contents of request for inference data into a proper structure
		// deltaLedger is accountID and list of Sku:delta
		var updateLedger deltaLedger
		if err := json.Unmarshal(body, &updateLedger); err != nil {
			errMsg := "Failed to unmarshal request body"
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, true)
			c.lc.Errorf("%s: %s", errMsg, err.Error())
			return
		}

		//Get all ledgers for all accounts
		accountLedgers, err := c.GetAllLedgers()
		if err != nil {
			errMsg := fmt.Sprintf("Failed to retrieve all ledgers for accounts %v", err.Error())
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, errMsg, true)
			c.lc.Error(errMsg)
			return
		}

		ledgerChanged := false
		var newLedger Ledger

		for accountIndex, account := range accountLedgers.Data {
			if updateLedger.AccountID == account.AccountID {
				newLedger = Ledger{
					TransactionID: time.Now().UnixNano(),
					TxTimeStamp:   time.Now().UnixNano(),
					LineTotal:     0,
					CreatedAt:     time.Now().UnixNano(),
					UpdatedAt:     time.Now().UnixNano(),
					IsPaid:        false,
					LineItems:     []LineItem{},
				}

				for _, deltaSKU := range updateLedger.DeltaSKUs {
					itemInfo, err := c.getInventoryItemInfo(c.inventoryEndpoint, deltaSKU.SKU)
					if err != nil {
						errMsg := fmt.Sprintf("Could not find product Info for %v errir: %v", deltaSKU.SKU, err.Error())
						utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, errMsg, true)
						c.lc.Error(errMsg)
						return
					}
					newLineItem := LineItem{
						SKU:         deltaSKU.SKU,
						ProductName: itemInfo.ProductName,
						ItemPrice:   itemInfo.ItemPrice,
						ItemCount:   int(math.Abs(float64(deltaSKU.Delta))),
					}
					newLedger.LineItems = append(newLedger.LineItems, newLineItem)
					newLedger.LineTotal = newLedger.LineTotal + (newLineItem.ItemPrice * float64(newLineItem.ItemCount))
				}

				// Add new Ledger to array of Ledgers for that account
				accountLedgers.Data[accountIndex].Ledgers = append(accountLedgers.Data[accountIndex].Ledgers, newLedger)
				ledgerChanged = true
			}
		}

		if !ledgerChanged {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Account not found", true)
			c.lc.Error("No ledger change in any account")
			return
		}

		err = utilities.WriteToJSONFile(LedgerFileName, &accountLedgers, 0644)
		if err != nil {
			errMsg := "Failed to update ledger"
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, errMsg, true)
			c.lc.Errorf("%s: %s", errMsg, err.Error())
			return
		}

		// return the new ledger as JSON, or if for some reason it cannot be processed back into
		// JSON for returning to the user, fallback to a simple string
		newLedgerJSON, err := utilities.GetAsJSON(newLedger)
		if err != nil {
			response.SetStringHTTPResponseFields(http.StatusOK, "Updated ledger successfully", false)
			c.lc.Warnf("Updated ledger successfully with error %s", err.Error())
		} else {
			response.SetJSONHTTPResponseFields(http.StatusOK, newLedgerJSON, false)
			c.lc.Infof("Updated ledger %s successfully", newLedgerJSON)
		}
		response.WriteHTTPResponse(writer, req)
	})
}

// getInventoryItemInfo is a helper function that will take the inference data (SKU)
// and return product details for a transaction to be recorded in the ledger
func (c *Controller) getInventoryItemInfo(inventoryEndpoint string, SKU string) (Product, error) {

	resp, err := c.sendCommand("GET", inventoryEndpoint+"/"+SKU, []byte(""))
	if err != nil {
		return Product{}, fmt.Errorf("Could not hit inventoryEndpoint, SKU may not exist")
	}

	defer resp.Body.Close()

	// Read the HTTP Response Body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Product{}, fmt.Errorf("Could not read response body from InventoryEndpoint")
	}

	// Prepare to store the http response in this variable
	var HTTPResponse utilities.HTTPResponse

	// Unmarshal the http response
	err = json.Unmarshal(body, &HTTPResponse)
	if err != nil {
		return Product{}, fmt.Errorf("Received an invalid data structure from InventoryEndpoint")
	}
	// Check the HTTP response error condition
	if HTTPResponse.Error {
		return Product{}, fmt.Errorf("Received an error response from the inventory service: " + HTTPResponse.Content.(string))
	}

	// Prepare to unmarshal the desired inventory item from the HTTP response's body (json)
	var inventoryItem Product
	err = json.Unmarshal([]byte(HTTPResponse.Content.(string)), &inventoryItem)
	if err != nil {
		return Product{}, fmt.Errorf("Received an invalid data structure from InventoryEndpoint")
	}

	return inventoryItem, nil
}
