// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"as-vending/functions"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestController_AddAllRoutes(t *testing.T) {

	tests := []struct {
		name         string
		badservice   bool
		want         int
		failAddRoute bool
	}{
		{
			name:         "valid case",
			failAddRoute: false,
		},
		{
			name:         "invalid case",
			failAddRoute: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAppService := &mocks.ApplicationService{}
			if !tt.failAddRoute {
				mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			} else {
				mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(fmt.Errorf("fail"))
			}

			c := &Controller{
				lc:      logger.NewMockClient(),
				service: mockAppService,
			}

			err := c.AddAllRoutes()

			if tt.failAddRoute {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

// TestGetMaintenanceMode tests the HTTP GET endpoint '/maintenanceMode' to
// verify that it reports the correct value of MaintenanceMode in its instance
// of VendingState.
func TestGetMaintenanceMode(t *testing.T) {
	maintModeTrue := functions.MaintenanceMode{
		MaintenanceMode: true,
	}
	maintModeFalse := functions.MaintenanceMode{
		MaintenanceMode: false,
	}

	t.Run("TestGetMaintenanceMode MaintenanceMode=True", func(t *testing.T) {
		var vendingState functions.VendingState
		var maintModeAPIResponse functions.MaintenanceMode
		c := NewController(logger.NewMockClient(), nil, vendingState)
		// set the vendingState's MaintenanceMode boolean accordingly
		c.vendingState.MaintenanceMode = true

		req := httptest.NewRequest("GET", "/maintenanceMode", nil)
		w := httptest.NewRecorder()

		// run the actual function in question
		c.GetMaintenanceMode(w, req)

		// parse the response
		resp := w.Result()
		_, err := utilities.ParseJSONHTTPResponseContent(resp.Body, &maintModeAPIResponse)
		require.NoError(t, err)

		defer resp.Body.Close()
		assert.Equal(t, maintModeAPIResponse, maintModeTrue, "Received a maintenance mode response that was different than anticipated")
	})
	t.Run("TestGetMaintenanceMode MaintenanceMode=False", func(t *testing.T) {
		var vendingState functions.VendingState
		var maintModeAPIResponse functions.MaintenanceMode
		c := NewController(logger.NewMockClient(), nil, vendingState)
		// set the vendingState's MaintenanceMode boolean accordingly
		c.vendingState.MaintenanceMode = false

		req := httptest.NewRequest("GET", "/maintenanceMode", nil)
		w := httptest.NewRecorder()

		// run the actual function in question
		c.GetMaintenanceMode(w, req)

		// parse the response
		resp := w.Result()
		_, err := utilities.ParseJSONHTTPResponseContent(resp.Body, &maintModeAPIResponse)
		require.NoError(t, err)

		defer resp.Body.Close()
		assert.Equal(t, maintModeAPIResponse, maintModeFalse, "Received a maintenance mode response that was different than anticipated")
	})
	t.Run("TestGetMaintenanceMode OPTIONS", func(t *testing.T) {
		var vendingState functions.VendingState

		req := httptest.NewRequest("OPTIONS", "/maintenanceMode", nil)
		w := httptest.NewRecorder()
		c := NewController(logger.NewMockClient(), nil, vendingState)
		c.GetMaintenanceMode(w, req)

		// parse the response
		resp := w.Result()
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err, "Parsing response body threw error")
		assert.Equal(t, string(body), "", "Response body was not an empty string, expected it to be empty for a pre-flight CORS OPTIONS response")
		assert.Equal(t, resp.Status, "200 OK", "OPTIONS request did not return 200")
	})
}

func TestResetDoorLock(t *testing.T) {
	stopChannel := make(chan int)
	var vendingState functions.VendingState
	vendingState.ThreadStopChannel = stopChannel
	c := NewController(logger.NewMockClient(), nil, vendingState)
	request, _ := http.NewRequest("POST", "", nil)
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(c.ResetDoorLock)
	handler.ServeHTTP(recorder, request)

	assert.Equal(t, false, c.vendingState.MaintenanceMode, "MaintanceMode should be false")
	assert.Equal(t, false, c.vendingState.CVWorkflowStarted, "CVWorkflowStarted should be false")
	assert.Equal(t, false, c.vendingState.DoorClosed, "DoorClosed should be false")
	assert.Equal(t, false, c.vendingState.DoorClosedDuringCVWorkflow, "DoorClosedDuringCVWorkflow should be false")
	assert.Equal(t, false, c.vendingState.DoorOpenedDuringCVWorkflow, "DoorOpenedDuringCVWorkflow should be false")
	assert.Equal(t, false, c.vendingState.InferenceDataReceived, "InferenceDataReceived should be false")
}

// TODO: BoardStatus handler needs to return proper http status code for unit tests
func TestBoardStatus(t *testing.T) {
	doorOpenStopChannel := make(chan int)
	doorCloseStopChannel := make(chan int)

	vendingState := functions.VendingState{
		DoorClosed:                     false,
		CVWorkflowStarted:              true,
		DoorOpenWaitThreadStopChannel:  doorOpenStopChannel,
		DoorCloseWaitThreadStopChannel: doorCloseStopChannel,
		Configuration:                  new(functions.ServiceConfiguration),
	}
	boardStatus := functions.ControllerBoardStatus{
		MaxTemperatureStatus: true,
		MinTemperatureStatus: true,
		DoorClosed:           true,
	}

	b, _ := json.Marshal(boardStatus)
	c := NewController(logger.NewMockClient(), nil, vendingState)
	request, _ := http.NewRequest("POST", "", bytes.NewBuffer(b))
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(c.BoardStatus)
	handler.ServeHTTP(recorder, request)
}
