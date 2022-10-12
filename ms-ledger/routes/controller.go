// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type Controller struct {
	lc                logger.LoggingClient
	service           interfaces.ApplicationService
	inventoryEndpoint string
	ledgerFileName    string
}

func NewController(lc logger.LoggingClient, service interfaces.ApplicationService, inventoryEndpoint string, ledgerFileName string) Controller {
	return Controller{
		lc:                lc,
		service:           service,
		inventoryEndpoint: inventoryEndpoint,
		ledgerFileName:    ledgerFileName,
	}
}

func (c *Controller) AddAllRoutes() error {
	var err error

	err = c.service.AddRoute("/ledger", c.AllAccountsGet, "OPTIONS", "GET")
	if errWithMsg := errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/ledger/{accountid}", c.LedgerAccountGet, "OPTIONS", "GET")
	if errWithMsg := errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/ledger", c.LedgerAddTransaction, "OPTIONS", "POST")
	if errWithMsg := errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/ledgerPaymentUpdate", c.SetPaymentStatus, "OPTIONS", "POST")
	if errWithMsg := errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/ledger/{accountid}/{tid}", c.LedgerDelete, "DELETE", "OPTIONS")
	if errWithMsg := errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	return nil

}

func errorAddRouteHandler(err error) error {
	if err != nil {
		return fmt.Errorf("error adding route: %s", err.Error())
	}
	return nil
}
