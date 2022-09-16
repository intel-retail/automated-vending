// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"ms-inventory/routes"

	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
)

const (
	serviceKey = "ms-inventory"
)

func main() {
	// TODO: See https://docs.edgexfoundry.org/2.2/microservices/application/ApplicationServices/
	//       for documentation on application services.
	service, ok := pkg.NewAppService(serviceKey)
	if !ok {
		os.Exit(1)
	}
	lc := service.LoggingClient()
	controller := routes.NewController(lc, service)
	err := controller.AddAllRoutes()
	if err != nil {
		lc.Errorf("failed to add all Routes: %s", err.Error())
		os.Exit(1)
	}
	if err := service.MakeItRun(); err != nil {
		lc.Errorf("MakeItRun returned error: %s", err.Error())
		os.Exit(1)
	}

	// Do any required cleanup here

	os.Exit(0)
}
