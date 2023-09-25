// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"as-controller-board-status/config"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

// These constants are used to simplify code repetition when working with the
// config map/struct
const (
	AverageTemperatureMeasurementDuration = "AverageTemperatureMeasurementDuration"
	DeviceName                            = "DeviceName"
	MaxTemperatureThreshold               = "MaxTemperatureThreshold"
	MinTemperatureThreshold               = "MinTemperatureThreshold"
	//DoorStatusCommandEndpoint                 = "DoorStatusCommandEndpoint"
	NotificationCategory                      = "NotificationCategory"
	NotificationEmailAddresses                = "NotificationEmailAddresses"
	NotificationLabels                        = "NotificationLabels"
	NotificationReceiver                      = "NotificationReceiver"
	NotificationSender                        = "NotificationSender"
	NotificationSeverity                      = "NotificationSeverity"
	NotificationName                          = "NotificationName"
	NotificationSubscriptionMaxRESTRetries    = "NotificationSubscriptionMaxRESTRetries"
	NotificationSubscriptionRESTRetryInterval = "NotificationSubscriptionRESTRetryInterval"
	NotificationThrottleDuration              = "NotificationThrottleDuration"
	RESTCommandTimeout                        = "RESTCommandTimeout"
	VendingEndpoint                           = "VendingEndpoint"
)

// GetCommonSuccessConfig is used in test cases to quickly build out
// an example of a successful ControllerBoardStatusAppSettings configuration
func GetCommonSuccessConfig() *config.ControllerBoardStatusConfig {
	return &config.ControllerBoardStatusConfig{
		AverageTemperatureMeasurementDuration: "-15s",
		DeviceName:                            "controller-board",
		MaxTemperatureThreshold:               83.0,
		MinTemperatureThreshold:               10.0,
		//DoorStatusCommandEndpoint:                         "http://localhost:48082/api/v3/device/name/Inference-device/vendingDoorStatus",
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

// GetHTTPTestServer returns a basic HTTP test server that does nothing more than respond with
// a desired status code
func GetHTTPTestServer(statusCodeResponse int, response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCodeResponse)
		_, err := w.Write([]byte(response))
		if err != nil {
			panic(err)
		}

	}))
}

// GetErrorHTTPTestServer returns a basic HTTP test server that produces a guaranteed error condition
// by simply closing client connections
func GetErrorHTTPTestServer() *httptest.Server {
	var testServerThrowError *httptest.Server
	testServerThrowError = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testServerThrowError.CloseClientConnections()
	}))
	return testServerThrowError
}

type testTableRESTCommandJSONStruct struct {
	TestCaseName     string
	CheckBoardStatus CheckBoardStatus
	InputRESTMethod  string
	InputInterface   interface{}
	HTTPTestServer   *httptest.Server
	Output           error
}

func prepRESTCommandJSONTest() ([]testTableRESTCommandJSONStruct, []*httptest.Server) {
	output := []testTableRESTCommandJSONStruct{}

	// This server returns 200 OK
	testServerStatusOK := GetHTTPTestServer(http.StatusOK, "")

	// This server throws HTTP 500 as part of a non-error response
	testServer500 := GetHTTPTestServer(http.StatusInternalServerError, "test response body")

	// This server throws errors when it receives a connection
	testServerThrowError := GetErrorHTTPTestServer()

	boardStatus := CheckBoardStatus{
		Configuration: &config.ControllerBoardStatusConfig{
			RESTCommandTimeoutDuration: "15s",
		},
	}

	invalidRestMethod := "invalid rest method"

	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:     "Success GET",
			CheckBoardStatus: boardStatus,
			InputRESTMethod:  http.MethodGet,
			InputInterface:   "",
			HTTPTestServer:   testServerStatusOK,
			Output:           nil,
		})

	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:     "Success POST",
			CheckBoardStatus: boardStatus,
			InputRESTMethod:  http.MethodPost,
			InputInterface:   "simple test string",
			HTTPTestServer:   testServerStatusOK,
			Output:           nil,
		})
	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:     "Success PUT",
			CheckBoardStatus: boardStatus,
			InputRESTMethod:  http.MethodPut,
			InputInterface:   "simple test string",
			HTTPTestServer:   testServerStatusOK,
			Output:           nil,
		})
	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:     "Unsuccessful GET due to undesired status code",
			CheckBoardStatus: boardStatus,
			InputRESTMethod:  http.MethodGet,
			InputInterface:   "",
			HTTPTestServer:   testServer500,
			Output:           fmt.Errorf("did not receive an HTTP 200 status OK response from %v, instead got a response code of %v, and the response body was: %v", testServer500.URL, http.StatusInternalServerError, "test response body"),
		})
	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:     "Unsuccessful GET due to connection closure",
			CheckBoardStatus: boardStatus,
			InputRESTMethod:  http.MethodGet,
			InputInterface:   "",
			HTTPTestServer:   testServerThrowError,
			Output:           fmt.Errorf("failed to submit REST %v request due to error: %v \"%v\": %v", http.MethodGet, "Get", testServerThrowError.URL, "EOF"),
		})
	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:     "Unsuccessful GET due to unserializable JSON input",
			CheckBoardStatus: boardStatus,
			InputRESTMethod:  http.MethodGet,
			InputInterface: map[string](chan bool){
				"test": make(chan bool),
			},
			HTTPTestServer: testServerStatusOK,
			Output:         fmt.Errorf("failed to serialize the input interface as JSON: Failed to marshal into JSON string: json: unsupported type: chan bool"),
		})
	output = append(output,
		testTableRESTCommandJSONStruct{
			TestCaseName:     "Unsuccessful call due to invalid REST Method",
			CheckBoardStatus: boardStatus,
			InputRESTMethod:  invalidRestMethod,
			InputInterface:   "",
			HTTPTestServer:   testServerStatusOK,
			Output:           fmt.Errorf("failed to build the REST %v request for the URL %v due to error: net/http: invalid method \"%v\"", invalidRestMethod, testServerStatusOK.URL, invalidRestMethod), // https://github.com/golang/go/blob/7d2473dc81c659fba3f3b83bc6e93ca5fe37a898/src/net/http/request.go#L846
		})
	return output, []*httptest.Server{
		testServerStatusOK,
		testServer500,
		testServerThrowError,
	}
}

// TestRESTCommandJSON validates that the RESTCommandJSON works in all
// possible error conditions & success conditions
func TestRESTCommandJSON(t *testing.T) {
	testTable, testServers := prepRESTCommandJSONTest()
	// We are responsible for closing the test servers
	for _, testServer := range testServers {
		defer testServer.Close()
	}

	for _, testCase := range testTable {
		ct := testCase // pinning to avoid concurrency issues
		t.Run(ct.TestCaseName, func(t *testing.T) {
			assert := assert.New(t)
			err := testCase.CheckBoardStatus.RESTCommandJSON(testCase.HTTPTestServer.URL, testCase.InputRESTMethod, testCase.InputInterface)
			assert.Equal(ct.Output, err, "Expected output to be the same")
		})
	}

}
