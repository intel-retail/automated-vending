// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"ms-ledger/config"
	"ms-ledger/routes"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	serviceKey = "ms-ledger"
)

func main() {
	// TODO: See https://docs.edgexfoundry.org/2.2/microservices/application/ApplicationServices/
	//       for documentation on application services.
	service, ok := pkg.NewAppService(serviceKey)
	if !ok {
		os.Exit(1)
	}

	lc := service.LoggingClient()

	serviceConfig := &config.ServiceConfig{}
	if err := service.LoadCustomConfig(serviceConfig, "AppCustom"); err != nil {
		lc.Errorf("failed load custom configuration: %s", err.Error())
		os.Exit(1)
	}

	if err := serviceConfig.AppCustom.Validate(); err != nil {
		lc.Errorf("custom configuration failed validation: %s", err.Error())
		os.Exit(1)
	}

	route := routes.NewRoute(lc, serviceConfig)

	var err error

	err = service.AddRoute("/ledger", route.AllAccountsGet, "OPTIONS", "GET")
	errorAddRouteHandler(lc, err)

	err = service.AddRoute("/ledger/{accountid}", route.LedgerAccountGet, "OPTIONS", "GET")
	errorAddRouteHandler(lc, err)

	err = service.AddRoute("/ledger", route.LedgerAddTransaction, "OPTIONS", "POST")
	errorAddRouteHandler(lc, err)

	err = service.AddRoute("/ledgerPaymentUpdate", route.SetPaymentStatus, "OPTIONS", "POST")
	errorAddRouteHandler(lc, err)

	err = service.AddRoute("/ledger/{accountid}/{tid}", route.LedgerDelete, "DELETE", "OPTIONS")
	errorAddRouteHandler(lc, err)

	if err := service.MakeItRun(); err != nil {
		lc.Errorf("MakeItRun returned error: %s", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

func errorAddRouteHandler(lc logger.LoggingClient, err error) {
	if err != nil {
		lc.Error("Error adding route: %s", err.Error())
		os.Exit(1)
	}
}
