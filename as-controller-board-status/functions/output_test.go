// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"as-controller-board-status/config"
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	client_mocks "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var now = time.Now()

func getCommonApplicationSettingsTyped() *config.ControllerBoardStatusConfig {
	return &config.ControllerBoardStatusConfig{
		AverageTemperatureMeasurementDuration:             "-15s",
		DeviceName:                                        "controller-board",
		MaxTemperatureThreshold:                           temp51,
		MinTemperatureThreshold:                           temp49,
		InferenceDeviceName:                               "Inference-device",
		InferenceDoorStatusCmd:                            "inferenceDoorStatus",
		NotificationCategory:                              "HW_HEALTH",
		NotificationEmailAddresses:                        "test@site.com,test@site.com",
		NotificationLabels:                                "HW_HEALTH",
		NotificationReceiver:                              "System Administrator",
		NotificationSender:                                "Automated Vending Maintenance Notification",
		NotificationSeverity:                              "CRITICAL",
		NotificationName:                                  "maintenance-notification",
		NotificationSubscriptionMaxRESTRetries:            10,
		NotificationSubscriptionRESTRetryIntervalDuration: "10s",
		NotificationThrottleDuration:                      "1m",
		RESTCommandTimeoutDuration:                        "15s",
		VendingEndpoint:                                   "http://localhost:48099/boardStatus",
	}
}

type testTableCheckControllerBoardStatusStruct struct {
	TestCaseName                              string
	InputEdgexContext                         interfaces.AppFunctionContext
	InputData                                 interface{}
	InputCheckBoardStatus                     CheckBoardStatus
	OutputCheckBoardStatus                    CheckBoardStatus
	OutputBool                                bool
	OutputInterface                           interface{}
	OutputLogs                                string
	ShouldLastNotifiedBeDifferent             bool
	ExpectedTemperatureMeasurementSliceLength int
	HTTPTestServer                            *httptest.Server
}

const (
	temp49 = 49.0
	temp50 = 50.0
	temp51 = 51.0
	temp52 = 52.0
)

