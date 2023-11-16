// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"ms-ledger/routes"
	"net/url"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg"
)

const (
	serviceKey = "ms-ledger"
)

func main() {
	// See https://docs.edgexfoundry.org/2.2/microservices/application/ApplicationServices/
	//       for documentation on application services.
	service, ok := pkg.NewAppService(serviceKey)
	if !ok {
		os.Exit(1)
	}

	lc := service.LoggingClient()

	inventoryEndpoint, err := service.GetAppSetting("InventoryEndpoint")
	if err != nil {
		lc.Errorf("failed load ApplicationSettings: %s", err.Error())
		os.Exit(1)
	}

	if len(inventoryEndpoint) == 0 {
		lc.Error("InventoryEndpoint is not set in ApplicationSettings")
		os.Exit(1)
	}

	if _, err := url.Parse(inventoryEndpoint); err != nil {
		lc.Errorf("InventoryEndpoint from ApplicationSettings is not a valid URL: %s", err.Error())
		os.Exit(1)
	}

	ledgerFileName, err := service.GetAppSetting("LedgerFileName")
	if err != nil {
		lc.Errorf("failed load LedgerFileName from ApplicationSettings: %s", err.Error())
		os.Exit(1)
	}

	if len(ledgerFileName) == 0 {
		lc.Error("LedgerFileName configuration setting is empty")
		os.Exit(1)
	}

	controller := routes.NewController(lc, service, inventoryEndpoint, ledgerFileName)
	err = controller.AddAllRoutes()
	if err != nil {
		lc.Errorf("failed to add all Routes: %s", err.Error())
		os.Exit(1)
	}

	if err := service.Run(); err != nil {
		lc.Errorf("Run returned error: %s", err.Error())
		os.Exit(1)
	}

	os.Exit(0)

}
