// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"io/ioutil"
	"strconv"
	"testing"

	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWritePeople tests the ability to write people
func TestWritePeople(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	require := require.New(t)

	expectedPeople := setupPeople()
	writePeople := expectedPeople.WritePeople()
	require.NoError(writePeople, "Failed WritePeople() function")

	actualPeople := People{}
	loadJSONFileErr := utilities.LoadFromJSONFile(PeopleFileName, &actualPeople)
	require.NoError(loadJSONFileErr, "Failed to load people from file")

	assert.Equal(t, expectedPeople, actualPeople, "Output JSON content mismatch in "+PeopleFileName)
}

// TestGetPeople tests the ability to get people
func TestGetPeople(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedPeople := setupPeople()
	writePeople := expectedPeople.WritePeople()
	require.NoError(t, writePeople, "Failed WritePeople() function")

	actualPeople, err := GetPeopleData()
	assert.NoError(err, "Error getting actualPeople")

	assert.Equal(expectedPeople, actualPeople, "Output JSON content mismatch in "+PeopleFileName)
}

// TestGetPeopleByPersonIDQueryFunctions tests the get people by person ID function
func TestGetPeopleByPersonIDQueryFunctions(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedPeople := setupPeople()
	writePeople := expectedPeople.WritePeople()
	require.NoError(t, writePeople, "Failed WritePeople() function")

	tests := []struct {
		Name           string
		PersonID       int
		ExpectedPerson Person
		ShouldMatch    bool
	}{
		{"Test ByPersonID", 1, expectedPeople.People[0], true},
		{"Test ByPersonID Non Existent Person", -1, expectedPeople.People[0], false},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			returnedPerson := expectedPeople.GetPersonByPersonID(currentTest.PersonID)
			if currentTest.ShouldMatch {
				assert.Equal(currentTest.ExpectedPerson, returnedPerson, "Person should match")
			} else {
				assert.NotEqual(currentTest.ExpectedPerson, returnedPerson, "Person should Not match")
			}
		})
	}
}

// TestGetPeopleByAccountIDQueryFunctions tests the get people by account ID function
func TestGetPeopleByAccountIDQueryFunctions(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedPeople := setupPeople()
	writePeople := expectedPeople.WritePeople()
	require.NoError(t, writePeople, "Failed WritePeople() function")

	tests := []struct {
		Name           string
		AccountID      int
		ExpectedPerson Person
		ShouldMatch    bool
	}{
		{"Test ByAccountID", 2, expectedPeople.People[1], true},
		{"Test ByAccountID Non Existent Account", -2, expectedPeople.People[1], false},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			returnedPerson := expectedPeople.GetPersonByAccountID(currentTest.AccountID)
			if currentTest.ShouldMatch {
				assert.Equal(currentTest.ExpectedPerson, returnedPerson, "Person should match")
			} else {
				assert.NotEqual(currentTest.ExpectedPerson, returnedPerson, "Person should Not match")
			}
		})
	}
}

// TestGetPeopleByFullNameQueryFunctions tests the get people by full name function
func TestGetPeopleByFullNameQueryFunctions(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedPeople := setupPeople()
	writePeople := expectedPeople.WritePeople()
	require.NoError(t, writePeople, "Failed WritePeople() function")

	tests := []struct {
		Name           string
		FullName       string
		ExpectedPerson Person
		ShouldMatch    bool
	}{
		{"Test ByFullName", "Test Person 3", expectedPeople.People[2], true},
		{"Test ByFullName Non Existent Person", "Non-Existent Test Person 3", expectedPeople.People[2], false},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			returnedPerson := expectedPeople.GetPersonByFullName(currentTest.FullName)
			if currentTest.ShouldMatch {
				assert.Equal(currentTest.ExpectedPerson, returnedPerson, "Person should match")
			} else {
				assert.NotEqual(currentTest.ExpectedPerson, returnedPerson, "Person should Not match")
			}
		})
	}
}

