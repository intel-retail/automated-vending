// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"as-controller-board-status/config"
	"as-controller-board-status/functions"
	"as-controller-board-status/routes"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/transforms"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

const (
	serviceKey = "as-controller-board-status"
)

type boardStatusAppService struct {
	service       interfaces.ApplicationService
	lc            logger.LoggingClient
	serviceConfig *config.ServiceConfig
	boardStatus   functions.CheckBoardStatus
}

func main() {
	app := boardStatusAppService{}
	code := app.CreateAndRunAppService(serviceKey, pkg.NewAppService)
	os.Exit(code)
}

func (app *boardStatusAppService) CreateAndRunAppService(serviceKey string, newServiceFactory func(string) (interfaces.ApplicationService, bool)) int {
	var ok bool
	app.service, ok = newServiceFactory(serviceKey)
	if !ok {
		return 1
	}

	app.lc = app.service.LoggingClient()

	subscriptionClient := app.service.SubscriptionClient()
	if subscriptionClient == nil {
		app.lc.Errorf("error notification service missing from client's configuration")
		return 1
	}

	// retrieve the required configurations
	app.serviceConfig = &config.ServiceConfig{}
	if err := app.service.LoadCustomConfig(app.serviceConfig, "ControllerBoardStatus"); err != nil {
		app.lc.Errorf("failed load custom ControllerBoardStatus configuration: %s", err.Error())
		return 1
	}

	if err := app.serviceConfig.ControllerBoardStatus.Validate(); err != nil {
		app.lc.Errorf("failed to validate ControllerBoardStatus configuration: %v", err)
		return 1
	}

	app.boardStatus.Configuration = &app.serviceConfig.ControllerBoardStatus

	app.boardStatus = functions.CheckBoardStatus{
		DoorClosed:            true, // Set default door state to closed
		Configuration:         app.boardStatus.Configuration,
		SubscriptionClient:    subscriptionClient,
		ControllerBoardStatus: new(functions.ControllerBoardStatus),
	}

	err := app.boardStatus.ParseStringConfigurations()
	if err != nil {
		app.lc.Errorf("failed to parse configs: %v", err)
		return 1
	}

	err = app.boardStatus.SubscribeToNotificationService()
	if err != nil {
		app.lc.Errorf("Error subscribing to EdgeX notification service %s", err.Error())
		return 1
	}

	notificationClient := app.service.NotificationClient()
	if notificationClient == nil {
		app.lc.Error("error notification service missing from client's configuration")
		return 1
	}

	app.boardStatus.NotificationClient = notificationClient

	app.boardStatus.MaxTemperatureThreshold = app.boardStatus.Configuration.MaxTemperatureThreshold
	app.boardStatus.MinTemperatureThreshold = app.boardStatus.Configuration.MinTemperatureThreshold

	app.boardStatus.CommandClient = app.service.CommandClient()
	if app.boardStatus.CommandClient == nil {
		app.lc.Error("error command service missing from client's configuration")
		return 1
	}

	// Create the function pipeline to run when an event is read on the device channels
	err = app.service.SetDefaultFunctionsPipeline(
		transforms.NewFilterFor([]string{app.boardStatus.Configuration.DeviceName}).FilterByDeviceName,
		app.boardStatus.CheckControllerBoardStatus,
	)

	if err != nil {
		app.lc.Errorf("SDK initialization failed: %s", err.Error())
		return 1
	}

	controller := routes.NewController(app.lc, app.service, &app.boardStatus)
	err = controller.AddAllRoutes()
	if err != nil {
		app.lc.Errorf("failed to add all Routes: %s", err.Error())
		return 1
	}

	// Tell the SDK to "start" and begin listening for events to trigger the pipeline
	err = app.service.Run()
	if err != nil {
		app.lc.Errorf("Run returned error: %s", err.Error())
		return 1
	}

	return 0
}
