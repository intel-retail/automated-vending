// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

var now = time.Now()

type testGetStatusStruct struct {
	TestCaseName          string
	ControllerBoardStatus ControllerBoardStatus
	OutputHTTPResponse    utilities.HTTPResponse
	RESTMethod            string
	RESTURL               string
}

func prepGetStatusTest() ([]testGetStatusStruct, error) {
	testControllerBoardStatus := ControllerBoardStatus{}
	controllerBoardStatusJSON, err := utilities.GetAsJSON(testControllerBoardStatus)
	if err != nil {
		return []testGetStatusStruct{}, err
	}
	return []testGetStatusStruct{
		{
			TestCaseName:          "Success",
			ControllerBoardStatus: testControllerBoardStatus,
			OutputHTTPResponse: utilities.HTTPResponse{
				Content:     controllerBoardStatusJSON,
				ContentType: "json",
				StatusCode:  200,
				Error:       false,
			},
			RESTMethod: "GET",
			RESTURL:    "/status",
		},
	}, nil
}

// TestGetStatus validates that the GetStatus function
// properly handles all error and success scenarios
func TestGetStatus(t *testing.T) {
	testTable := []testGetStatusStruct{}
	err := fmt.Errorf("")
	t.Run("GetStatus test setup", func(t *testing.T) {
		assert := assert.New(t)
		testTable, err = prepGetStatusTest()

		assert.NoError(err, "Failed to set up test")
	})
	for _, testCase := range testTable {
		ct := testCase // pinning to avoid concurrency issues
		t.Run(ct.TestCaseName, func(t *testing.T) {
			assert := assert.New(t)

			req := httptest.NewRequest(ct.RESTMethod, ct.RESTURL, nil)
			w := httptest.NewRecorder()
			GetStatus(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(err, "Failed to read response body")

			// Prepare to unmarshal the response body into the helpers struct
			responseContent := utilities.HTTPResponse{}
			err = json.Unmarshal(body, &responseContent)

			assert.NoError(err, "Failed to unmarshal response body into HTTPResponse struct")
			assert.Equal(ct.OutputHTTPResponse, responseContent)
		})
	}
}

func getCommonApplicationSettings() map[string]string {
	return map[string]string{
		AverageTemperatureMeasurementDuration:     "-15s",
		DeviceName:                                "ds-controller-board",
		MaxTemperatureThreshold:                   temp51s,
		MinTemperatureThreshold:                   temp49s,
		MQTTEndpoint:                              "http://localhost:48082/api/v1/device/name/Inference-MQTT-device/command/vendingDoorStatus",
		NotificationCategory:                      "HW_HEALTH",
		NotificationEmailAddresses:                "test@site.com,test@site.com",
		NotificationHost:                          "http://localhost:48060/api/v1/notification",
		NotificationLabels:                        "HW_HEALTH",
		NotificationReceiver:                      "System Administrator",
		NotificationSender:                        "Automated Checkout Maintenance Notification",
		NotificationSeverity:                      "CRITICAL",
		NotificationSlug:                          "sys-admin",
		NotificationSlugPrefix:                    "maintenance-notification",
		NotificationSubscriptionMaxRESTRetries:    "10",
		NotificationSubscriptionRESTRetryInterval: "10s",
		NotificationThrottleDuration:              "1m",
		RESTCommandTimeout:                        "15s",
		SubscriptionHost:                          "http://localhost:48060/api/v1/subscription",
		VendingEndpoint:                           "http://localhost:48099/boardStatus",
	}
}

type testTableCheckControllerBoardStatusStruct struct {
	TestCaseName                              string
	InputEdgexContext                         *appcontext.Context
	InputParams                               []interface{}
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
	checkControllerBoardStatusLogFileName = "./test_CheckControllerBoardStatus.log"
	temp49                                = 49.0
	temp50                                = 50.0
	temp51                                = 51.0
	temp52                                = 52.0
	temp49s                               = "49.0"
	// temp50s                               = "50.0"
	temp51s = "51.0"
	// temp52s                               = "52.0"
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
// - http server that returns 200 for edgexconfig MQTTEndpoint,VendingEndpoint
// - http server that returns Accepted edgexconfig NotificationHost
//
// Error conditions:
// x len(params) == 0
// x failure to run ProcessApplicationSettings due to missing config option
// x failure to unmarshal ControllerBoardStatus from event reading
// x failure to call processTemperature, which can be created by sending a
//   status other than "Accepted" via the NotificationHost
// x failure to call processVendingDoorState, which can be created by
//   sending a status other than status OK to the MQTTEndpoint
//
// = 6 test cases total, 3 httptest servers
//
// Nested functions will require more curation:
// x minimum temperature threshold exceeded
// x VendingEndpoint not responding with 200 OK on door close status changes
// x NotificationHost endpoint responding (without error) with non-Accepted
//
// = 11 test cases total
func prepCheckControllerBoardStatusTest() ([]testTableCheckControllerBoardStatusStruct, []*httptest.Server) {
	// This server returns 200 OK when hit with a request
	testServerStatusOK := GetHTTPTestServer(http.StatusOK, "")

	// This server throws HTTP status accepted as part of a non-error response
	testServerAccepted := GetHTTPTestServer(http.StatusAccepted, "")

	// This server throws HTTP status 500 as part of a non-error response
	testServer500 := GetHTTPTestServer(http.StatusInternalServerError, "")

	// This server throws errors when it receives a connection
	testServerThrowError := GetErrorHTTPTestServer()

	// Set up a generic EdgeX logger
	lc := logger.NewClient("output_test", false, checkControllerBoardStatusLogFileName, "DEBUG")

	// The success condition is ideal, and is configured to use URL's that all
	// respond with responses that correspond to successful scenarios.
	edgexcontextSuccess := &appcontext.Context{
		LoggingClient: lc,
	}
	edgexcontextSuccess.Configuration.ApplicationSettings = getCommonApplicationSettings()
	edgexcontextSuccess.Configuration.ApplicationSettings[MQTTEndpoint] = testServerStatusOK.URL
	edgexcontextSuccess.Configuration.ApplicationSettings[VendingEndpoint] = testServerStatusOK.URL
	edgexcontextSuccess.Configuration.ApplicationSettings[NotificationHost] = testServerAccepted.URL
	edgexcontextSuccess.Configuration.ApplicationSettings[MinTemperatureThreshold] = temp51s
	edgexcontextSuccess.Configuration.ApplicationSettings[MaxTemperatureThreshold] = temp49s

	// Create the condition of exceeding the minimum temperature threshold,
	// and make everything else successful. Additionally, use a controller
	// board state that has more measurements than the cutoff
	edgexcontextSuccessMinThresholdExceeded := &appcontext.Context{
		LoggingClient: lc,
	}
	edgexcontextSuccessMinThresholdExceeded.Configuration.ApplicationSettings = getCommonApplicationSettings()
	edgexcontextSuccessMinThresholdExceeded.Configuration.ApplicationSettings[MQTTEndpoint] = testServerStatusOK.URL
	edgexcontextSuccessMinThresholdExceeded.Configuration.ApplicationSettings[VendingEndpoint] = testServerStatusOK.URL
	edgexcontextSuccessMinThresholdExceeded.Configuration.ApplicationSettings[NotificationHost] = testServerAccepted.URL
	edgexcontextSuccessMinThresholdExceeded.Configuration.ApplicationSettings[MinTemperatureThreshold] = temp51s

	// Create the condition of exceeding the maximum temperature threshold,
	// and make the VendingEndpoint throw an error.
	edgexcontextBadVendingEndpointMaxThresholdExceeded := &appcontext.Context{
		LoggingClient: lc,
	}
	edgexcontextBadVendingEndpointMaxThresholdExceeded.Configuration.ApplicationSettings = getCommonApplicationSettings()
	edgexcontextBadVendingEndpointMaxThresholdExceeded.Configuration.ApplicationSettings[MQTTEndpoint] = testServerStatusOK.URL
	edgexcontextBadVendingEndpointMaxThresholdExceeded.Configuration.ApplicationSettings[VendingEndpoint] = testServerThrowError.URL
	edgexcontextBadVendingEndpointMaxThresholdExceeded.Configuration.ApplicationSettings[NotificationHost] = testServerAccepted.URL
	edgexcontextBadVendingEndpointMaxThresholdExceeded.Configuration.ApplicationSettings[MaxTemperatureThreshold] = temp49s

	// Create the condition of exceeding the maximum temperature threshold,
	// and make the NotificationHost endpoint throw something other than what
	// we want. We want Accepted, but we're going to get 500
	edgexcontextUnacceptingNotificationHostMaxThresholdExceeded := &appcontext.Context{
		LoggingClient: lc,
	}
	edgexcontextUnacceptingNotificationHostMaxThresholdExceeded.Configuration.ApplicationSettings = getCommonApplicationSettings()
	edgexcontextUnacceptingNotificationHostMaxThresholdExceeded.Configuration.ApplicationSettings[MQTTEndpoint] = testServerStatusOK.URL
	edgexcontextUnacceptingNotificationHostMaxThresholdExceeded.Configuration.ApplicationSettings[VendingEndpoint] = testServerStatusOK.URL
	edgexcontextUnacceptingNotificationHostMaxThresholdExceeded.Configuration.ApplicationSettings[NotificationHost] = testServer500.URL
	edgexcontextUnacceptingNotificationHostMaxThresholdExceeded.Configuration.ApplicationSettings[MaxTemperatureThreshold] = temp49s

	// Create the condition of exceeding the maximum threshold, but also
	// create the condition where the NotificationHost is unreachable,
	// which creates an error condition when attempting to send a "max
	// temperature exceeded" notification
	edgexcontextBadNotificationHostThresholdsExceeded := &appcontext.Context{
		LoggingClient: lc,
	}
	edgexcontextBadNotificationHostThresholdsExceeded.Configuration.ApplicationSettings = getCommonApplicationSettings()
	edgexcontextBadNotificationHostThresholdsExceeded.Configuration.ApplicationSettings[NotificationHost] = testServerThrowError.URL
	edgexcontextBadNotificationHostThresholdsExceeded.Configuration.ApplicationSettings[MaxTemperatureThreshold] = temp49s

	// Set bad MQTT and Vending endpoints to produce specific error conditions
	// in processTemperature, which first sends a request to MQTT, then
	// another request to the vending endpoint
	edgexcontextBadMQTTEndpoint := &appcontext.Context{
		LoggingClient: lc,
	}
	edgexcontextBadMQTTEndpoint.Configuration.ApplicationSettings = getCommonApplicationSettings()
	edgexcontextBadMQTTEndpoint.Configuration.ApplicationSettings[MQTTEndpoint] = testServerThrowError.URL

	// As described above, in order to produce the error condition for
	// processTemperature failing to hit the VendingEndpoint, we have to hit
	// the MQTTEndpoint successfully first
	edgexcontextBadVendingEndpoint := &appcontext.Context{
		LoggingClient: lc,
	}
	edgexcontextBadVendingEndpoint.Configuration.ApplicationSettings = getCommonApplicationSettings()
	edgexcontextBadVendingEndpoint.Configuration.ApplicationSettings[MQTTEndpoint] = testServerStatusOK.URL
	edgexcontextBadVendingEndpoint.Configuration.ApplicationSettings[VendingEndpoint] = testServerThrowError.URL

	// By not properly specifying valid application settings values, we create
	// the error condition in ProcessApplicationSettings that claims
	// the passed-in configuration is invalid
	edgexcontextBadApplicationSettings := &appcontext.Context{
		LoggingClient: lc,
	}
	edgexcontextBadApplicationSettings.Configuration.ApplicationSettings = map[string]string{}

	// The expected incoming event reading from the controller board device
	// service looks like this. Humidity and lock values don't matter at this
	// time, since there's no business logic to handle them
	controllerBoardStatusSuccessReadingValue := `{"door_closed":true,"temperature":50.0,"minTemperatureStatus":true,"maxTemperatureStatus":true}`
	controllerBoardStatusSuccessReadingSerialValue := `{\"door_closed\":true,\"temperature\":50.0,\"minTemperatureStatus\":true,\"maxTemperatureStatus\":true}`

	// Following up from the previous 2 lines, the actual event itself (which
	// contains the reading from above) looks like this in the ideal case
	controllerBoardStatusEventSuccess := models.Event{
		Device: ControllerBoardDeviceServiceDeviceName,
		Readings: []models.Reading{
			{
				Value: controllerBoardStatusSuccessReadingValue,
			},
		},
	}

	// We have to put the event into a slice, because the
	// CheckControllerBoardStatus expects variadic interface parameters
	controllerBoardStatusEventSuccessSlice := []interface{}{
		controllerBoardStatusEventSuccess,
	}

	// Similar to above, create an event that contains a reading with an
	// unmarshalable value
	controllerBoardStatusEventUnsuccessfulJSON := models.Event{
		Device: ControllerBoardDeviceServiceDeviceName,
		Readings: []models.Reading{
			{
				Value: `invalid json value`,
			},
		},
	}

	// Similar to above, put the event into a slice because it is a
	// variadic interface parameter to CheckControllerBoardStatus
	controllerBoardStatusEventUnsuccessfulJSONSlice := []interface{}{
		controllerBoardStatusEventUnsuccessfulJSON,
	}

	// The empty input parameters slice creates the error condition in
	// CheckControllerBoardStatus that intentionally fails if there is no
	// event contained in the input interface slice
	emptyInputParamsSlice := []interface{}{}

	// The initial state of the board needs to be controlled. In order to
	// test the nested functions that will send notifications, we have to
	// set a specific minimum amount of time in the past for the value of
	// LastNotified
	checkBoardStatusInitial := CheckBoardStatus{
		LastNotified: time.Now().Add(time.Minute * -3),
	}

	return []testTableCheckControllerBoardStatusStruct{
			{
				TestCaseName:      "Success, no pre-existing measurements, no recent notifications sent",
				InputEdgexContext: edgexcontextSuccess,
				InputParams:       controllerBoardStatusEventSuccessSlice,
				InputCheckBoardStatus: CheckBoardStatus{
					MinTemperatureThreshold: temp49,
					MaxTemperatureThreshold: temp51,
					DoorClosed:              true,
					Measurements:            []TempMeasurement{},
					LastNotified:            time.Now().Add(time.Minute * -3),
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
				InputParams:       controllerBoardStatusEventSuccessSlice,
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
					LastNotified: time.Now().Add(time.Minute * -3),
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
				InputParams:       controllerBoardStatusEventSuccessSlice,
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
					LastNotified: time.Now().Add(time.Minute * -3),
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
				InputParams:       emptyInputParamsSlice,
				OutputBool:        false,
				OutputInterface:   nil,
				OutputLogs:        "",
			},
			{
				TestCaseName:      "Unsuccessful due to unset config item",
				InputEdgexContext: edgexcontextBadApplicationSettings,
				InputParams:       controllerBoardStatusEventSuccessSlice,
				OutputBool:        false,
				OutputInterface:   nil,
				OutputLogs:        `Failed to load the EdgeX application settings configuration. Please make sure that the values in configuration.toml are set correctly. The error is: \"The \"AverageTemperatureMeasurementDuration\" application setting has not been set. Please set this to an acceptable time.Duration value in configuration.toml\", and the event is: {\"device\":\"ds-controller-board\",\"readings\":[{\"value\":\"{\\\"door_closed\\\":true,\\\"temperature\\\":50.0,\\\"minTemperatureStatus\\\":true,\\\"maxTemperatureStatus\\\":true}\"}]}`,
			},
			{
				TestCaseName:      "Unsuccessful due to unserializable controller board status data",
				InputEdgexContext: edgexcontextSuccess,
				InputParams:       controllerBoardStatusEventUnsuccessfulJSONSlice,
				OutputBool:        false,
				OutputInterface:   nil,
				OutputLogs:        `Failed to unmarshal controller board data, the event data is: {\"value\":\"invalid json value\"}`,
			},
			{
				TestCaseName:      "Unsuccessful due to NotificationHost not responding with HTTP Accepted",
				InputEdgexContext: edgexcontextBadNotificationHostThresholdsExceeded,
				InputParams:       controllerBoardStatusEventSuccessSlice,
				InputCheckBoardStatus: CheckBoardStatus{
					LastNotified: time.Now().Add(time.Minute * -3),
				},
				OutputBool:      true,
				OutputInterface: controllerBoardStatusEventSuccess,
				OutputLogs: fmt.Sprintf(
					`Encountered error while checking temperature thresholds: Failed to send temperature threshold exceeded notification(s) due to error: Encountered error sending the maximum temperature threshold message: Encountered error sending notification for exceeding temperature threshold: Failed to perform REST POST API call to send a notification to \"%v\", error: %v %v: %v`, edgexcontextBadNotificationHostThresholdsExceeded.Configuration.ApplicationSettings[NotificationHost], "Post", edgexcontextBadNotificationHostThresholdsExceeded.Configuration.ApplicationSettings[NotificationHost], "EOF"),
				ShouldLastNotifiedBeDifferent:             false,
				ExpectedTemperatureMeasurementSliceLength: 1,
			},
			{
				TestCaseName:                  "Unsuccessful due to MQTTEndpoint not responding with HTTP 200 OK, no temperature notification sent",
				InputEdgexContext:             edgexcontextBadMQTTEndpoint,
				InputParams:                   controllerBoardStatusEventSuccessSlice,
				InputCheckBoardStatus:         checkBoardStatusInitial,
				OutputBool:                    true,
				OutputInterface:               controllerBoardStatusEventSuccess,
				OutputLogs:                    fmt.Sprintf("Encountered error while checking the open/closed state of the door: Failed to submit the vending door state to the MQTT device service: Failed to submit REST PUT request due to error: %v %v: %v", "Put", edgexcontextBadMQTTEndpoint.Configuration.ApplicationSettings[MQTTEndpoint], "EOF"),
				ShouldLastNotifiedBeDifferent: false,
				ExpectedTemperatureMeasurementSliceLength: 1,
			},
			{
				TestCaseName:                  "Unsuccessful due to VendingEndpoint not responding with HTTP 200 OK, no temperature notification sent",
				InputEdgexContext:             edgexcontextBadVendingEndpoint,
				InputParams:                   controllerBoardStatusEventSuccessSlice,
				InputCheckBoardStatus:         checkBoardStatusInitial,
				OutputBool:                    true,
				OutputInterface:               controllerBoardStatusEventSuccess,
				OutputLogs:                    fmt.Sprintf("Encountered error while checking the open/closed state of the door: Failed to submit the controller board's status to the central vending state service: Failed to submit REST POST request due to error: %v %v: %v", "Post", edgexcontextBadVendingEndpoint.Configuration.ApplicationSettings[VendingEndpoint], "EOF"),
				ShouldLastNotifiedBeDifferent: false,
				ExpectedTemperatureMeasurementSliceLength: 1,
			},
			{
				TestCaseName:      "Unsuccessful due to VendingEndpoint not responding with HTTP 200 OK, max temperature threshold exceeded",
				InputEdgexContext: edgexcontextBadVendingEndpointMaxThresholdExceeded,
				InputParams:       controllerBoardStatusEventSuccessSlice,
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
					LastNotified: time.Now().Add(time.Minute * -3),
				},
				OutputBool:                    true,
				OutputInterface:               controllerBoardStatusEventSuccess,
				OutputLogs:                    fmt.Sprintf("Encountered error while checking temperature thresholds: Encountered error sending the controller board's status to the central vending endpoint: Failed to submit REST POST request due to error: %v %v: %v", "Post", edgexcontextBadVendingEndpoint.Configuration.ApplicationSettings[VendingEndpoint], "EOF"),
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
	logFileName := checkControllerBoardStatusLogFileName
	testTable, testServers := prepCheckControllerBoardStatusTest()
	for _, testCase := range testTable {
		ct := testCase // pinning "current test" solves concurrency issues
		t.Run(testCase.TestCaseName, func(t *testing.T) {
			// Set up the test assert/require functions
			assert := assert.New(t)
			require := require.New(t)
			// Attempt to create the log file
			file, err := os.Create(logFileName)
			require.NoError(err, "Failed to set up log file")

			// Clear the contents of the log file
			_, err = file.WriteString("")
			require.NoError(err, "Failed to clear contents of log file")

			// Pin the test's checkBoardStatus to prevent concurrency issues
			cbs := ct.InputCheckBoardStatus

			// Run the function that needs to be tested
			testBool, testInterface := cbs.CheckControllerBoardStatus(ct.InputEdgexContext, ct.InputParams...)

			// Validate the output
			assert.Equal(ct.OutputBool, testBool, "Expected output boolean to match test boolean")
			assert.Equal(ct.OutputInterface, testInterface, "Expected output interface to match test interface")
			assert.Equal(ct.ExpectedTemperatureMeasurementSliceLength, len(cbs.Measurements), "The number of temperature measurements contained in CheckBoardStatus should match expected, since the AvgTemp function should be purging old values")
			if ct.ShouldLastNotifiedBeDifferent {
				assert.NotEqual(cbs.LastNotified, ct.InputCheckBoardStatus.LastNotified, "Expected the CheckBoardStatus.LastNotified value to be different")
			}

			// Review the logs
			fileContentsAsBytes, err := ioutil.ReadFile(logFileName)
			require.NoError(err, "Failed to read from log file")
			fileContents := string(fileContentsAsBytes)

			// ignore the contents of logs if the expected string is empty
			if ct.OutputLogs != "" {
				assert.Contains(fileContents, ct.OutputLogs, "Expected logs to contain expected test case log output")
			}

			// For each test, reset the contents of the log file to nothing
			err = os.Remove(logFileName)
			require.NoError(err, "Failed to clear contents of log file before ending test")
			err = file.Close()
			assert.NoError(err, "Failed to close file descriptor for log file")

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