// TestDeletePerson tests the ability to delete a person
func TestDeletePerson(t *testing.T) {
	expectedPeople := setupPeople()
	writePeople := expectedPeople.WritePeople()
	require.NoError(t, writePeople, "Failed WritePeople() function")

	deletedPerson := expectedPeople.People[0]
	expectedPeople.DeletePerson(expectedPeople.People[0])

	assert.NotEqual(t, expectedPeople.People[0], deletedPerson, "Deleted person with ID "+strconv.Itoa(expectedPeople.People[0].PersonID)+" but it still exists in the test list")
}

// TestDeletePeople tests the ability to delete people
func TestDeletePeople(t *testing.T) {
	expectedPeople := setupPeople()
	writePeople := expectedPeople.WritePeople()
	require.NoError(t, writePeople, "Failed WritePeople() function")

	deleteErr := DeletePeople()
	assert.NoError(t, deleteErr, "Failed DeletePeople() function")

	actualPeople := People{People: []Person{}}
	loadJSONFileErr := utilities.LoadFromJSONFile(PeopleFileName, &actualPeople)
	require.NoError(t, loadJSONFileErr, "Failed to load people from file")

	assert.Equal(t, actualPeople, People{People: []Person{}}, "Output JSON content mismatch in "+PeopleFileName)
}

// TestGetPeopleDataError tests error case for get people data
func TestGetPeopleDataError(t *testing.T) {
	expectedPeople := setupPeople()
	writePeople := expectedPeople.WritePeople()
	require.NoError(t, writePeople, "Failed WritePeople() function")

	writeErr := ioutil.WriteFile(PeopleFileName, []byte("invalid json test"), 0644)
	require.NoError(t, writeErr, "Failed to write to test file")
	_, err := GetPeopleData()
	assert.Error(t, err, "Expected failure calling GetPeopleData() for invalid JSON contents but did not get one")
}

// TestWriteAccounts tests the ability to write accounts
func TestWriteAccounts(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	require := require.New(t)

	expectedAccounts := setupAccounts()
	writeAccounts := expectedAccounts.WriteAccounts()
	require.NoError(writeAccounts, "Failed WriteAccounts() function")

	// load accounts from file to validate
	accountsFromFile := Accounts{}
	loadJSONFileErr := utilities.LoadFromJSONFile(AccountsFileName, &accountsFromFile)
	require.NoError(loadJSONFileErr, "Failed to load accounts from file")

	assert.Equal(t, expectedAccounts, accountsFromFile, "Output JSON content mismatch in "+AccountsFileName)
}

// TestGetAccounts tests the ability to write accounts
func TestGetAccounts(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedAccounts := setupAccounts()
	writeAccounts := expectedAccounts.WriteAccounts()
	require.NoError(t, writeAccounts, "Failed WriteAccounts() function")

	// run GetAccountsData and get the result as JSON
	accountsFromFile, err := GetAccountsData()
	assert.NoError(err, "Error getting accountsFromFile")

	assert.Equal(expectedAccounts, accountsFromFile, "Output JSON content mismatch in "+AccountsFileName)
}

// TestGetAccountByAccountIDQueryFunctions tests the get account by account ID functions
func TestGetAccountByAccountIDQueryFunctions(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedAccounts := setupAccounts()
	writeAccounts := expectedAccounts.WriteAccounts()
	require.NoError(t, writeAccounts, "Failed WriteAccounts() function")

	tests := []struct {
		Name           string
		AccountID      int
		ExpectedOutput Account
		ShouldMatch    bool
	}{
		{"Test ByAccountID", 1, expectedAccounts.Accounts[0], true},
		{"Test ByAccountID Non Existent Account", -1, expectedAccounts.Accounts[0], false},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			returnedAccount := expectedAccounts.GetAccountByAccountID(currentTest.AccountID)
			if currentTest.ShouldMatch {
				assert.Equal(currentTest.ExpectedOutput, returnedAccount, "Accounts should match")
			} else {
				assert.NotEqual(currentTest.ExpectedOutput, returnedAccount, "Accounts should Not match")
			}
		})
	}
}

