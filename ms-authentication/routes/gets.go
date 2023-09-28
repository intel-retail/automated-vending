// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// AuthenticationGet accepts a 10-character URL parameter in the form:
// /authentication/0001230001
// It will look up the associated Person and Account for the given card and
// return an instance of AuthData
func (c *Controller) AuthenticationGet(writer http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	cardID := vars["cardid"]
	// check if the passed cardID is valid
	if cardID == "" || len(cardID) != 10 {
		c.lc.Infof("Please pass in a 10-character card ID as a URL parameter, like this: /authentication/0001230001")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Please pass in a 10-character card ID as a URL parameter, like this: /authentication/0001230001"))
		return
	}

	// load up all card data so we can find our card
	cards, err := GetCardsData()
	if err != nil {
		c.lc.Errorf("Failed to read authentication data: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("failed to read authentication data"))
		return
	}

	// check if the card's ID matches our given cardID
	card := cards.GetCardByCardID(cardID)
	if card.CardID != cardID {
		c.lc.Infof("Card ID: %s is not an authorized card", cardID)
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("Card ID is not an authorized card"))
		return
	}
	if !card.IsValid {
		c.lc.Infof("Card ID: %s is not an valid card", cardID)
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("Card ID is not a valid card"))
		return
	}

	// card is found, get the cardholder's AccountID, RoleID, and PersonID
	accounts, err := GetAccountsData()
	if err != nil {
		c.lc.Errorf("Failed to read accounts data: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("failed to read accounts data"))
		return
	}
	people, err := GetPeopleData()
	if err != nil {
		c.lc.Errorf("Failed to read people data: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("failed to read people data"))
		return
	}

	// begin to store the output AuthData
	authData := AuthData{CardID: cardID, RoleID: card.RoleID}

	// check if the associated person is valid
	person := people.GetPersonByPersonID(card.PersonID)
	if person.PersonID != card.PersonID {
		c.lc.Infof("Card ID is associated with an unknown person %s", person.PersonID)
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("Card ID is associated with an unknown person"))
		return
	}
	if !person.IsActive {
		c.lc.Infof("Card ID is associated with an inactive person %s", person.PersonID)
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("Card ID is associated with an inactive person"))
		return
	}

	// store the personID in the output AuthData
	authData.PersonID = person.PersonID

	// check if the associated account is valid
	account := accounts.GetAccountByAccountID(person.AccountID)
	if account.AccountID != person.AccountID {
		c.lc.Infof("Card ID is associated with an unknown account %s", person.AccountID)
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("Card ID is associated with an unknown account"))
		return
	}
	if !account.IsActive {
		c.lc.Infof("Card ID is associated with an inactive account %s", person.AccountID)
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("Card ID is associated with an inactive account"))
		return
	}

	// store the accountID in the output AuthData
	authData.AccountID = account.AccountID

	authDataJSON, err := json.Marshal(authData)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("failed to marshal authentication data"))
	}

	// Because of how type-safe Go is, it's actually impossible to
	// reach this error condition based on how this function is written
	// Generally json.Marshal can throw errors if you pass a chan
	// or something unmarshalable, but since authData is simply a struct
	// with only ints and strings, we can't actually _not_ marshal it ever
	// (I did some searching and that is my conclusion, I'm not stating this
	// as fact)

	c.lc.Infof("Successfully authenticated person and card")
	writer.Write(authDataJSON)
}