// The CheckControllerBoardStatus function is the main entrypoint from EdgeX
// into this entire application service. It effectively calls every function
// written, so we can rely on this prepTest function to have curated
// success/error conditions that not only satisfy all of the test cases
// for the CheckControllerBoardStatus function itself, but also
// functions nested deep inside the flow of the function, if possible.
//
// Top-level cases:
// Success case requires:
// - http server that returns 200 for edgexconfig DoorStatusCommandEndpoint,VendingEndpoint
// - http server that returns Accepted edgexconfig NotificationHost
//
// Error conditions:
// x len(params) == 0
// x failure to run ProcessApplicationSettings due to missing config option
// x failure to unmarshal ControllerBoardStatus from event reading
// x failure to call processTemperature, which can be created by sending a
//
// status other than "Accepted" via the NotificationHost
//
// x failure to call processVendingDoorState, which can be created by
//
// sending a status other than status OK to the DoorStatusCommandEndpoint
//
// = 6 test cases total, 3 httptest servers
//
// Nested functions will require more curation:
// x minimum temperature threshold exceeded
// x VendingEndpoint not responding with 200 OK on door close status changes
// x NotificationHost endpoint responding (without error) with non-Accepted
//
// = 11 test cases total
func prepCheckControllerBoardStatusTest() (testTable []testTableCheckControllerBoardStatusStruct, testServers []*httptest.Server) {
	// This server returns 200 OK when hit with a request
	testServerStatusOK := GetHTTPTestServer(http.StatusOK, "")

	// This server throws HTTP status accepted as part of a non-error response
	testServerAccepted := GetHTTPTestServer(http.StatusAccepted, "")

	// This server throws HTTP status 500 as part of a non-error response
	testServer500 := GetHTTPTestServer(http.StatusInternalServerError, "")

	// This server throws errors when it receives a connection
	testServerThrowError := GetErrorHTTPTestServer()

	// Set up a generic EdgeX logger
	lc := logger.NewMockClient()
	correlationID := "test"

	// mock for service
	mockNotificationClient := &client_mocks.NotificationClient{}
	mockNotificationClient.On("SendNotification", mock.Anything, mock.Anything).Return(nil, nil)

	resp := common.BaseResponse{
		StatusCode: http.StatusOK,
	}
	mockCommandClient := &client_mocks.CommandClient{}
	mockCommandClient.On("IssueSetCommandByName", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	resp = common.BaseResponse{
		StatusCode: http.StatusInternalServerError,
	}
	mockErrCommandClient := &client_mocks.CommandClient{}
	mockErrCommandClient.On("IssueSetCommandByName", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)

	// The success condition is ideal, and is configured to use URL's that all
	// respond with responses that correspond to successful scenarios.
	edgexcontextSuccess := pkg.NewAppFuncContextForTest(correlationID, lc)
	configSuccess := getCommonApplicationSettingsTyped()
	configSuccess.VendingEndpoint = testServerStatusOK.URL
	configSuccess.MinTemperatureThreshold = temp51
	configSuccess.MaxTemperatureThreshold = temp49

	// Create the condition of exceeding the minimum temperature threshold,
	// and make everything else successful. Additionally, use a controller
	// board state that has more measurements than the cutoff
	edgexcontextSuccessMinThresholdExceeded := pkg.NewAppFuncContextForTest(correlationID, lc)
	configSuccessMinThresholdExceeded := getCommonApplicationSettingsTyped()
	configSuccessMinThresholdExceeded.VendingEndpoint = testServerStatusOK.URL
	configSuccessMinThresholdExceeded.MinTemperatureThreshold = temp51

	// Create the condition of exceeding the maximum temperature threshold,
	// and make the VendingEndpoint throw an error.
	edgexcontextBadVendingEndpointMaxThresholdExceeded := pkg.NewAppFuncContextForTest(correlationID, lc)
	configBadVendingEndpointMaxThresholdExceeded := getCommonApplicationSettingsTyped()
	configBadVendingEndpointMaxThresholdExceeded.VendingEndpoint = testServerThrowError.URL
	configBadVendingEndpointMaxThresholdExceeded.MaxTemperatureThreshold = temp49

	// Create the condition of exceeding the maximum temperature threshold,
	// and make the NotificationHost endpoint throw something other than what
	// we want. We want Accepted, but we're going to get 500
	edgexcontextUnacceptingNotificationHostMaxThresholdExceeded := pkg.NewAppFuncContextForTest(correlationID, lc)
	configUnacceptingNotificationHostMaxThresholdExceeded := getCommonApplicationSettingsTyped()
	configUnacceptingNotificationHostMaxThresholdExceeded.VendingEndpoint = testServerStatusOK.URL
	configUnacceptingNotificationHostMaxThresholdExceeded.MaxTemperatureThreshold = temp49

	// Create the condition of exceeding the maximum threshold, but also
	// create the condition where the NotificationHost is unreachable,
	// which creates an error condition when attempting to send a "max
	// temperature exceeded" notification
	edgexcontextBadNotificationHostThresholdsExceeded := pkg.NewAppFuncContextForTest(correlationID, lc)
	configBadNotificationHostThresholdsExceeded := getCommonApplicationSettingsTyped()
	configBadNotificationHostThresholdsExceeded.VendingEndpoint = testServerStatusOK.URL
	configBadNotificationHostThresholdsExceeded.MaxTemperatureThreshold = temp49

	// Setup for Device Command (mockErrCommandClient used in one of the below testcases) to throw error to produce specific error conditions
	// in processTemperature, which sends a request to the vending endpoint
	edgexcontextBadDoorStatusCommand := pkg.NewAppFuncContextForTest(correlationID, lc)
	configBadDoorStatusCommand := getCommonApplicationSettingsTyped()
	configBadDoorStatusCommand.VendingEndpoint = testServerStatusOK.URL

	// Set the Vending endpoint to throw error for
	// processTemperature failing to hit the VendingEndpoint
	edgexcontextBadVendingEndpoint := pkg.NewAppFuncContextForTest(correlationID, lc)
	configBadVendingEndpoint := getCommonApplicationSettingsTyped()
	configBadVendingEndpoint.VendingEndpoint = testServerThrowError.URL

	// The expected incoming event reading from the controller board device
	// service looks like this. Humidity and lock values don't matter at this
	// time, since there's no business logic to handle them
	controllerBoardStatusSuccessReadingValue := `{"door_closed":true,"temperature":50.0,"minTemperatureStatus":true,"maxTemperatureStatus":true}`
	controllerBoardStatusSuccessReadingSerialValue := `{\"door_closed\":true,\"temperature\":50.0,\"minTemperatureStatus\":true,\"maxTemperatureStatus\":true}`

	// Following up from the previous 2 lines, the actual event itself (which
	// contains the reading from above) looks like this in the ideal case
	controllerBoardStatusEventSuccess := dtos.Event{
		DeviceName: ControllerBoardDeviceServiceDeviceName,
		Readings: []dtos.BaseReading{
			{
				DeviceName: ControllerBoardDeviceServiceDeviceName,
				SimpleReading: dtos.SimpleReading{
					Value: controllerBoardStatusSuccessReadingValue,
				},
			},
		},
	}

	// Similar to above, create an event that contains a reading with an
	// unmarshalable value
	controllerBoardStatusEventUnsuccessfulJSON := dtos.Event{
		DeviceName: ControllerBoardDeviceServiceDeviceName,
		Readings: []dtos.BaseReading{
			{
				DeviceName: ControllerBoardDeviceServiceDeviceName,
				SimpleReading: dtos.SimpleReading{
					Value: `invalid json value`,
				},
			},
		},
	}

	// The empty input data creates the error condition in
	// CheckControllerBoardStatus that intentionally fails if there is no
	// event contained in the input interface
	var emptyInputData interface{}
	return []testTableCheckControllerBoardStatusStruct{
			{
				TestCaseName:      "Success, no pre-existing measurements, no recent notifications sent",
				InputEdgexContext: edgexcontextSuccess,
				InputData:         controllerBoardStatusEventSuccess,
				InputCheckBoardStatus: CheckBoardStatus{
					MinTemperatureThreshold: temp49,
					MaxTemperatureThreshold: temp51,
					DoorClosed:              true,
					Measurements:            []TempMeasurement{},
					LastNotified:            time.Now().Add(time.Minute * -3),
					Configuration:           configSuccess,
					NotificationClient:      mockNotificationClient,
					CommandClient:           mockCommandClient,
				},
				OutputBool:                    true,
				OutputInterface:               controllerBoardStatusEventSuccess,
				OutputLogs:                    fmt.Sprintf("Received event reading value: %v", controllerBoardStatusSuccessReadingSerialValue),
				ShouldLastNotifiedBeDifferent: true,
				ExpectedTemperatureMeasurementSliceLength: 1,
			},
			{
				TestCaseName:      "Success, minimum temperature threshold exceeded, no recent notifications sent",
				InputEdgexContext: edgexcontextSuccessMinThresholdExceeded,
				InputData:         controllerBoardStatusEventSuccess,
				InputCheckBoardStatus: CheckBoardStatus{
					MinTemperatureThreshold: temp51,
					MaxTemperatureThreshold: temp52,
					DoorClosed:              true,
					Measurements: []TempMeasurement{
						{Timestamp: now.Add(time.Second * time.Duration(-1)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-2)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-3)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-4)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-5)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-17)), Measurement: temp50},
					},
					LastNotified:       time.Now().Add(time.Minute * -3),
					Configuration:      configSuccessMinThresholdExceeded,
					NotificationClient: mockNotificationClient,
				},
				OutputBool:                    true,
				OutputInterface:               controllerBoardStatusEventSuccess,
				OutputLogs:                    fmt.Sprintf("Received event reading value: %v", controllerBoardStatusSuccessReadingSerialValue),
				ShouldLastNotifiedBeDifferent: true,
				ExpectedTemperatureMeasurementSliceLength: 6,
			},
			{
				TestCaseName:      "Unsuccessful due to maximum temperature threshold exceeded, no recent notifications sent, NotificationHost not sending HTTP status Accepted",
				InputEdgexContext: edgexcontextUnacceptingNotificationHostMaxThresholdExceeded,
				InputData:         controllerBoardStatusEventSuccess,
				InputCheckBoardStatus: CheckBoardStatus{
					MinTemperatureThreshold: temp51,
					MaxTemperatureThreshold: temp49,
					DoorClosed:              true,
					Measurements: []TempMeasurement{
						{Timestamp: now.Add(time.Second * time.Duration(-1)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-2)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-3)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-4)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-5)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-17)), Measurement: temp50},
					},
					LastNotified:       time.Now().Add(time.Minute * -3),
					Configuration:      configUnacceptingNotificationHostMaxThresholdExceeded,
					NotificationClient: mockNotificationClient,
				},
				OutputBool:                    true,
				OutputInterface:               controllerBoardStatusEventSuccess,
				OutputLogs:                    fmt.Sprintf("Encountered error while checking temperature thresholds: Failed to send temperature threshold exceeded notification(s) due to error: Encountered error sending the %v temperature threshold message: Encountered error sending notification for exceeding temperature threshold: The REST API HTTP status code response from the server when attempting to send a notification was not %v, instead got: %v", maximum, http.StatusAccepted, http.StatusInternalServerError),
				ShouldLastNotifiedBeDifferent: false,
				ExpectedTemperatureMeasurementSliceLength: 6,
			},
			{
				TestCaseName:      "Unsuccessful due to empty params",
				InputEdgexContext: edgexcontextSuccess,
				InputData:         emptyInputData,
				InputCheckBoardStatus: CheckBoardStatus{
					Configuration:      configSuccess,
					NotificationClient: mockNotificationClient,
				},
				OutputBool:      false,
				OutputInterface: nil,
				OutputLogs:      "",
			},
			{
				TestCaseName:      "Unsuccessful due to unserializable controller board status data",
				InputEdgexContext: edgexcontextSuccess,
				InputData:         controllerBoardStatusEventUnsuccessfulJSON,
				InputCheckBoardStatus: CheckBoardStatus{
					Configuration:      configSuccess,
					NotificationClient: mockNotificationClient,
				},
				OutputBool:      false,
				OutputInterface: nil,
				OutputLogs:      `Failed to unmarshal controller board data, the event data is: {\"value\":\"invalid json value\"}`,
			},
			{
				TestCaseName:      "Unsuccessful due to NotificationHost not responding with HTTP Accepted",
				InputEdgexContext: edgexcontextBadNotificationHostThresholdsExceeded,
				InputData:         controllerBoardStatusEventSuccess,
				InputCheckBoardStatus: CheckBoardStatus{
					LastNotified:       time.Now().Add(time.Minute * -3),
					Configuration:      configBadNotificationHostThresholdsExceeded,
					NotificationClient: mockNotificationClient,
					CommandClient:      mockCommandClient,
				},
				OutputBool:      true,
				OutputInterface: controllerBoardStatusEventSuccess,
				OutputLogs: fmt.Sprintf(
					`Encountered error while checking temperature thresholds: Failed to send temperature threshold exceeded notification(s) due to error: Encountered error sending the maximum temperature threshold message: Encountered error sending notification for exceeding temperature threshold: Failed to perform REST POST API call to send a notification to , error: %v : %v`, "Post", "EOF"),
				ShouldLastNotifiedBeDifferent:             false,
				ExpectedTemperatureMeasurementSliceLength: 1,
			},
			{
				TestCaseName:      "Unsuccessful due to Device Command not responding with HTTP 200 OK, no temperature notification sent",
				InputEdgexContext: edgexcontextBadDoorStatusCommand,
				InputData:         controllerBoardStatusEventSuccess,
				InputCheckBoardStatus: CheckBoardStatus{
					LastNotified:       time.Now().Add(time.Minute * -3),
					Configuration:      configBadDoorStatusCommand,
					NotificationClient: mockNotificationClient,
					CommandClient:      mockErrCommandClient,
				},
				OutputBool:                    true,
				OutputInterface:               controllerBoardStatusEventSuccess,
				OutputLogs:                    fmt.Sprintf("Encountered error while checking the open/closed state of the door: failed to submit the vending door state to the device command: Failed to submit REST PUT request due to error: %v : %v", "Put", "EOF"),
				ShouldLastNotifiedBeDifferent: false,
				ExpectedTemperatureMeasurementSliceLength: 1,
			},
			{
				TestCaseName:      "Unsuccessful due to VendingEndpoint not responding with HTTP 200 OK, no temperature notification sent",
				InputEdgexContext: edgexcontextBadVendingEndpoint,
				InputData:         controllerBoardStatusEventSuccess,
				InputCheckBoardStatus: CheckBoardStatus{
					LastNotified:       time.Now().Add(time.Minute * -3),
					Configuration:      configBadVendingEndpoint,
					NotificationClient: mockNotificationClient,
				},
				OutputBool:                    true,
				OutputInterface:               controllerBoardStatusEventSuccess,
				OutputLogs:                    fmt.Sprintf("Encountered error while checking the open/closed state of the door: failed to submit the controller board's status to the central vending state service: Failed to submit REST POST request due to error: %v \\\"%v\\\": %v", "Post", configBadVendingEndpoint.VendingEndpoint, "EOF"),
				ShouldLastNotifiedBeDifferent: false,
				ExpectedTemperatureMeasurementSliceLength: 1,
			},
			{
				TestCaseName:      "Unsuccessful due to VendingEndpoint not responding with HTTP 200 OK, max temperature threshold exceeded",
				InputEdgexContext: edgexcontextBadVendingEndpointMaxThresholdExceeded,
				InputData:         controllerBoardStatusEventSuccess,
				InputCheckBoardStatus: CheckBoardStatus{
					MinTemperatureThreshold: temp51,
					MaxTemperatureThreshold: temp49,
					DoorClosed:              true,
					Measurements: []TempMeasurement{
						{Timestamp: now.Add(time.Second * time.Duration(-1)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-2)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-3)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-4)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-5)), Measurement: temp50},
						{Timestamp: now.Add(time.Second * time.Duration(-17)), Measurement: temp50},
					},
					LastNotified:       time.Now().Add(time.Minute * -3),
					Configuration:      configBadVendingEndpointMaxThresholdExceeded,
					NotificationClient: mockNotificationClient,
				},
				OutputBool:                    true,
				OutputInterface:               controllerBoardStatusEventSuccess,
				OutputLogs:                    fmt.Sprintf("Encountered error while checking temperature thresholds: Encountered error sending the controller board's status to the central vending endpoint: Failed to submit REST POST request due to error: %v \\\"%v\\\": %v", "Post", configBadVendingEndpointMaxThresholdExceeded.VendingEndpoint, "EOF"),
				ShouldLastNotifiedBeDifferent: true,
				ExpectedTemperatureMeasurementSliceLength: 6,
			},
		}, []*httptest.Server{
			testServerStatusOK,
			testServerThrowError,
			testServerAccepted,
			testServer500,
		}
}

