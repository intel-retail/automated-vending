// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"as-controller-board-status/functions"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

const (
	serviceKey = "as-controller-board-status"
)

func main() {
	service, ok := pkg.NewAppService(serviceKey)
	if !ok {
		os.Exit(1)
	}

	lc := service.LoggingClient()

	// Get the application's settings from the configuration.toml file
	appSettings := service.ApplicationSettings()
	if appSettings == nil {
		lc.Error("No application settings found")
		os.Exit(1)
	}

	boardStatus := functions.CheckBoardStatus{
		DoorClosed:    true, // Set default door state to closed
		Configuration: new(functions.ControllerBoardStatusAppSettings),
		Service:       service,
	}

	// Retrieve & parse the required application settings into a proper
	// configuration struct
	err := utilities.MarshalSettings(appSettings, boardStatus.Configuration, false)
	if err != nil {
		lc.Errorf("Application settings could not be processed: %s", err.Error())
		os.Exit(1)
	}

	boardStatus.MaxTemperatureThreshold = boardStatus.Configuration.MaxTemperatureThreshold
	boardStatus.MinTemperatureThreshold = boardStatus.Configuration.MinTemperatureThreshold

	// Create the function pipeline to run when an event is read on the device channels
	err = service.SetFunctionsPipeline(
		transforms.NewFilterFor([]string{boardStatus.Configuration.DeviceName}).FilterByDeviceName,
		boardStatus.CheckControllerBoardStatus,
	)
	if err != nil {
		lc.Errorf("SDK initialization failed: %s", err.Error())
		os.Exit(1)
	}

	// Subscribe to the EdgeX notification service
	err = boardStatus.SubscribeToNotificationService()
	if err != nil {
		lc.Errorf("Error subscribing to EdgeX notification service %s", err.Error())
		os.Exit(1)
	}

	// Add the "status" REST API route
	err = service.AddRoute("/status", functions.GetStatus, "GET", "OPTIONS")
	if err != nil {
		lc.Errorf("Error adding route: %s", err.Error())
		os.Exit(1)
	}

	// Tell the SDK to "start" and begin listening for events to trigger the pipeline
	err = service.MakeItRun()
	if err != nil {
		lc.Errorf("MakeItRun returned error: %s", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
