// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"os"
	"strings"

	"as-vending/config"
	"as-vending/functions"
	"as-vending/routes"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	serviceKey = "as-vending"
)

type vendingAppService struct {
	service       interfaces.ApplicationService
	lc            logger.LoggingClient
	serviceConfig *config.ServiceConfig
	vendingState  functions.VendingState
}

func main() {
	app := vendingAppService{}
	code := app.CreateAndRunAppService(serviceKey, pkg.NewAppService)
	os.Exit(code)
}

func (app *vendingAppService) CreateAndRunAppService(serviceKey string, newServiceFactory func(string) (interfaces.ApplicationService, bool)) int {
	var ok bool
	app.service, ok = newServiceFactory(serviceKey)
	if !ok {
		return 1
	}

	app.lc = app.service.LoggingClient()

	// retrieve the required configurations
	app.serviceConfig = &config.ServiceConfig{}
	if err := app.service.LoadCustomConfig(app.serviceConfig, "Vending"); err != nil {
		app.lc.Errorf("failed load custom Vending configuration: %s", err.Error())
		return 1
	}

	if err := app.serviceConfig.Vending.Validate(); err != nil {
		app.lc.Errorf("failed to validate Vending configuration: %v", err)
		return 1
	}

	app.vendingState.Configuration = &app.serviceConfig.Vending
	// parse configuration durations to a time.Duration object
	if err := app.vendingState.ParseDurationFromConfig(); err != nil {
		app.lc.Errorf("failed to parse configuration: %v", err)
		return 1
	}

	app.vendingState.CommandClient = app.service.CommandClient()
	if app.vendingState.CommandClient == nil {
		app.lc.Error("Error command service missing from client's configuration")
		return 1
	}

	app.lc.Infof("running the application functions for %s devices...", app.vendingState.Configuration.DeviceNames)

	// create stop channels for each of the wait threads
	stopChannel := make(chan int)
	doorOpenStopChannel := make(chan int)
	doorCloseStopChannel := make(chan int)
	inferenceStopChannel := make(chan int)

	// Set default values for vending state
	app.vendingState.CVWorkflowStarted = false
	app.vendingState.MaintenanceMode = false
	app.vendingState.CurrentUserData = functions.OutputData{}
	app.vendingState.DoorClosed = true
	// global stop channel for threads
	app.vendingState.ThreadStopChannel = stopChannel
	// open event thread
	app.vendingState.DoorOpenedDuringCVWorkflow = false
	app.vendingState.DoorOpenWaitThreadStopChannel = doorOpenStopChannel
	// close event thread
	app.vendingState.DoorClosedDuringCVWorkflow = false
	app.vendingState.DoorCloseWaitThreadStopChannel = doorCloseStopChannel
	// inference thread
	app.vendingState.InferenceDataReceived = false
	app.vendingState.InferenceWaitThreadStopChannel = inferenceStopChannel

	controller := routes.NewController(app.lc, app.service, app.vendingState)
	err := controller.AddAllRoutes()
	if err != nil {
		app.lc.Errorf("failed to add all Routes: %s", err.Error())
		return 1
	}

	deviceNames := strings.Split(app.vendingState.Configuration.DeviceNames, ",")
	// create the function pipeline to run when an event is read on the device channels
	err = app.service.SetFunctionsPipeline(
		transforms.NewFilterFor(deviceNames).FilterByDeviceName,
		app.vendingState.DeviceHelper,
	)
	if err != nil {
		app.lc.Errorf("SDK initialization failed: %s", err.Error())
		return 1
	}

	// tell the SDK to "start" and begin listening for events to trigger the pipeline.
	err = app.service.MakeItRun()
	if err != nil {
		app.lc.Errorf("MakeItRun returned error: %s", err.Error())
		return 1
	}

	// do any required cleanup here

	return 0
}
