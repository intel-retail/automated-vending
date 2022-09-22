// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"as-controller-board-status/functions"
	"as-controller-board-status/routes"
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

	subscriptionClient := service.SubscriptionClient()
	if subscriptionClient == nil {
		lc.Errorf("error notification service missing from client's configuration")
		os.Exit(1)
	}

	config := new(functions.ControllerBoardStatusAppSettings)
	// Retrieve & parse the required application settings into a proper
	// configuration struct
	err := utilities.MarshalSettings(appSettings, config, false)
	if err != nil {
		lc.Errorf("Application settings could not be processed: %s", err.Error())
		os.Exit(1)
	}

	boardStatus := functions.CheckBoardStatus{
		DoorClosed:            true, // Set default door state to closed
		Configuration:         config,
		SubscriptionClient:    subscriptionClient,
		ControllerBoardStatus: new(functions.ControllerBoardStatus),
	}

	err = boardStatus.SubscribeToNotificationService()
	if err != nil {
		lc.Errorf("Error subscribing to EdgeX notification service %s", err.Error())
		os.Exit(1)
	}

	notificationClient := service.NotificationClient()
	if notificationClient == nil {
		lc.Error("error notification service missing from client's configuration")
		os.Exit(1)
	}
	boardStatus.NotificationClient = notificationClient

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

	controller := routes.NewController(lc, service, &boardStatus)
	err = controller.AddAllRoutes()
	if err != nil {
		lc.Errorf("failed to add all Routes: %s", err.Error())
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
