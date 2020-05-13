// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"as-controller-board-status/functions"
	"fmt"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
)

const (
	serviceKey = "as-controller-board-status"
)

func main() {
	// Create an instance of the EdgeX SDK and initialize it
	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	err := edgexSdk.Initialize()
	if err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v", err.Error()))
		os.Exit(-1)
	}

	// Get the application's settings from the configuration.toml file
	appSettings := edgexSdk.ApplicationSettings()
	if appSettings == nil {
		edgexSdk.LoggingClient.Error("No application settings found")
		os.Exit(-1)
	}

	// Retrieve & parse the required application settings into a proper
	// configuration struct
	edgexconfig, err := functions.ProcessApplicationSettings(appSettings)
	if err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("Application settings could not be processed: %v", err.Error()))
		os.Exit(-1)
	}

	boardStatus := functions.CheckBoardStatus{
		MaxTemperatureThreshold: edgexconfig.MaxTemperatureThreshold,
		MinTemperatureThreshold: edgexconfig.MinTemperatureThreshold,
		DoorClosed:              true, // Set default door state to closed
	}

	// Create the function pipeline to run when an event is read on the device channels
	err = edgexSdk.SetFunctionsPipeline(
		transforms.NewFilter([]string{edgexconfig.DeviceName}).FilterByDeviceName,
		boardStatus.CheckControllerBoardStatus,
	)
	if err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v", err.Error()))
		os.Exit(-1)
	}

	// Subscribe to the EdgeX notification service
	err = functions.SubscribeToNotificationService(edgexconfig)
	if err != nil {
		edgexSdk.LoggingClient.Info(fmt.Sprintf("Error subscribing to edgex notification service %s", err.Error()))
		os.Exit(-1)
	}

	// Add the "status" REST API route
	err = edgexSdk.AddRoute("/status", functions.GetStatus, "GET", "OPTIONS")
	if err != nil {
		edgexSdk.LoggingClient.Error("Error adding route: %v", err.Error())
		os.Exit(-1)
	}

	// Tell the SDK to "start" and begin listening for events to trigger the pipeline
	err = edgexSdk.MakeItRun()
	if err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("MakeItRun returned error: %v", err.Error()))
		os.Exit(-1)
	}

	os.Exit(0)
}
