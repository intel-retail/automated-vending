// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"errors"

	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

// PeopleFileName is the name of the respective struct data file that
// contains sample authentication data
const PeopleFileName = "people.json"

// AccountsFileName is the name of the respective struct data file that
// contains sample authentication data
const AccountsFileName = "accounts.json"

// CardsFileName is the name of the respective struct data file that
// contains sample authentication data
const CardsFileName = "cards.json"

// WritePeople writes data to the respective JSON file
func (people *People) WritePeople() (err error) {
	return utilities.WriteToJSONFile(PeopleFileName, people, 0644)
}

// WriteAccounts writes data to the respective JSON file
func (accounts *Accounts) WriteAccounts() (err error) {
	return utilities.WriteToJSONFile(AccountsFileName, accounts, 0644)
}

// WriteCards writes data to the respective JSON file
func (cards *Cards) WriteCards() (err error) {
	return utilities.WriteToJSONFile(CardsFileName, cards, 0644)
}

// DeletePeople writes an empty list to the respective JSON file
func DeletePeople() (err error) {
	return utilities.WriteToJSONFile(PeopleFileName, People{People: []Person{}}, 0644)
}

// DeleteAccounts writes an empty list to the respective JSON file
func DeleteAccounts() (err error) {
	return utilities.WriteToJSONFile(AccountsFileName, Accounts{Accounts: []Account{}}, 0644)
}

// DeleteCards writes an empty list to the respective JSON file
func DeleteCards() (err error) {
	return utilities.WriteToJSONFile(CardsFileName, Cards{Cards: []Card{}}, 0644)
}

// DeletePerson deletes from the list
func (people *People) DeletePerson(personToDelete Person) {
	for i, person := range people.People {
		if personToDelete.PersonID == person.PersonID {
			people.People = append(people.People[:i], people.People[i+1:]...)
			return
		}
	}
}

// DeleteAccount deletes from the list
func (accounts *Accounts) DeleteAccount(accountToDelete Account) {
	for i, account := range accounts.Accounts {
		if accountToDelete.AccountID == account.AccountID {
			accounts.Accounts = append(accounts.Accounts[:i], accounts.Accounts[i+1:]...)
			return
		}
	}
}

// DeleteCard deletes from the list
func (cards *Cards) DeleteCard(cardToDelete Card) {
	for i, card := range cards.Cards {
		if cardToDelete.CardID == card.CardID {
			cards.Cards = append(cards.Cards[:i], cards.Cards[i+1:]...)
			return
		}
	}
}

// GetPeopleData reads the data from the respective JSON file
func GetPeopleData() (people People, err error) {
	err = utilities.LoadFromJSONFile(PeopleFileName, &people)
	if err != nil {
		return people, errors.New(
			"Failed to load people from JSON file: " + err.Error(),
		)
	}
	return
}

// GetAccountsData reads the data from the respective JSON file
func GetAccountsData() (accounts Accounts, err error) {
	err = utilities.LoadFromJSONFile(AccountsFileName, &accounts)
	if err != nil {
		return accounts, errors.New(
			"Failed to load accounts from JSON file: " + err.Error(),
		)
	}
	return
}

// GetCardsData reads the data from the respective JSON file
func GetCardsData() (cards Cards, err error) {
	err = utilities.LoadFromJSONFile(CardsFileName, &cards)
	if err != nil {
		return cards, errors.New(
			"Failed to load cards from JSON file: " + err.Error(),
		)
	}
	return
}

// GetPersonByPersonID queries and returns the respective data
func (people *People) GetPersonByPersonID(personID int) (person Person) {
	for _, person := range people.People {
		if person.PersonID == personID {
			return person
		}
	}
	return
}

// GetPersonByAccountID queries and returns the respective data
func (people *People) GetPersonByAccountID(accountID int) (person Person) {
	for _, person := range people.People {
		if person.AccountID == accountID {
			return person
		}
	}
	return
}

// GetPersonByFullName queries and returns the respective data
func (people *People) GetPersonByFullName(fullName string) (person Person) {
	for _, person := range people.People {
		if person.FullName == fullName {
			return person
		}
	}
	return
}

// GetAccountByAccountID queries and returns the respective data
func (accounts *Accounts) GetAccountByAccountID(accountID int) (account Account) {
	for _, account := range accounts.Accounts {
		if account.AccountID == accountID {
			return account
		}
	}
	return
}

// GetAccountByAddress queries and returns the respective data
func (accounts *Accounts) GetAccountByAddress(address string) (account Account) {
	for _, account := range accounts.Accounts {
		if account.Address == address {
			return account
		}
	}
	return
}

// GetAccountByCreditCardNumber queries and returns the respective data
func (accounts *Accounts) GetAccountByCreditCardNumber(creditCardNumber string) (account Account) {
	for _, account := range accounts.Accounts {
		if account.CreditCardNumber == creditCardNumber {
			return account
		}
	}
	return
}

// GetAccountByPhoneNumber queries and returns the respective data
func (accounts *Accounts) GetAccountByPhoneNumber(phoneNumber string) (account Account) {
	for _, account := range accounts.Accounts {
		if account.PhoneNumber == phoneNumber {
			return account
		}
	}
	return
}

// GetAccountByEmailAddress queries and returns the respective data
func (accounts *Accounts) GetAccountByEmailAddress(emailAddress string) (account Account) {
	for _, account := range accounts.Accounts {
		if account.EmailAddress == emailAddress {
			return account
		}
	}
	return
}

// GetCardByCardID queries and returns the respective data
func (cards *Cards) GetCardByCardID(cardID string) (card Card) {
	for _, card := range cards.Cards {
		if card.CardID == cardID {
			return card
		}
	}
	return
}

// GetCardByRoleID queries and returns the respective data
func (cards *Cards) GetCardByRoleID(roleID int) (card Card) {
	for _, card := range cards.Cards {
		if card.RoleID == roleID {
			return card
		}
	}
	return
}

// GetCardByPersonID queries and returns the respective data
func (cards *Cards) GetCardByPersonID(personID int) (card Card) {
	for _, card := range cards.Cards {
		if card.PersonID == personID {
			return card
		}
	}
	return
}
