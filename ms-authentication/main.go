// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"fmt"
	"os"

	"ms-authentication/routes"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
)

const (
	serviceKey = "ms-authentication"
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

	if err := edgexSdk.AddRoute("/authentication/{cardid}", routes.AuthenticationGet, "GET"); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("Unable to add /authentication/{cardid} GET route: %v\n", err))
		os.Exit(-1)
	}

	// Tell the SDK to "start" and begin listening for events to trigger the pipeline.
	err := edgexSdk.MakeItRun()
	if err != nil {
		edgexSdk.LoggingClient.Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}

	// Do any required cleanup here

	os.Exit(0)
}
