// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/gorilla/mux"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllAccountsGet(t *testing.T) {
	// Use community-recommended shorthand (known name clash)
	require := require.New(t)

	// Accounts slice
	accountLedgers := getDefaultAccountLedgers()

	tests := []struct {
		Name               string
		InvalidLedger      bool
		ExpectedStatusCode int
	}{
		{"Valid Ledger", false, http.StatusOK},
		{"Invalid Ledger ", true, http.StatusInternalServerError},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			mockAppService := &mocks.ApplicationService{}
			mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil)
			c := Controller{
				lc:                logger.NewMockClient(),
				service:           mockAppService,
				inventoryEndpoint: "test.com",
				ledgerFileName:    LedgerFileName,
			}
			err := c.DeleteAllLedgers()
			require.NoError(err)
			if currentTest.InvalidLedger {
				err = ioutil.WriteFile(c.ledgerFileName, []byte("invalid json test"), 0644)
			} else {
				err = utilities.WriteToJSONFile(c.ledgerFileName, &accountLedgers, 0644)
			}
			require.NoError(err)
			defer func() {
				os.Remove(c.ledgerFileName)
			}()

			req := httptest.NewRequest("GET", "http://localhost:48093/ledger", nil)
			w := httptest.NewRecorder()
			c.AllAccountsGet(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")
		})
	}
}

func TestLedgerAccountGet(t *testing.T) {
	// Use community-recommended shorthand (known name clash)
	require := require.New(t)

	// Accounts slice
	accountLedgers := getDefaultAccountLedgers()
	invalidAccountID := "0"

	tests := []struct {
		Name               string
		InvalidLedger      bool
		AccountID          string
		ExpectedStatusCode int
	}{
		{"Valid Account ID", false, strconv.Itoa(accountLedgers.Data[0].AccountID), http.StatusOK},
		{"Bad data Account ID", false, "invalidAccountSyntax", http.StatusBadRequest},
		{"Nonexistent AccountID ", false, invalidAccountID, http.StatusBadRequest},
		{"Invalid Ledger", true, strconv.Itoa(accountLedgers.Data[0].AccountID), http.StatusInternalServerError},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			mockAppService := &mocks.ApplicationService{}
			mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil)
			c := Controller{
				lc:                logger.NewMockClient(),
				service:           mockAppService,
				inventoryEndpoint: "test.com",
				ledgerFileName:    LedgerFileName,
			}
			err := c.DeleteAllLedgers()
			require.NoError(err)
			if currentTest.InvalidLedger {
				err = ioutil.WriteFile(c.ledgerFileName, []byte("invalid json test"), 0644)
			} else {
				err = utilities.WriteToJSONFile(c.ledgerFileName, &accountLedgers, 0644)
			}
			require.NoError(err)
			defer func() {
				os.Remove(c.ledgerFileName)
			}()

			req := httptest.NewRequest("GET", "http://localhost:48093/ledger/"+test.AccountID, nil)
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"accountid": currentTest.AccountID})
			c.LedgerAccountGet(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")
		})
	}
}
