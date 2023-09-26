// Copyright © 2022-2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"as-vending/config"
	"as-vending/functions"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestControllerAddAllRoutes(t *testing.T) {

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
		c := NewController(logger.NewMockClient(), nil, &vendingState)
		// set the vendingState's MaintenanceMode boolean accordingly
		c.vendingState.MaintenanceMode = true

		req := httptest.NewRequest(http.MethodGet, "/maintenanceMode", nil)
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
		c := NewController(logger.NewMockClient(), nil, &vendingState)
		// set the vendingState's MaintenanceMode boolean accordingly
		c.vendingState.MaintenanceMode = false

		req := httptest.NewRequest(http.MethodGet, "/maintenanceMode", nil)
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
}

func TestResetDoorLock(t *testing.T) {
	stopChannel := make(chan int)
	var vendingState functions.VendingState
	vendingState.ThreadStopChannel = stopChannel
	c := NewController(logger.NewMockClient(), nil, &vendingState)
	request, _ := http.NewRequest(http.MethodPost, "", nil)
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(c.ResetDoorLock)
	handler.ServeHTTP(recorder, request)

	assert.Equal(t, false, c.vendingState.MaintenanceMode, "MaintanceMode should be false")
	assert.Equal(t, false, c.vendingState.CVWorkflowStarted, "CVWorkflowStarted should be false")
	assert.Equal(t, true, c.vendingState.DoorClosed, "DoorClosed should be false")
	assert.Equal(t, false, c.vendingState.DoorClosedDuringCVWorkflow, "DoorClosedDuringCVWorkflow should be false")
	assert.Equal(t, false, c.vendingState.DoorOpenedDuringCVWorkflow, "DoorOpenedDuringCVWorkflow should be false")
	assert.Equal(t, false, c.vendingState.InferenceDataReceived, "InferenceDataReceived should be false")
}

func TestController_BoardStatus(t *testing.T) {

	type fields struct {
		vendingState functions.VendingState
		boardStatus  functions.ControllerBoardStatus
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"Board Status Open", fields{
			vendingState: functions.VendingState{
				DoorClosed:        false,
				CVWorkflowStarted: true,
				Configuration:     new(config.VendingConfig),
			},
			boardStatus: functions.ControllerBoardStatus{
				MaxTemperatureStatus: true,
				MinTemperatureStatus: true,
				DoorClosed:           true,
			},
		},
		},
		{"Board Status Closed", fields{
			vendingState: functions.VendingState{
				DoorClosed:        true,
				CVWorkflowStarted: true,
				Configuration:     new(config.VendingConfig),
			},
			boardStatus: functions.ControllerBoardStatus{
				MaxTemperatureStatus: true,
				MinTemperatureStatus: true,
				DoorClosed:           false,
			},
		},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doorOpenStopChannel := make(chan int)
			doorCloseStopChannel := make(chan int)
			tt.fields.vendingState.DoorOpenWaitThreadStopChannel = doorOpenStopChannel
			tt.fields.vendingState.DoorCloseWaitThreadStopChannel = doorCloseStopChannel
			b, _ := json.Marshal(tt.fields.boardStatus)
			c := NewController(logger.NewMockClient(), nil, &tt.fields.vendingState)
			request, _ := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(b))
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(c.BoardStatus)
			handler.ServeHTTP(recorder, request)
			if tt.fields.vendingState.DoorClosedDuringCVWorkflow {
				close(doorOpenStopChannel)
			}
			if tt.fields.vendingState.DoorOpenedDuringCVWorkflow {
				close(doorCloseStopChannel)
			}
		})
	}
}
