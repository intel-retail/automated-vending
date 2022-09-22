// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	client_mocks "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces/mocks"
	edgex_errors "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

	mockSubscriptionClient := &client_mocks.SubscriptionClient{}
	mockSubscriptionClient.On("Add", mock.Anything, mock.Anything).Return(nil, nil)

	mockSubscriptionClientFailed := &client_mocks.SubscriptionClient{}
	mockSubscriptionClientFailed.On("Add", mock.Anything, mock.Anything).Return(nil, edgex_errors.NewCommonEdgeXWrapper(errors.New("test failed")))

	boardStatusSuccess := CheckBoardStatus{
		Configuration:      &successConfig,
		SubscriptionClient: mockSubscriptionClient,
	}
	boardStatusFailure := CheckBoardStatus{
		Configuration:      &failureConfig,
		SubscriptionClient: mockSubscriptionClientFailed,
	}

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
			Output:         fmt.Errorf("failed to subscribe to the EdgeX notification service: test failed"),
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
	mockNotificationClient := &client_mocks.NotificationClient{}
	mockNotificationClient.On("SendNotification", mock.Anything, mock.Anything).Return(nil, edgex_errors.NewCommonEdgeXWrapper(errors.New("test failed")))

	configSuccess := GetCommonSuccessConfig()
	boardStatus := CheckBoardStatus{
		Configuration:      &configSuccess,
		NotificationClient: mockNotificationClient,
	}

	err := boardStatus.SendNotification("test notification")

	assert.EqualError(t, err, "failed to send the notification: test failed")
}
