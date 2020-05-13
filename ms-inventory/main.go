// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"ms-inventory/routes"

	"fmt"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
)

const (
	serviceKey = "ms-inventory"
)

func main() {
	// Create an instance of the EdgeX SDK and initialize it.
	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		os.Exit(-1)
	}

	// Retrieve the application settings from configuration.toml
	appSettings := edgexSdk.ApplicationSettings()
	if appSettings == nil {
		edgexSdk.LoggingClient.Error("No application settings found")
		os.Exit(-1)
	}

	var err error
	err = edgexSdk.AddRoute("/inventory", routes.InventoryGet, "GET")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/inventory", routes.InventoryPost, "POST", "OPTIONS")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/inventory/delta", routes.DeltaInventorySKUPost, "POST", "OPTIONS")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/inventory/{sku}", routes.InventoryItemGet, "GET")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/inventory/{sku}", routes.InventoryDelete, "DELETE", "OPTIONS")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/auditlog", routes.AuditLogGetAll, "GET", "OPTIONS")
	errorAddRouteHandler(edgexSdk, err)
	err = edgexSdk.AddRoute("/auditlog", routes.AuditLogPost, "POST")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/auditlog/{entry}", routes.AuditLogGetEntry, "GET", "OPTIONS")
	errorAddRouteHandler(edgexSdk, err)
	err = edgexSdk.AddRoute("/auditlog/{entry}", routes.AuditLogDelete, "DELETE")
	errorAddRouteHandler(edgexSdk, err)

	// Tell the SDK to "start" and begin listening for events to trigger the pipeline
	err = edgexSdk.MakeItRun()
	if err != nil {
		edgexSdk.LoggingClient.Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}

	// Do any required cleanup here

	os.Exit(0)
}

func errorAddRouteHandler(edgexSdk *appsdk.AppFunctionsSDK, err error) {
	if err != nil {
		edgexSdk.LoggingClient.Error("Error adding route: %v", err.Error())
		os.Exit(-1)
	}
}