// TestCheckControllerBoardStatus validates that the
// CheckControllerBoardStatus function behaves as expected
func TestCheckControllerBoardStatus(t *testing.T) {
	testTable, testServers := prepCheckControllerBoardStatusTest()
	for _, testCase := range testTable {
		ct := testCase // pinning "current test" solves concurrency issues
		t.Run(testCase.TestCaseName, func(t *testing.T) {
			// Set up the test assert functions
			assert := assert.New(t)

			// this needs to be set for upgraded config value
			err := ct.InputCheckBoardStatus.ParseStringConfigurations()
			require.NoError(t, err)
			// Run the function that needs to be tested
			testBool, testInterface := ct.InputCheckBoardStatus.CheckControllerBoardStatus(ct.InputEdgexContext, ct.InputData)

			// Validate the output
			assert.Equal(ct.OutputBool, testBool, "Expected output boolean to match test boolean")
			assert.Equal(ct.OutputInterface, testInterface, "Expected output interface to match test interface")
			assert.Equal(ct.ExpectedTemperatureMeasurementSliceLength, len(ct.InputCheckBoardStatus.Measurements), "The number of temperature measurements contained in CheckBoardStatus should match expected, since the AvgTemp function should be purging old values")
			if ct.ShouldLastNotifiedBeDifferent {
				assert.NotEqual(ct.ShouldLastNotifiedBeDifferent, ct.InputCheckBoardStatus.LastNotified, "Expected the CheckBoardStatus.LastNotified value to be different")
			}
		})
	}
	// We are responsible for closing the test servers
	for _, testServer := range testServers {
		testServer.Close()
	}
}

// Finish by testing the remaining edge cases that were not covered by
// the above test table

// TestGetTempThresholdExceededMessage tests that the edge case for
// getTempThresholdExceededMessage returns an error when passed in a value
// other than "maximum" or "minimum"
func TestGetTempThresholdExceededMessage(t *testing.T) {
	assert := assert.New(t)

	tval := "neither min nor max"
	result, err := getTempThresholdExceededMessage(tval, temp50, temp50)

	assert.EqualError(err, fmt.Sprintf("Please specify minOrMax as \"%v\" or \"%v\", the value given was \"%v\"", maximum, minimum, tval))
	assert.Empty(result, "Expected error result to be an empty string")
}
