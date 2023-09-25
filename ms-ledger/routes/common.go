// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

const (
	connectionTimeout = 15
)

// GetAllLedgers is a common function to get all ledgers for all accounts
func (c *Controller) GetAllLedgers() (Accounts, error) {
	var accountLedgers Accounts

	err := utilities.LoadFromJSONFile(c.ledgerFileName, &accountLedgers)
	if err != nil {
		return Accounts{}, errors.New(
			"Failed to load ledger JSON file: " + err.Error(),
		)
	}
	return accountLedgers, nil
}

// DeleteAllLedgers will reset the content of the inventory JSON file
func (c *Controller) DeleteAllLedgers() error {
	return utilities.WriteToJSONFile(c.ledgerFileName, Accounts{Data: []Account{}}, 0644)
}

// TODO: refactor this into the utilities package
func (c *Controller) sendCommand(method string, commandURL string, inputBytes []byte) (*http.Response, error) {
	// Create the http request based on the parameters
	request, _ := http.NewRequest(method, commandURL, bytes.NewBuffer(inputBytes))
	timeout := time.Duration(connectionTimeout) * time.Second
	client := &http.Client{
		Timeout: timeout,
	}

	// Execute the http request
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error sending data: %v", err.Error())
	}

	// Check the status code and return any errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error sending request: Received status code %v", resp.Status)
	}

	return resp, nil
}
