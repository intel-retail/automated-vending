// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

// Cards is a struct that simply holds a list of Cards (things used to
// authenticate a person for a role)
type Cards struct {
	Cards []Card `json:"cards"`
}

// People is a struct that simply holds a list of people
type People struct {
	People []Person `json:"people"`
}

// Accounts is a struct that simply holds a list of accounts
type Accounts struct {
	Accounts []Account `json:"accounts"`
}

// Card contains role, person and card associations. A person can have multiple
// cards, but a card can only be associated with one role. If a person needs to
// take on a different role, they must use a different card with the desired
// role
type Card struct {
	CardID    string `json:"cardID"`
	RoleID    int    `json:"roleID"`
	IsValid   bool   `json:"isValid"`
	PersonID  int    `json:"personID"`
	CreatedAt int64  `json:"createdAt,string"`
	UpdatedAt int64  `json:"updatedAt,string"`
}

// Person contains person, account, and full name associations. A person
// is associated to one account, and multiple people can be associated to one
// account
type Person struct {
	PersonID  int    `json:"personID"`
	AccountID int    `json:"accountID"`
	FullName  string `json:"fullName"`
	CreatedAt int64  `json:"createdAt,string"`
	UpdatedAt int64  `json:"updatedAt,string"`
	IsActive  bool   `json:"isActive"`
}

// Account contains payment and billing information. Multiple people can
// be associated with a single account
type Account struct {
	AccountID        int    `json:"accountID"`
	Address          string `json:"address"`
	CreditCardNumber string `json:"creditCardNumber"`
	PhoneNumber      string `json:"phoneNumber"`
	EmailAddress     string `json:"emailAddress"`
	CreatedAt        int64  `json:"createdAt,string"`
	UpdatedAt        int64  `json:"updatedAt,string"`
	IsActive         bool   `json:"isActive"`
}

// AuthData is what is expected to be sent back as a response when something
// hits this endpoint. A card number is passed in, and this code will
// resolve the card's corresponding role, person, and account
type AuthData struct {
	AccountID int    `json:"accountID"`
	PersonID  int    `json:"personID"`
	RoleID    int    `json:"roleID"`
	CardID    string `json:"cardID"`
}
