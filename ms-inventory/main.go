// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"ms-inventory/routes"

	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
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

	var err error
	err = service.AddRoute("/inventory", routes.InventoryGet, "GET")
	errorAddRouteHandler(lc, err)

	err = service.AddRoute("/inventory", routes.InventoryPost, "POST", "OPTIONS")
	errorAddRouteHandler(lc, err)

	err = service.AddRoute("/inventory/delta", routes.DeltaInventorySKUPost, "POST", "OPTIONS")
	errorAddRouteHandler(lc, err)

	err = service.AddRoute("/inventory/{sku}", routes.InventoryItemGet, "GET")
	errorAddRouteHandler(lc, err)

	err = service.AddRoute("/inventory/{sku}", routes.InventoryDelete, "DELETE", "OPTIONS")
	errorAddRouteHandler(lc, err)

	err = service.AddRoute("/auditlog", routes.AuditLogGetAll, "GET", "OPTIONS")
	errorAddRouteHandler(lc, err)

	err = service.AddRoute("/auditlog", routes.AuditLogPost, "POST")
	errorAddRouteHandler(lc, err)

	err = service.AddRoute("/auditlog/{entry}", routes.AuditLogGetEntry, "GET", "OPTIONS")
	errorAddRouteHandler(lc, err)

	err = service.AddRoute("/auditlog/{entry}", routes.AuditLogDelete, "DELETE")
	errorAddRouteHandler(lc, err)

	if err := service.MakeItRun(); err != nil {
		lc.Errorf("MakeItRun returned error: %s", err.Error())
		os.Exit(1)
	}

	// Do any required cleanup here

	os.Exit(0)
}

func errorAddRouteHandler(lc logger.LoggingClient, err error) {
	if err != nil {
		lc.Errorf("Error adding route: %s", err.Error())
		os.Exit(1)
	}
}
