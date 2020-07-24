// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"fmt"
	"os"

	"as-vending/functions"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

const (
	serviceKey = "as-vending"
)

func main() {
	// create an instance of the EdgeX SDK and initialize it
	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		os.Exit(-1)
	}

	// get the application settings from configuration.toml
	appSettings := edgexSdk.ApplicationSettings()
	if appSettings == nil {
		edgexSdk.LoggingClient.Error("No application settings found")
		os.Exit(-1)
	}

	var vendingState functions.VendingState
	vendingState.Configuration = new(functions.ServiceConfiguration)

	// retrieve & parse the required application settings into a proper
	// configuration struct
	if err := utilities.MarshalSettings(appSettings, vendingState.Configuration, true); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("Application settings could not be processed: %v", err.Error()))
		os.Exit(-1)
	}

	edgexSdk.LoggingClient.Info(fmt.Sprintf("Running the application functions for %v devices...", vendingState.Configuration.DeviceNames))

	// create stop channels for each of the wait threads
	stopChannel := make(chan int)
	doorOpenStopChannel := make(chan int)
	doorCloseStopChannel := make(chan int)
	inferenceStopChannel := make(chan int)

	// Set default values for vending state
	vendingState.CVWorkflowStarted = false
	vendingState.MaintenanceMode = false
	vendingState.CurrentUserData = functions.OutputData{}
	vendingState.DoorClosed = true
	// global stop channel for threads
	vendingState.ThreadStopChannel = stopChannel
	// open event thread
	vendingState.DoorOpenedDuringCVWorkflow = false
	vendingState.DoorOpenWaitThreadStopChannel = doorOpenStopChannel
	// close event thread
	vendingState.DoorClosedDuringCVWorkflow = false
	vendingState.DoorCloseWaitThreadStopChannel = doorCloseStopChannel
	// inference thread
	vendingState.InferenceDataReceived = false
	vendingState.InferenceWaitThreadStopChannel = inferenceStopChannel

	var err error

	err = edgexSdk.AddRoute("/boardStatus", vendingState.BoardStatus, "POST")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/resetDoorLock", vendingState.ResetDoorLock, "POST")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/maintenanceMode", vendingState.GetMaintenanceMode, "GET", "OPTIONS")
	errorAddRouteHandler(edgexSdk, err)

	// create the function pipeline to run when an event is read on the device channels
	err = edgexSdk.SetFunctionsPipeline(
		transforms.NewFilter(vendingState.Configuration.DeviceNames).FilterByDeviceName,
		vendingState.DeviceHelper,
	)
	if err != nil {
		edgexSdk.LoggingClient.Error("SDK initialization failed: " + err.Error())
		os.Exit(-1)
	}

	// tell the SDK to "start" and begin listening for events to trigger the pipeline.
	err = edgexSdk.MakeItRun()
	if err != nil {
		edgexSdk.LoggingClient.Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}

	// do any required cleanup here

	os.Exit(0)
}

func errorAddRouteHandler(edgexSdk *appsdk.AppFunctionsSDK, err error) {
	if err != nil {
		edgexSdk.LoggingClient.Error("Error adding route: %v", err.Error())
		os.Exit(-1)
	}
}
