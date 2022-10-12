// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
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
		utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Please pass in a 10-character card ID as a URL parameter, like this: /authentication/0001230001", false)
		return
	}

	// load up all card data so we can find our card
	cards, err := GetCardsData()
	if err != nil {
		c.lc.Errorf("Failed to read authentication data: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to read authentication data", true)
		return
	}

	// check if the card's ID matches our given cardID
	card := cards.GetCardByCardID(cardID)
	if card.CardID != cardID {
		c.lc.Infof("CardID: %s is not an authorized card", cardID)
		utilities.WriteStringHTTPResponse(writer, req, http.StatusUnauthorized, "Card ID is not an authorized card", false)
		return
	}
	if !card.IsValid {
		c.lc.Infof("CardID: %s is not an valid card", cardID)
		utilities.WriteStringHTTPResponse(writer, req, http.StatusUnauthorized, "Card ID is not a valid card", false)
		return
	}

	// card is found, get the cardholder's AccountID, RoleID, and PersonID
	accounts, err := GetAccountsData()
	if err != nil {
		c.lc.Errorf("Failed to read accounts data: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to read accounts data", true)
		return
	}
	people, err := GetPeopleData()
	if err != nil {
		c.lc.Errorf("Failed to read people data: %s", err.Error())
		utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to read people data", true)
		return
	}

	// begin to store the output AuthData
	authData := AuthData{CardID: cardID, RoleID: card.RoleID}

	// check if the associated person is valid
	person := people.GetPersonByPersonID(card.PersonID)
	if person.PersonID != card.PersonID {
		c.lc.Infof("CardID is associated with an unknown person %s", person.PersonID)
		utilities.WriteStringHTTPResponse(writer, req, http.StatusUnauthorized, "Card ID is associated with an unknown person", false)
		return
	}
	if !person.IsActive {
		c.lc.Infof("CardID is associated with an inactive person %s", person.PersonID)
		utilities.WriteStringHTTPResponse(writer, req, http.StatusUnauthorized, "Card ID is associated with an inactive person", false)
		return
	}

	// store the personID in the output AuthData
	authData.PersonID = person.PersonID

	// check if the associated account is valid
	account := accounts.GetAccountByAccountID(person.AccountID)
	if account.AccountID != person.AccountID {
		c.lc.Infof("CardID is associated with an unknown account %s", person.AccountID)
		utilities.WriteStringHTTPResponse(writer, req, http.StatusUnauthorized, "Card ID is associated with an unknown account", false)
		return
	}
	if !account.IsActive {
		c.lc.Infof("CardID is associated with an inactive account %s", person.AccountID)
		utilities.WriteStringHTTPResponse(writer, req, http.StatusUnauthorized, "Card ID is associated with an inactive account", false)
		return
	}

	// store the accountID in the output AuthData
	authData.AccountID = account.AccountID

	authDataJSON, _ := utilities.GetAsJSON(authData)

	// Because of how type-safe Go is, it's actually impossible to
	// reach this error condition based on how this function is written
	// Generally json.Marshal can throw errors if you pass a chan
	// or something unmarshalable, but since authData is simply a struct
	// with only ints and strings, we can't actually _not_ marshal it ever
	// (I did some searching and that is my conclusion, I'm not stating this
	// as fact)

	c.lc.Infof("Successfully authenticated person and card")
	utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, authDataJSON, false)
}
