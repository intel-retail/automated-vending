// Copyright © 2020 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

type testTableBuildSubscriptionMessageStruct struct {
	TestCaseName string
	Config       ControllerBoardStatusAppSettings
	Output       map[string]interface{}
}

func prepBuildSubscriptionMessage() []testTableBuildSubscriptionMessageStruct {
	commonSuccessConfig := GetCommonSuccessConfig()
	return []testTableBuildSubscriptionMessageStruct{
		{
			TestCaseName: "Success",
			Config:       commonSuccessConfig,
			Output: map[string]interface{}{
				"slug":     commonSuccessConfig.NotificationSlug,
				"receiver": commonSuccessConfig.NotificationReceiver,
				"subscribedCategories": []string{
					commonSuccessConfig.NotificationCategory,
				},
				"subscribedLabels": []string{
					commonSuccessConfig.NotificationCategory,
				},
				"channels": []map[string]interface{}{
					{
						"type":          "EMAIL",
						"mailAddresses": commonSuccessConfig.NotificationEmailAddresses,
					},
				},
			},
		},
	}
}

// TestBuildSubscriptionMessage validates that the BuildSubscriptionMessage function
// returns a subscription message in the proper format
func TestBuildSubscriptionMessage(t *testing.T) {
	testTable := prepBuildSubscriptionMessage()
	for _, testCase := range testTable {
		ct := testCase // pinning solves concurrency issues
		t.Run(ct.TestCaseName, func(t *testing.T) {
			output := buildSubscriptionMessage(testCase.Config)
			assert.Equal(t, output, ct.Output, "Expected output to match")
		})
	}
}

type testTablePostSubscriptionToAPIStruct struct {
	TestCaseName        string
	Config              ControllerBoardStatusAppSettings
	SubscriptionMessage map[string]interface{}
	Output              error
	HTTPTestServer      *httptest.Server
}

func prepPostSubscriptionToAPITest() ([]testTablePostSubscriptionToAPIStruct, []*httptest.Server) {
	output := []testTablePostSubscriptionToAPIStruct{}

	// This server returns 200 OK
	testServerStatusOK := GetHTTPTestServer(http.StatusOK, "")

	// This server throws HTTP 500 as part of a non-error response
	testServer500 := GetHTTPTestServer(http.StatusInternalServerError, "test response body")

	// This server throws HTTP status conflict as part of a non-error response
	testServerConflict := GetHTTPTestServer(http.StatusConflict, "")

	// This server throws HTTP status created as part of a non-error response
	testServerCreated := GetHTTPTestServer(http.StatusConflict, "")

	// This server throws errors when it receives a connection
	testServerThrowError := GetErrorHTTPTestServer()

	// Assemble a typical set of configs
	edgexconfig := GetCommonSuccessConfig()
	edgexconfigCreatedServer := GetCommonSuccessConfig()
	edgexconfigConflictServer := GetCommonSuccessConfig()
	edgexconfigThrowErrorServer := GetCommonSuccessConfig()
	edgexconfigConflictServer.SubscriptionHost = testServerConflict.URL
	edgexconfigCreatedServer.SubscriptionHost = testServerCreated.URL
	edgexconfigThrowErrorServer.SubscriptionHost = testServerThrowError.URL

	// Assemble a configuration that doesn't want to try very many times at all
	edgexconfigImpatient := GetCommonSuccessConfig()
	edgexconfigImpatient.NotificationSubscriptionMaxRESTRetries = 2
	edgexconfigImpatient.NotificationSubscriptionRESTRetryInterval = 1
	edgexconfigImpatient.SubscriptionHost = testServer500.URL

	// Assemble a common subscription message
	commonSubscriptionMessage := buildSubscriptionMessage(edgexconfig)

	output = append(output,
		testTablePostSubscriptionToAPIStruct{
			TestCaseName:        "Success Created",
			Config:              edgexconfigCreatedServer,
			SubscriptionMessage: commonSubscriptionMessage,
			Output:              nil,
			HTTPTestServer:      testServerCreated,
		},
		testTablePostSubscriptionToAPIStruct{
			TestCaseName:        "Success Conflict",
			Config:              edgexconfigConflictServer,
			SubscriptionMessage: commonSubscriptionMessage,
			Output:              nil,
			HTTPTestServer:      testServerConflict,
		},
		testTablePostSubscriptionToAPIStruct{
			TestCaseName:        "Unsuccessful due to HTTP connection closed error",
			Config:              edgexconfigThrowErrorServer,
			SubscriptionMessage: commonSubscriptionMessage,
			Output:              fmt.Errorf("Failed to submit REST request to subscription API endpoint: %v %v: %v", "Post", testServerThrowError.URL, "EOF"),
			HTTPTestServer:      testServerThrowError,
		},
		testTablePostSubscriptionToAPIStruct{
			TestCaseName: "Unsuccessful due to unserializable input interface",
			Config:       edgexconfig,
			SubscriptionMessage: map[string]interface{}{
				"test": make(chan bool),
			},
			Output:         fmt.Errorf("Failed to serialize the subscription message: %v", "json: unsupported type: chan bool"),
			HTTPTestServer: testServerStatusOK,
		},
		testTablePostSubscriptionToAPIStruct{
			TestCaseName:        "Unsuccessful due to always receiving 500",
			Config:              edgexconfigImpatient,
			SubscriptionMessage: commonSubscriptionMessage,
			Output:              fmt.Errorf("REST request to subscribe to the notification service failed after %v attempts. The last API response returned a %v status code, and the response body was: %v", edgexconfigImpatient.NotificationSubscriptionMaxRESTRetries, 500, "test response body"),
			HTTPTestServer:      testServer500,
		},
	)

	return output, []*httptest.Server{
		testServer500,
		testServerConflict,
		testServerCreated,
		testServerStatusOK,
		testServerThrowError,
	}
}

