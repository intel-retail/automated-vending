// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupPeople() People {
	return People{
		People: []Person{
			{
				PersonID:  1,
				AccountID: 1,
				FullName:  "Test Person 1",
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
				IsActive:  true,
			}, {
				PersonID:  2,
				AccountID: 2,
				FullName:  "Test Person 2",
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
				IsActive:  false,
			}, {
				PersonID:  3,
				AccountID: 3,
				FullName:  "Test Person 3",
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
				IsActive:  true,
			}, {
				PersonID:  4,
				AccountID: 4,
				FullName:  "Test Person 4",
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
				IsActive:  false,
			}, {
				PersonID:  5,
				AccountID: -1,
				FullName:  "Test Person 5",
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
				IsActive:  true,
			}, {
				PersonID:  6,
				AccountID: 6,
				FullName:  "Test Person 6",
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
				IsActive:  true,
			}, {
				PersonID:  7,
				AccountID: 7,
				FullName:  "Test Person 7",
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
				IsActive:  false,
			},
		},
	}
}

func setupAccounts() Accounts {
	return Accounts{
		Accounts: []Account{
			{
				AccountID:        1,
				Address:          "1 Test Lane",
				CreditCardNumber: "1234 4567 9012 3456",
				PhoneNumber:      "1234567890",
				EmailAddress:     "test1@example.com",
				CreatedAt:        1560815799,
				UpdatedAt:        1560815799,
				IsActive:         true,
			}, {
				AccountID:        2,
				Address:          "2 Test Lane",
				CreditCardNumber: "2234 4567 9012 3456",
				PhoneNumber:      "2234567890",
				EmailAddress:     "test2@example.com",
				CreatedAt:        1560815799,
				UpdatedAt:        1560815799,
				IsActive:         true,
			}, {
				AccountID:        3,
				Address:          "3 Test Lane",
				CreditCardNumber: "3234 4567 9012 3456",
				PhoneNumber:      "3234567890",
				EmailAddress:     "test3@example.com",
				CreatedAt:        1560815799,
				UpdatedAt:        1560815799,
				IsActive:         false,
			}, {
				AccountID:        4,
				Address:          "4 Test Lane",
				CreditCardNumber: "4234 4567 9012 3456",
				PhoneNumber:      "4234567890",
				EmailAddress:     "test4@example.com",
				CreatedAt:        1560815799,
				UpdatedAt:        1560815799,
				IsActive:         true,
			}, {
				AccountID:        5,
				Address:          "5 Test Lane",
				CreditCardNumber: "5234 4567 9012 3456",
				PhoneNumber:      "5234567890",
				EmailAddress:     "test5@example.com",
				CreatedAt:        1560815799,
				UpdatedAt:        1560815799,
				IsActive:         true,
			},
		},
	}
}

func setupCards() Cards {
	return Cards{
		Cards: []Card{
			{
				CardID:    "0001230001",
				RoleID:    1,
				IsValid:   true,
				PersonID:  1,
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
			}, {
				CardID:    "0001230002",
				RoleID:    2,
				IsValid:   true,
				PersonID:  2,
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
			}, {
				CardID:    "0001230003",
				RoleID:    3,
				IsValid:   true,
				PersonID:  3,
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
			}, {
				CardID:    "0001230004",
				RoleID:    1,
				IsValid:   false,
				PersonID:  4,
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
			}, {
				CardID:    "0001230005",
				RoleID:    1,
				IsValid:   true,
				PersonID:  -1,
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
			}, {
				CardID:    "0001230006",
				RoleID:    1,
				IsValid:   true,
				PersonID:  5,
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
			}, {
				CardID:    "TEST000000",
				RoleID:    1,
				IsValid:   true,
				PersonID:  1,
				CreatedAt: 1560815799,
				UpdatedAt: 1560815799,
			},
		},
	}
}

