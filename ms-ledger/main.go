// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"fmt"
	"ms-ledger/routes"
	nethttp "net/http"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
)

const (
	serviceKey = "ms-ledger"
)

func main() {
	// Create an instance of the EdgeX SDK and initialize it.
	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		os.Exit(-1)
	}

	// How to access the application's specific configuration settings.
	appSettings := edgexSdk.ApplicationSettings()
	if appSettings == nil {
		edgexSdk.LoggingClient.Error("No application settings found")
		os.Exit(-1)
	}

	var err error

	err = edgexSdk.AddRoute("/ledger", routes.AllAccountsGet, "OPTIONS", "GET")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/ledger/{accountid}", routes.LedgerAccountGet, "OPTIONS", "GET")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/ledger", addAppSettingsToContext(appSettings, routes.LedgerAddTransaction), "OPTIONS", "POST")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/ledgerPaymentUpdate", routes.SetPaymentStatus, "OPTIONS", "POST")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/ledger/{accountid}/{tid}", routes.LedgerDelete, "DELETE", "OPTIONS")
	errorAddRouteHandler(edgexSdk, err)

	// Tell the SDK to "start" and begin listening for events to trigger the pipeline.
	err = edgexSdk.MakeItRun()
	if err != nil {
		edgexSdk.LoggingClient.Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}

	// Do any required cleanup here

	os.Exit(0)
}

func addAppSettingsToContext(appSettings map[string]string, next func(nethttp.ResponseWriter, *nethttp.Request)) func(nethttp.ResponseWriter, *nethttp.Request) {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		ctx := context.WithValue(r.Context(), routes.AppSettingsKey, appSettings)
		next(w, r.WithContext(ctx))
	}
}

func errorAddRouteHandler(edgexSdk *appsdk.AppFunctionsSDK, err error) {
	if err != nil {
		edgexSdk.LoggingClient.Error("Error adding route: %v", err.Error())
		os.Exit(-1)
	}
}
