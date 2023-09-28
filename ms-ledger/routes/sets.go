// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

// SetPaymentStatus sets the `isPaid` field for a transaction to true/false
func (c *Controller) SetPaymentStatus(writer http.ResponseWriter, req *http.Request) {
	// Read request body
	body := make([]byte, req.ContentLength)
	_, err := io.ReadFull(req.Body, body)
	if err != nil {
		errMsg := "Failed to parse request body"
		c.lc.Errorf("%s: %s", errMsg, err.Error())
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(errMsg))
		return
	}

	// Unmarshal the string contents of request into a proper structure
	var paymentStatus paymentInfo
	if err := json.Unmarshal(body, &paymentStatus); err != nil {
		errMsg := "Failed to unmarshal body"
		c.lc.Errorf("%s: %s", errMsg, err.Error())
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(errMsg))
		return
	}

	//Get all ledgers for all accounts
	accountLedgers, err := c.GetAllLedgers()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve all ledgers for accounts: %v", err.Error())
		c.lc.Error(errMsg)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(errMsg))
		return
	}

	for accountIndex, account := range accountLedgers.Data {
		if paymentStatus.AccountID == account.AccountID {
			for transactionIndex, transaction := range account.Ledgers {
				if paymentStatus.TransactionID == transaction.TransactionID {
					accountLedgers.Data[accountIndex].Ledgers[transactionIndex].IsPaid = paymentStatus.IsPaid

					data, err := json.Marshal(accountLedgers)
					if err != nil {
						errMsg := fmt.Sprintf("failed to marshal ledger JSON file for set: " + err.Error())
						c.lc.Errorf("%s: %s", errMsg, err.Error())
						writer.WriteHeader(http.StatusInternalServerError)
						writer.Write([]byte(errMsg))
						return
					}
					if err = os.WriteFile(c.ledgerFileName, data, 0644); err != nil {
						errMsg := fmt.Sprintf("failed to write ledger JSON file for set: " + err.Error())
						c.lc.Errorf("%s: %s", errMsg, err.Error())
						writer.WriteHeader(http.StatusInternalServerError)
						writer.Write([]byte(errMsg))
						return
					}

					infoMsg := fmt.Sprintf("Updated Payment Status for transaction %v", strconv.FormatInt(paymentStatus.TransactionID, 10))
					c.lc.Info(infoMsg)
					writer.WriteHeader(http.StatusOK)
					writer.Write([]byte(infoMsg))
					return
				}
			}
			errMsg := fmt.Sprintf("Could not find Transaction %v", strconv.FormatInt(paymentStatus.TransactionID, 10))
			c.lc.Error(errMsg)
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(errMsg))
			return
		}
	}
	errMsg := fmt.Sprintf("Could not find account %v", strconv.Itoa(paymentStatus.AccountID))
	c.lc.Error(errMsg)
	writer.WriteHeader(http.StatusBadRequest)
	writer.Write([]byte(errMsg))
}

// LedgerAddTransaction adds a new transaction to the Account Ledger
func (c *Controller) LedgerAddTransaction(writer http.ResponseWriter, req *http.Request) {

	// Read request body (this is the inference data)
	body := make([]byte, req.ContentLength)
	_, err := io.ReadFull(req.Body, body)
	if err != nil {
		errMsg := "Failed to parse request body"
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(errMsg))
		c.lc.Errorf("%s: %s", errMsg, err.Error())
		return
	}

	// Unmarshal the string contents of request for inference data into a proper structure
	// deltaLedger is accountID and list of Sku:delta
	var updateLedger deltaLedger
	if err := json.Unmarshal(body, &updateLedger); err != nil {
		errMsg := "Failed to unmarshal request body"
		c.lc.Errorf("%s: %s", errMsg, err.Error())
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(errMsg))
		return
	}

	//Get all ledgers for all accounts
	accountLedgers, err := c.GetAllLedgers()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve all ledgers for accounts %v", err.Error())
		c.lc.Error(errMsg)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(errMsg))
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
					c.lc.Error(errMsg)
					writer.WriteHeader(http.StatusBadRequest)
					writer.Write([]byte(errMsg))
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
		c.lc.Error("No ledger change in any account")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Account not found"))
		return
	}

	data, err := json.Marshal(accountLedgers)
	if err != nil {
		errMsg := fmt.Sprintf("failed to marshal ledger JSON file for update: " + err.Error())
		c.lc.Errorf("%s: %s", errMsg, err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(errMsg))
		return
	}
	if err = os.WriteFile(c.ledgerFileName, data, 0644); err != nil {
		errMsg := fmt.Sprintf("failed to write ledger JSON file for update: " + err.Error())
		c.lc.Errorf("%s: %s", errMsg, err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(errMsg))
		return
	}

	// return the new ledger as JSON, or if for some reason it cannot be processed back into
	// JSON for returning to the user, fallback to a simple string
	newLedgerJSON, err := json.Marshal(newLedger)
	if err != nil {
		c.lc.Warnf("Updated ledger successfully with error %s", err.Error())
		writer.Write([]byte("Updated ledger successfully, but could not marshal to json"))
	} else {
		c.lc.Infof("Updated ledger %s successfully", newLedgerJSON)
		writer.Write(newLedgerJSON)
	}
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Product{}, fmt.Errorf("Could not read response body from InventoryEndpoint")
	}

	// Prepare to unmarshal the desired inventory item from the HTTP response's body (json)
	var inventoryItem Product
	err = json.Unmarshal(body, &inventoryItem)
	if err != nil {
		return Product{}, fmt.Errorf("Received an invalid data structure from InventoryEndpoint")
	}

	return inventoryItem, nil
}
