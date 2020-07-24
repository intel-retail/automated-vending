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
	BoardStatus  CheckBoardStatus
	Output       map[string]interface{}
}

func prepBuildSubscriptionMessage() []testTableBuildSubscriptionMessageStruct {
	commonSuccessConfig := GetCommonSuccessConfig()
	return []testTableBuildSubscriptionMessageStruct{
		{
			TestCaseName: "Success",
			BoardStatus: CheckBoardStatus{
				Configuration: &commonSuccessConfig,
			},
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
			output := ct.BoardStatus.buildSubscriptionMessage()
			assert.Equal(t, ct.Output, output, "Expected output to match")
		})
	}
}

type testTablePostSubscriptionToAPIStruct struct {
	TestCaseName        string
	BoardStatus         CheckBoardStatus
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
	configSuccess := GetCommonSuccessConfig()
	configCreated := GetCommonSuccessConfig()
	configConflict := GetCommonSuccessConfig()
	configThrowError := GetCommonSuccessConfig()
	configConflict.SubscriptionHost = testServerConflict.URL
	configCreated.SubscriptionHost = testServerCreated.URL
	configThrowError.SubscriptionHost = testServerThrowError.URL

	// Assemble a configuration that doesn't want to try very many times at all
	configImpatient := GetCommonSuccessConfig()
	configImpatient.NotificationSubscriptionMaxRESTRetries = 2
	configImpatient.NotificationSubscriptionRESTRetryInterval = 1
	configImpatient.SubscriptionHost = testServer500.URL

	// Assemble a common subscription message
	commonBoardStatus := CheckBoardStatus{
		Configuration: &configSuccess,
	}
	commonSubscriptionMessage := commonBoardStatus.buildSubscriptionMessage()

	output = append(output,
		testTablePostSubscriptionToAPIStruct{
			TestCaseName: "Success Created",
			BoardStatus: CheckBoardStatus{
				Configuration: &configCreated,
			},
			SubscriptionMessage: commonSubscriptionMessage,
			Output:              nil,
			HTTPTestServer:      testServerCreated,
		},
		testTablePostSubscriptionToAPIStruct{
			TestCaseName: "Success Conflict",
			BoardStatus: CheckBoardStatus{
				Configuration: &configConflict,
			},
			SubscriptionMessage: commonSubscriptionMessage,
			Output:              nil,
			HTTPTestServer:      testServerConflict,
		},
		testTablePostSubscriptionToAPIStruct{
			TestCaseName: "Unsuccessful due to HTTP connection closed error",
			BoardStatus: CheckBoardStatus{
				Configuration: &configThrowError,
			},
			SubscriptionMessage: commonSubscriptionMessage,
			Output:              fmt.Errorf("Failed to submit REST request to subscription API endpoint: %v \"%v\": %v", "Post", testServerThrowError.URL, "EOF"),
			HTTPTestServer:      testServerThrowError,
		},
		testTablePostSubscriptionToAPIStruct{
			TestCaseName: "Unsuccessful due to unserializable input interface",
			BoardStatus: CheckBoardStatus{
				Configuration: &configSuccess,
			},
			SubscriptionMessage: map[string]interface{}{
				"test": make(chan bool),
			},
			Output:         fmt.Errorf("Failed to serialize the subscription message: %v", "json: unsupported type: chan bool"),
			HTTPTestServer: testServerStatusOK,
		},
		testTablePostSubscriptionToAPIStruct{
			TestCaseName: "Unsuccessful due to always receiving 500",
			BoardStatus: CheckBoardStatus{
				Configuration: &configImpatient,
			},
			SubscriptionMessage: commonSubscriptionMessage,
			Output:              fmt.Errorf("REST request to subscribe to the notification service failed after %v attempts. The last API response returned a %v status code, and the response body was: %v", configImpatient.NotificationSubscriptionMaxRESTRetries, 500, "test response body"),
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
			err := ct.BoardStatus.PostSubscriptionToAPI(ct.SubscriptionMessage)
			assert.Equal(t, ct.Output, err, "Expected output to match")
		})
	}
}

type testTableSubscribeToNotificationServiceStruct struct {
	TestCaseName   string
	BoardStatus    CheckBoardStatus
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

	successConfig := GetCommonSuccessConfig()
	failureConfig := GetCommonSuccessConfig()

	boardStatusSuccess := CheckBoardStatus{
		Configuration: &successConfig,
	}
	boardStatusFailure := CheckBoardStatus{
		Configuration: &failureConfig,
	}

	successConfig.SubscriptionHost = testServerCreated.URL
	failureConfig.SubscriptionHost = testServerThrowError.URL

	output = append(output,
		testTableSubscribeToNotificationServiceStruct{
			TestCaseName:   "Success",
			BoardStatus:    boardStatusSuccess,
			Output:         nil,
			HTTPTestServer: testServerCreated,
		},
		testTableSubscribeToNotificationServiceStruct{
			TestCaseName:   "Failure",
			BoardStatus:    boardStatusFailure,
			Output:         fmt.Errorf("Failed to subscribe to the EdgeX notification service due to an error thrown while performing the HTTP POST subscription to the notification service: Failed to submit REST request to subscription API endpoint: %v \"%v\": %v", "Post", testServerThrowError.URL, "EOF"),
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
			err := testCase.BoardStatus.SubscribeToNotificationService()
			assert.Equal(t, ct.Output, err, "Expected error to match output")
		})
	}
}

// TestSendNotification validates that the edge cases that aren't handled
// elsewhere are covered
func TestSendNotification(t *testing.T) {
	configSuccess := GetCommonSuccessConfig()
	boardStatus := CheckBoardStatus{
		Configuration: &configSuccess,
	}

	err := boardStatus.SendNotification(make(chan bool))

	assert.EqualError(t, err, "Failed to marshal the notification message into a JSON byte array: json: unsupported type: chan bool")
}