// TestGetAccountByAddressQueryFunctions tests the get account by address functions
func TestGetAccountByAddressQueryFunctions(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedAccounts := setupAccounts()
	writeAccounts := expectedAccounts.WriteAccounts()
	require.NoError(t, writeAccounts, "Failed WriteAccounts() function")

	tests := []struct {
		Name           string
		Address        string
		ExpectedOutput Account
		ShouldMatch    bool
	}{
		{"Test ByAddress", "2 Test Lane", expectedAccounts.Accounts[1], true},
		{"Test ByAddress Non Existent Account", "-1 Test Lane", expectedAccounts.Accounts[1], false},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			returnedAccount := expectedAccounts.GetAccountByAddress(currentTest.Address)
			if currentTest.ShouldMatch {
				assert.Equal(currentTest.ExpectedOutput, returnedAccount, "Accounts should match")
			} else {
				assert.NotEqual(currentTest.ExpectedOutput, returnedAccount, "Accounts should Not match")
			}
		})
	}
}

// TestGetAccountByCreditCardNumberQueryFunctions tests the get account by credit card number functions
func TestGetAccountByCreditCardNumberQueryFunctions(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedAccounts := setupAccounts()
	writeAccounts := expectedAccounts.WriteAccounts()
	require.NoError(t, writeAccounts, "Failed WriteAccounts() function")

	tests := []struct {
		Name             string
		CreditCardNumber string
		ExpectedOutput   Account
		ShouldMatch      bool
	}{
		{"Test ByCreditCardNumber", "3234 4567 9012 3456", expectedAccounts.Accounts[2], true},
		{"Test ByCreditCardNumber Non Existent Account", "-3234 4567 9012 3456", expectedAccounts.Accounts[2], false},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			returnedAccount := expectedAccounts.GetAccountByCreditCardNumber(currentTest.CreditCardNumber)
			if currentTest.ShouldMatch {
				assert.Equal(currentTest.ExpectedOutput, returnedAccount, "Accounts should match")
			} else {
				assert.NotEqual(currentTest.ExpectedOutput, returnedAccount, "Accounts should Not match")
			}
		})
	}
}

// TestGetAccountByPhoneNumberQueryFunctions tests the get account by phone number functions
func TestGetAccountByPhoneNumberQueryFunctions(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedAccounts := setupAccounts()
	writeAccounts := expectedAccounts.WriteAccounts()
	require.NoError(t, writeAccounts, "Failed WriteAccounts() function")

	tests := []struct {
		Name           string
		PhoneNumber    string
		ExpectedOutput Account
		ShouldMatch    bool
	}{
		{"Test ByPhoneNumber", "4234567890", expectedAccounts.Accounts[3], true},
		{"Test ByPhoneNumber Non Existent Account", "-4234567890", expectedAccounts.Accounts[3], false},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			returnedAccount := expectedAccounts.GetAccountByPhoneNumber(currentTest.PhoneNumber)
			if currentTest.ShouldMatch {
				assert.Equal(currentTest.ExpectedOutput, returnedAccount, "Accounts should match")
			} else {
				assert.NotEqual(currentTest.ExpectedOutput, returnedAccount, "Accounts should Not match")
			}
		})
	}
}

// TestGetAccountByEmailAddressQueryFunctions tests the get account by Email address functions
func TestGetAccountByEmailAddressQueryFunctions(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedAccounts := setupAccounts()
	writeAccounts := expectedAccounts.WriteAccounts()
	require.NoError(t, writeAccounts, "Failed WriteAccounts() function")

	tests := []struct {
		Name           string
		EmailAddress   string
		ExpectedOutput Account
		ShouldMatch    bool
	}{
		{"Test ByEmailAddress", "test5@example.com", expectedAccounts.Accounts[4], true},
		{"Test ByEmailAddress Non Existent Account", "test-5@example.com", expectedAccounts.Accounts[4], false},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			returnedAccount := expectedAccounts.GetAccountByEmailAddress(currentTest.EmailAddress)
			if currentTest.ShouldMatch {
				assert.Equal(currentTest.ExpectedOutput, returnedAccount, "Accounts should match")
			} else {
				assert.NotEqual(currentTest.ExpectedOutput, returnedAccount, "Accounts should Not match")
			}
		})
	}
}

