// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"ms-ledger/routes"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
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

	// serviceConfig := &config.ServiceConfig{}
	// if err := service.LoadCustomConfig(serviceConfig, "AppCustom"); err != nil {
	// 	lc.Errorf("failed load custom configuration: %s", err.Error())
	// 	os.Exit(1)
	// }

	// if err := serviceConfig.AppCustom.Validate(); err != nil {
	// 	lc.Errorf("custom configuration failed validation: %s", err.Error())
	// 	os.Exit(1)
	// }

	inventoryEndpoint, err := service.GetAppSetting("InventoryEndpoint")
	if err != nil {
		lc.Errorf("failed load ApplicationSettings: %s", err.Error())
		os.Exit(1)
	}

	if len(inventoryEndpoint) == 0 {
		lc.Errorf("InventoryEndpoint is not set in ApplicationSettings")
		os.Exit(1)
	}

	controller := routes.NewController(lc, service, inventoryEndpoint)
	if controller.AddAllRoutes() == 1 {
		os.Exit(1)
	}

	if err := service.MakeItRun(); err != nil {
		lc.Errorf("MakeItRun returned error: %s", err.Error())
		os.Exit(1)
	}

	os.Exit(0)

}