func writeJSONFiles(people People, accounts Accounts, cards Cards) error {
	err := people.WritePeople()
	if err != nil {
		return fmt.Errorf("Failed to write people to JSON file: %s", err.Error())
	}
	err = accounts.WriteAccounts()
	if err != nil {
		return fmt.Errorf("Failed to write accounts to JSON file: " + err.Error())
	}
	err = cards.WriteCards()
	if err != nil {
		return fmt.Errorf("Failed to write cards to JSON file: " + err.Error())
	}
	return nil
}

// TestAuthenticationGet tests the function AuthenticationGet, which
// is the primary endpoint of this application
func TestAuthenticationGet(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	require := require.New(t)
	assert := assert.New(t)

	people := setupPeople()
	accounts := setupAccounts()
	cards := setupCards()

	validAuthData := AuthData{
		AccountID: accounts.Accounts[0].AccountID,
		PersonID:  people.People[0].PersonID,
		RoleID:    cards.Cards[0].RoleID,
		CardID:    cards.Cards[0].CardID,
	}

	tests := []struct {
		Name             string
		WriteFiles       bool
		WriteInvalidFile string
		AuthData         AuthData
		StatusCode       int
		ExpectedError    string
	}{
		{"cardID length not eq 10", true, "", AuthData{CardID: "00"}, http.StatusBadRequest, "Please pass in a 10-character card ID as a URL parameter, like this: /authentication/0001230001"},
		{"Successful auth sequence", true, "", validAuthData, http.StatusOK, ""},
		{"Test inactive person", true, "", AuthData{CardID: cards.Cards[1].CardID}, http.StatusUnauthorized, "Card ID is associated with an inactive person"},
		{"Test inactive account", true, "", AuthData{CardID: cards.Cards[2].CardID}, http.StatusUnauthorized, "Card ID is associated with an inactive account"},
		{"Test inactive card", true, "", AuthData{CardID: cards.Cards[3].CardID}, http.StatusUnauthorized, "Card ID is not a valid card"},
		{"Test unknown card", true, "", AuthData{CardID: "ffffffffff"}, http.StatusUnauthorized, "Card ID is not an authorized card"},
		{"Test unknown person", true, "", AuthData{CardID: "0001230005"}, http.StatusUnauthorized, "Card ID is associated with an unknown person"},
		{"Test unknown account", true, "", AuthData{CardID: "0001230006"}, http.StatusUnauthorized, "Card ID is associated with an unknown account"},
		{"TestAuthenticationGet invalid cards JSON", true, CardsFileName, validAuthData, http.StatusInternalServerError, "failed to read authentication data"},
		{"TestAuthenticationGet invalid accounts JSON", true, AccountsFileName, validAuthData, http.StatusInternalServerError, "failed to read accounts data"},
		{"TestAuthenticationGet invalid people JSON", true, PeopleFileName, validAuthData, http.StatusInternalServerError, "failed to read people data"},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			mockAppService := &mocks.ApplicationService{}
			mockAppService.On("LoggingClient").Return(logger.NewMockClient())
			mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			c := NewController(mockAppService)

			if currentTest.WriteFiles {
				err := writeJSONFiles(people, accounts, cards)
				require.NoError(err, "Failed to write to test file")
			}

			if currentTest.WriteInvalidFile != "" {
				err := os.WriteFile(currentTest.WriteInvalidFile, []byte("invalid json test"), 0644)
				require.NoError(err, "Failed to write to test file")
			}

			req := httptest.NewRequest("GET", "/authentication/"+currentTest.AuthData.CardID, nil)
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"cardid": currentTest.AuthData.CardID})
			c.AuthenticationGet(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(err, "Failed to read response body")

			assert.Equal(resp.StatusCode, currentTest.StatusCode, "Expected status code to be OK: "+strconv.Itoa(resp.StatusCode))

			if resp.StatusCode == http.StatusOK {
				// Unmarshal the string contents of request into a proper structure
				responseAuthData := AuthData{}
				err := json.Unmarshal(body, &responseAuthData)
				assert.NoError(err, "Failed to unmarshal the authentication data")

				assert.Equal(responseAuthData, currentTest.AuthData)
				return
			}
			// check that the error message is as expected
			assert.Contains(string(body), currentTest.ExpectedError)
		})
	}
}