// TestPostSubscriptionToAPI validates that the PostSubscriptionToAPI function
// properly handles all error and success scenarios
func TestPostSubscriptionToAPI(t *testing.T) {
	testTable, testServers := prepPostSubscriptionToAPITest()
	// We are responsible for closing the test servers
	for _, testServer := range testServers {
		defer testServer.Close()
	}
	for _, testCase := range testTable {
		ct := testCase // pinning solves concurrency issues
		t.Run(ct.TestCaseName, func(t *testing.T) {
			err := PostSubscriptionToAPI(ct.Config, ct.SubscriptionMessage)
			assert.Equal(t, err, ct.Output, "Expected output to match")
		})
	}
}

type testTableSubscribeToNotificationServiceStruct struct {
	TestCaseName   string
	Config         ControllerBoardStatusAppSettings
	Output         error
	HTTPTestServer *httptest.Server
}

func prepSubscribeToNotificationServiceTest() ([]testTableSubscribeToNotificationServiceStruct, []*httptest.Server) {
	output := []testTableSubscribeToNotificationServiceStruct{}

	// This server throws errors when it receives a connection
	testServerThrowError := GetErrorHTTPTestServer()

	// This server throws HTTP status created as part of a non-error response
	testServerCreated := GetHTTPTestServer(http.StatusConflict, "")

	// Assemble a typical set of configs
	edgexconfigCreatedServer := GetCommonSuccessConfig()
	edgexconfigThrowErrorServer := GetCommonSuccessConfig()

	edgexconfigCreatedServer.SubscriptionHost = testServerCreated.URL
	edgexconfigThrowErrorServer.SubscriptionHost = testServerThrowError.URL

	output = append(output,
		testTableSubscribeToNotificationServiceStruct{
			TestCaseName:   "Success",
			Config:         edgexconfigCreatedServer,
			Output:         nil,
			HTTPTestServer: testServerCreated,
		},
		testTableSubscribeToNotificationServiceStruct{
			TestCaseName:   "Failure",
			Config:         edgexconfigThrowErrorServer,
			Output:         fmt.Errorf("Failed to subscribe to the EdgeX notification service due to an error thrown while performing the HTTP POST subscription to the notification service: Failed to submit REST request to subscription API endpoint: %v %v: %v", "Post", testServerThrowError.URL, "EOF"),
			HTTPTestServer: testServerThrowError,
		},
	)

	return output, []*httptest.Server{
		testServerThrowError,
		testServerCreated,
	}
}

// TestSubscribeToNotificationService validates that the
// SubscribeToNotificationService function handles its success/failure cases
// as expected
func TestSubscribeToNotificationService(t *testing.T) {
	testTable, testServers := prepSubscribeToNotificationServiceTest()
	// We are responsible for closing the test servers
	for _, testServer := range testServers {
		defer testServer.Close()
	}

	for _, testCase := range testTable {
		ct := testCase // pinning solves concurrency issues
		t.Run(ct.TestCaseName, func(t *testing.T) {
			err := SubscribeToNotificationService(testCase.Config)
			assert.Equal(t, err, ct.Output, "Expected error to match output")
		})
	}
}

// TestSendNotification validates that the edge cases that aren't handled
// elsewhere are covered
func TestSendNotification(t *testing.T) {
	edgexconfig := GetCommonSuccessConfig()
	err := SendNotification(edgexconfig, make(chan bool))

	assert.EqualError(t, err, "Failed to marshal the notification message into a JSON byte array: json: unsupported type: chan bool")
}
