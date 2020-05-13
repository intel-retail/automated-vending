// Copyright © 2020 Intel Corporation. All rights reserved.
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
func AuthenticationGet(writer http.ResponseWriter, req *http.Request) {
	utilities.ProcessCORS(writer, req, func(writer http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		cardID := vars["cardid"]
		// check if the passed cardID is valid
		if cardID == "" || len(cardID) != 10 {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusBadRequest, "Please pass in a 10-character card ID as a URL parameter, like this: /authentication/0001230001", false)
			return
		}

		// load up all card data so we can find our card
		cards, err := GetCardsData()
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to read authentication data", true)
			return
		}

		// check if the card's ID matches our given cardID
		card := cards.GetCardByCardID(cardID)
		if card.CardID != cardID {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusUnauthorized, "Card ID is not an authorized card", false)
			return
		}
		if !card.IsValid {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusUnauthorized, "Card ID is not a valid card", false)
			return
		}

		// card is found, get the cardholder's AccountID, RoleID, and PersonID
		accounts, err := GetAccountsData()
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to read accounts data", true)
			return
		}
		people, err := GetPeopleData()
		if err != nil {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to read people data", true)
			return
		}

		// begin to store the output AuthData
		authData := AuthData{CardID: cardID, RoleID: card.RoleID}

		// check if the associated person is valid
		person := people.GetPersonByPersonID(card.PersonID)
		if person.PersonID != card.PersonID {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusUnauthorized, "Card ID is associated with an unknown person", false)
			return
		}
		if !person.IsActive {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusUnauthorized, "Card ID is associated with an inactive person", false)
			return
		}

		// store the personID in the output AuthData
		authData.PersonID = person.PersonID

		// check if the associated account is valid
		account := accounts.GetAccountByAccountID(person.AccountID)
		if account.AccountID != person.AccountID {
			utilities.WriteStringHTTPResponse(writer, req, http.StatusUnauthorized, "Card ID is associated with an unknown account", false)
			return
		}
		if !account.IsActive {
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

		// if err != nil {
		// 	utilities.WriteStringHTTPResponse(writer, req, http.StatusInternalServerError, "Failed to return authentication data properly", true)
		// 	return
		// }

		utilities.WriteJSONHTTPResponse(writer, req, http.StatusOK, authDataJSON, false)
	})
}
