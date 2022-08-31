// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

type Accounts struct {
	Data []Account `json:"data"`
}

type Ledger struct {
	TransactionID int64      `json:"transactionID,string"`
	TxTimeStamp   int64      `json:"txTimeStamp,string"`
	LineTotal     float64    `json:"lineTotal"`
	CreatedAt     int64      `json:"createdAt,string"`
	UpdatedAt     int64      `json:"updatedAt,string"`
	IsPaid        bool       `json:"isPaid"`
	LineItems     []LineItem `json:"lineItems"`
}

type LineItem struct {
	SKU         string  `json:"sku"`
	ProductName string  `json:"productName"`
	ItemPrice   float64 `json:"itemPrice"`
	ItemCount   int     `json:"itemCount"`
}

type Account struct {
	AccountID int      `json:"accountID"`
	Ledgers   []Ledger `json:"ledgers"`
}

type Product struct {
	SKU                string  `json:"sku"`
	ItemPrice          float64 `json:"itemPrice"`
	ProductName        string  `json:"productName"`
	UnitsOnHand        int     `json:"unitsOnHand"`
	MaxRestockingLevel int     `json:"maxRestockingLevel"`
	MinRestockingLevel int     `json:"minRestockingLevel"`
	CreatedAt          int64   `json:"createdAt,string"`
	UpdatedAt          int64   `json:"updatedAt,string"`
	IsActive           bool    `json:"isActive"`
}

type paymentInfo struct {
	AccountID     int   `json:"accountID"`
	TransactionID int64 `json:"transactionID,string"`
	IsPaid        bool  `json:"isPaid"`
}

type deltaLedger struct {
	AccountID int        `json:"accountId"`
	DeltaSKUs []deltaSKU `json:"deltaSKUs"`
}

type deltaSKU struct {
	SKU   string `json:"sku"`
	Delta int    `json:"delta"`
}
