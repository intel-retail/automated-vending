// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
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
			err := DeleteAllLedgers()
			require.NoError(err)
			if currentTest.InvalidLedger {
				err = ioutil.WriteFile(LedgerFileName, []byte("invalid json test"), 0644)
			} else {
				err = utilities.WriteToJSONFile(LedgerFileName, &accountLedgers, 0644)
			}
			require.NoError(err)

			req := httptest.NewRequest("GET", "http://localhost:48093/ledger", nil)
			w := httptest.NewRecorder()
			AllAccountsGet(w, req)
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
			err := DeleteAllLedgers()
			require.NoError(err)
			if currentTest.InvalidLedger {
				err = ioutil.WriteFile(LedgerFileName, []byte("invalid json test"), 0644)
			} else {
				err = utilities.WriteToJSONFile(LedgerFileName, &accountLedgers, 0644)
			}
			require.NoError(err)

			req := httptest.NewRequest("GET", "http://localhost:48093/ledger/"+test.AccountID, nil)
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"accountid": currentTest.AccountID})
			LedgerAccountGet(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, currentTest.ExpectedStatusCode, resp.StatusCode, "invalid status code")
		})
	}
}
