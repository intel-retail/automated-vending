// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"fmt"
	"os"
	"strings"

	"as-vending/functions"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
)

const (
	serviceKey = "as-vending"
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

	// Get the device name from the app settings. This could have multiple devices to listen to.
	deviceNameList, ok := appSettings["DeviceName"]
	if !ok {
		edgexSdk.LoggingClient.Error("DeviceName application setting not found")
		os.Exit(-1)
	}

	// Clean up the device list from the toml file and put them in a string array
	deviceNameList = strings.Replace(deviceNameList, " ", "", -1)
	deviceName := strings.Split(deviceNameList, ",")
	edgexSdk.LoggingClient.Info(fmt.Sprintf("Running the application functions for %v devices...", deviceName))

	// Create stop channels for each of the wait threads
	stopChannel := make(chan int)
	doorOpenStopChannel := make(chan int)
	doorCloseStopChannel := make(chan int)
	inferenceStopChannel := make(chan int)

	// Set default values for vending state
	var vendingState functions.VendingState
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

	functions.SetDefaultTimeouts(&vendingState, appSettings, edgexSdk.LoggingClient)

	var err error

	err = edgexSdk.AddRoute("/boardStatus", vendingState.BoardStatus, "POST")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/resetDoorLock", vendingState.ResetDoorLock, "POST")
	errorAddRouteHandler(edgexSdk, err)

	err = edgexSdk.AddRoute("/maintenanceMode", vendingState.GetMaintenanceMode, "GET", "OPTIONS")
	errorAddRouteHandler(edgexSdk, err)

	// Create the function pipeline to run when an event is read on the device channels
	err = edgexSdk.SetFunctionsPipeline(
		transforms.NewFilter(deviceName).FilterByDeviceName,
		vendingState.DeviceHelper,
	)
	if err != nil {
		edgexSdk.LoggingClient.Error("SDK initialization failed: " + err.Error())
		os.Exit(-1)
	}

	// Tell the SDK to "start" and begin listening for events to trigger the pipeline.
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
