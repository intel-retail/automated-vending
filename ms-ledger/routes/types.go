// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type Controller struct {
	lc                logger.LoggingClient
	service           interfaces.ApplicationService
	inventoryEndpoint string
}

func NewController(lc logger.LoggingClient, service interfaces.ApplicationService, inventoryEndpoint string) Controller {
	return Controller{
		lc:                lc,
		service:           service,
		inventoryEndpoint: inventoryEndpoint,
	}
}

func (c *Controller) AddAllRoutes() int {
	var err error

	err = c.service.AddRoute("/ledger", c.AllAccountsGet, "OPTIONS", "GET")
	if errorAddRouteHandler(c.lc, err) == 1 {
		return 1
	}

	err = c.service.AddRoute("/ledger/{accountid}", c.LedgerAccountGet, "OPTIONS", "GET")
	if errorAddRouteHandler(c.lc, err) == 1 {
		return 1
	}

	err = c.service.AddRoute("/ledger", c.LedgerAddTransaction, "OPTIONS", "POST")
	if errorAddRouteHandler(c.lc, err) == 1 {
		return 1
	}

	err = c.service.AddRoute("/ledgerPaymentUpdate", c.SetPaymentStatus, "OPTIONS", "POST")
	if errorAddRouteHandler(c.lc, err) == 1 {
		return 1
	}

	err = c.service.AddRoute("/ledger/{accountid}/{tid}", c.LedgerDelete, "DELETE", "OPTIONS")
	if errorAddRouteHandler(c.lc, err) == 1 {
		return 1
	}

	return 0

}

func errorAddRouteHandler(lc logger.LoggingClient, err error) int {
	if err != nil {
		lc.Error("Error adding route: %s", err.Error())
		return 1
	}
	return 0
}