// TestDeleteAccount tests the ability to delete account
func TestDeleteAccount(t *testing.T) {
	expectedAccounts := setupAccounts()
	writeAccounts := expectedAccounts.WriteAccounts()
	require.NoError(t, writeAccounts, "Failed WriteAccounts() function")

	deletedAccount := expectedAccounts.Accounts[0]
	expectedAccounts.DeleteAccount(expectedAccounts.Accounts[0])

	assert.NotEqual(t, expectedAccounts.Accounts[0], deletedAccount, "Deleted account with ID "+strconv.Itoa(expectedAccounts.Accounts[0].AccountID)+" but it still exists in the test list")
}

// TestDeleteAccounts tests the ability to delete accounts
func TestDeleteAccounts(t *testing.T) {
	expectedAccounts := setupAccounts()
	writeAccounts := expectedAccounts.WriteAccounts()
	require.NoError(t, writeAccounts, "Failed WriteAccounts() function")

	deleteErr := DeleteAccounts()
	assert.NoError(t, deleteErr, "Failed DeleteAccounts() function")

	// load accounts from file to validate
	accountsFromFile := Accounts{Accounts: []Account{}}
	loadJSONFileErr := utilities.LoadFromJSONFile(AccountsFileName, &accountsFromFile)
	require.NoError(t, loadJSONFileErr, "Failed to load accounts from file")

	assert.Equal(t, accountsFromFile, Accounts{Accounts: []Account{}}, "Output JSON content mismatch in "+AccountsFileName)
}

// TestGetAccountsDataError tests error case for get accounts data
func TestGetAccountsDataError(t *testing.T) {
	expectedAccounts := setupAccounts()
	writeAccounts := expectedAccounts.WriteAccounts()
	require.NoError(t, writeAccounts, "Failed WriteAccounts() function")

	writeErr := ioutil.WriteFile(AccountsFileName, []byte("invalid json test"), 0644)
	require.NoError(t, writeErr, "Failed to write to test file")
	_, err := GetAccountsData()
	assert.Error(t, err, "Expected failure calling GetAccountsData() for invalid JSON contents but did not get one")
}

// TestWriteCards tests the ability to write cards
func TestWriteCards(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	require := require.New(t)

	expectedCards := setupCards()
	writeCards := expectedCards.WriteCards()
	require.NoError(writeCards, "Failed WriteCards() function")

	// load cards from file to validate
	cardsFromFile := Cards{}
	loadJSONFileErr := utilities.LoadFromJSONFile(CardsFileName, &cardsFromFile)
	require.NoError(loadJSONFileErr, "Failed to load cards from file")

	assert.Equal(t, expectedCards, cardsFromFile, "Output JSON content mismatch in "+CardsFileName)
}

// TestGetCards tests the ability to write cards
func TestGetCards(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedCards := setupCards()
	writeCards := expectedCards.WriteCards()
	require.NoError(t, writeCards, "Failed WriteCards() function")

	// load cards from file to validate
	cardsFromFile, err := GetCardsData()
	assert.NoError(err, "Error getting cardsFromFile")

	assert.Equal(expectedCards, cardsFromFile, "Output JSON content mismatch in "+CardsFileName)
}

// TestGetCardByCardIDQueryFunctions tests the get card by card ID functions
func TestGetCardByCardIDQueryFunctions(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedCards := setupCards()
	writeCards := expectedCards.WriteCards()
	require.NoError(t, writeCards, "Failed WriteCards() function")

	tests := []struct {
		Name           string
		CardID         string
		ExpectedOutput Card
		ShouldMatch    bool
	}{
		{"Test ByPersonID", "0001230001", expectedCards.Cards[0], true},
		{"Test ByPersonID Non Existent Account", "000123000-1", expectedCards.Cards[0], false},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			returnedCard := expectedCards.GetCardByCardID(currentTest.CardID)
			if currentTest.ShouldMatch {
				assert.Equal(currentTest.ExpectedOutput, returnedCard, "Accounts should match")
			} else {
				assert.NotEqual(currentTest.ExpectedOutput, returnedCard, "Accounts should Not match")
			}
		})
	}
}

