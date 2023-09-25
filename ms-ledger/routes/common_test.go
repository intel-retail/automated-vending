// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"os"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
	"github.com/stretchr/testify/require"
)

const (
	LedgerFileName = "test-ledger.json"
)

func getDefaultAccountLedgers() Accounts {
	return Accounts{
		Data: []Account{{
			AccountID: 1,
			Ledgers: []Ledger{{
				TransactionID: 1579215712984890248,
				TxTimeStamp:   1579215712984890363,
				LineTotal:     1.99,
				CreatedAt:     1579215712984890443,
				UpdatedAt:     1579215712984890517,
				IsPaid:        false,
				LineItems: []LineItem{{
					SKU:         "1200050408",
					ProductName: "Mountain Dew - 16.9 oz",
					ItemPrice:   1.99,
					ItemCount:   1,
				}},
			}},
		}, {
			AccountID: 2,
			Ledgers: []Ledger{{
				TransactionID: 2579215712984890248,
				TxTimeStamp:   2579215712984890363,
				LineTotal:     2.99,
				CreatedAt:     2579215712984890443,
				UpdatedAt:     2579215712984890517,
				IsPaid:        false,
				LineItems: []LineItem{{
					SKU:         "2200050408",
					ProductName: "Mountain Blue - 16.9 oz",
					ItemPrice:   2.99,
					ItemCount:   1,
				}},
			}},
		}}}
}

func TestGetAllLedgers(t *testing.T) {
	// Use community-recommended shorthand (known name clash)
	c := Controller{
		lc:                logger.NewMockClient(),
		service:           nil,
		inventoryEndpoint: "test.com",
		ledgerFileName:    LedgerFileName,
	}
	require := require.New(t)
	// Accounts slice
	accountLedgers := getDefaultAccountLedgers()

	// Write the ledger
	err := utilities.WriteToJSONFile(c.ledgerFileName, &accountLedgers, 0644)
	require.NoError(err)
	defer func() {
		os.Remove(c.ledgerFileName)
	}()

	// run GetAllLedgers and get the result as JSON
	actualAccountLedgers, err := c.GetAllLedgers()
	require.NoError(err)

	// Check to make sure items match
	require.Equal(accountLedgers, actualAccountLedgers, "Ledgers should match")
}

func TestDeleteAllLedgers(t *testing.T) {
	// Use community-recommended shorthand (known name clash)

	mockAppService := &mocks.ApplicationService{}
	mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	c := Controller{
		lc:                logger.NewMockClient(),
		service:           mockAppService,
		inventoryEndpoint: "test.com",
		ledgerFileName:    LedgerFileName,
	}

	require := require.New(t)
	// Accounts slice
	accountLedgers := getDefaultAccountLedgers()

	expectedLedger := Accounts{Data: []Account{}}

	// Write the ledger
	err := utilities.WriteToJSONFile(c.ledgerFileName, &accountLedgers, 0644)
	require.NoError(err)
	defer func() {
		os.Remove(c.ledgerFileName)
	}()

	// Delete Ledger
	err = c.DeleteAllLedgers()
	require.NoError(err)
	updatedLedger, err := c.GetAllLedgers()
	require.NoError(err)

	// Check that deleted Ledger has no ledger data
	require.Equal(updatedLedger, expectedLedger, "Ledger should have no data")
}