// TestGetCardByRoleIDQueryFunctions tests the get card by role ID functions
func TestGetCardByRoleIDQueryFunctions(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedCards := setupCards()
	writeCards := expectedCards.WriteCards()
	require.NoError(t, writeCards, "Failed WriteCards() function")

	tests := []struct {
		Name           string
		RoleID         int
		ExpectedOutput Card
		ShouldMatch    bool
	}{
		{"Test ByAccountID", 2, expectedCards.Cards[1], true},
		{"Test ByAccountID Non Existent Account", -2, expectedCards.Cards[1], false},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			returnedCard := expectedCards.GetCardByRoleID(currentTest.RoleID)
			if currentTest.ShouldMatch {
				assert.Equal(currentTest.ExpectedOutput, returnedCard, "Accounts should match")
			} else {
				assert.NotEqual(currentTest.ExpectedOutput, returnedCard, "Accounts should Not match")
			}
		})
	}
}

// TestGetCardByPersonIDQueryFunctions tests the get card by person ID functions
func TestGetCardByPersonIDQueryFunctions(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	assert := assert.New(t)

	expectedCards := setupCards()
	writeCards := expectedCards.WriteCards()
	require.NoError(t, writeCards, "Failed WriteCards() function")

	tests := []struct {
		Name           string
		PersonID       int
		ExpectedOutput Card
		ShouldMatch    bool
	}{
		{"Test ByFullName", 3, expectedCards.Cards[2], true},
		{"Test ByFullName Non Existent Account", -3, expectedCards.Cards[2], false},
	}
	for _, test := range tests {
		currentTest := test
		t.Run(currentTest.Name, func(t *testing.T) {
			returnedCard := expectedCards.GetCardByPersonID(currentTest.PersonID)
			if currentTest.ShouldMatch {
				assert.Equal(currentTest.ExpectedOutput, returnedCard, "Accounts should match")
			} else {
				assert.NotEqual(currentTest.ExpectedOutput, returnedCard, "Accounts should Not match")
			}
		})
	}
}

// TestDeleteCard tests the ability to delete card
func TestDeleteCard(t *testing.T) {
	expectedCards := setupCards()
	writeCards := expectedCards.WriteCards()
	require.NoError(t, writeCards, "Failed WriteCards() function")

	deletedCard := expectedCards.Cards[0]
	expectedCards.DeleteCard(expectedCards.Cards[0])

	assert.NotEqual(t, expectedCards.Cards[0], deletedCard, "Deleted card with ID "+expectedCards.Cards[0].CardID+" but it still exists in the test list")
}

// TestDeleteCards tests the ability to delete cards
func TestDeleteCards(t *testing.T) {
	// use community-recommended shorthand (known name clash)
	require := require.New(t)

	expectedCards := setupCards()
	writeCards := expectedCards.WriteCards()
	require.NoError(writeCards, "Failed WriteCards() function")

	deleteErr := DeleteCards()
	require.NoError(deleteErr, "Failed DeleteCards() function")

	// load cards from file to validate
	cardsFromFile := Cards{Cards: []Card{}}
	loadJSONFileErr := utilities.LoadFromJSONFile(CardsFileName, &cardsFromFile)
	require.NoError(loadJSONFileErr, "Failed to load cards from file")

	assert.Equal(t, cardsFromFile, Cards{Cards: []Card{}}, "Output JSON content mismatch in "+CardsFileName)
}

// TestGetCardsDataError tests error case for get cards data
func TestGetCardsDataError(t *testing.T) {
	expectedCards := setupCards()
	writeCards := expectedCards.WriteCards()
	require.NoError(t, writeCards, "Failed WriteCards() function")

	writeErr := ioutil.WriteFile(CardsFileName, []byte("invalid json test"), 0644)
	require.NoError(t, writeErr, "Failed to write to test file")
	_, err := GetCardsData()
	assert.Error(t, err, "Expected failure calling GetCardsData() for invalid JSON contents but did not get one")
}
